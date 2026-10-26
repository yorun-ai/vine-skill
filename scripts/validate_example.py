#!/usr/bin/env python3
"""Generate and exercise published-baseline fixtures in disposable temporary directories."""
from __future__ import annotations

import json
import os
from pathlib import Path
import re
import shutil
import socket
import subprocess
import tempfile
import time
import urllib.error
import urllib.request

ROOT = Path(__file__).resolve().parents[1]
BASELINE = json.loads((ROOT / "scripts/baseline.json").read_text())
ENV = {**os.environ, "GOWORK": "off", "GOTOOLCHAIN": "local"}
# Isolate fixtures from user runtime settings so services and assets stay local.
ENV = {key: value for key, value in ENV.items() if not key.startswith("VINE_")}


def run(*args: str, cwd: Path, capture: bool = False) -> str:
    print("+", " ".join(args), flush=True)
    result = subprocess.run(args, cwd=cwd, env=ENV, check=True, text=True,
                            stdout=subprocess.PIPE if capture else None)
    return result.stdout or ""


def block(language: str, title: str) -> str:
    document = (ROOT / "references/frontend-vrpc.md").read_text()
    match = re.search(rf'```{language} title="{re.escape(title)}"\n(.*?)\n```', document, re.S)
    if not match:
        raise RuntimeError(f"missing executable reference {title}")
    return match[1] + "\n"


def free_ports() -> tuple[int, int]:
    with socket.socket() as business, socket.socket() as admin:
        business.bind(("127.0.0.1", 0))
        admin.bind(("127.0.0.1", 0))
        return business.getsockname()[1], admin.getsockname()[1]


def smoke_process(binary: Path, cwd: Path, admin_port: int, command: list[str]) -> None:
    with tempfile.TemporaryFile(mode="w+") as log:
        process = subprocess.Popen([str(binary)], cwd=cwd, env=ENV, stdout=log, stderr=log)
        try:
            deadline = time.monotonic() + 20
            while True:
                if process.poll() is not None:
                    raise RuntimeError(f"fixture exited with {process.returncode}")
                try:
                    with urllib.request.urlopen(f"http://127.0.0.1:{admin_port}/", timeout=1) as response:
                        if response.status == 200 and b"<html" in response.read().lower():
                            break
                except (urllib.error.URLError, TimeoutError):
                    pass
                if time.monotonic() > deadline:
                    raise RuntimeError("Dashboard did not become ready")
                time.sleep(0.1)
            run(*command, cwd=cwd)
        except BaseException:
            log.seek(0)
            print(log.read()[-12000:])
            raise
        finally:
            process.terminate()
            try:
                process.wait(timeout=15)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait()
                raise RuntimeError("fixture did not stop within 15 seconds")


def validate_server(work: Path) -> None:
    fixture = work / "greeting"
    shutil.copytree(ROOT / "references/example-greeting", fixture)
    run("skelc", "check", "--skel-in", "./skel", cwd=fixture)
    run("skelc", "gen", "go-module", "--skel-in", "./skel", "--go-out", "./skeled/golang",
        "--go-module", "example.com/greeting/skeled/golang", "--go-vine-version", BASELINE["vine"], cwd=fixture)
    generated = fixture / "skeled/golang"
    run("go", "mod", "tidy", cwd=generated)
    server = fixture / "src/server"
    run("go", "mod", "init", "example.com/greeting/src/server", cwd=server)
    run("go", "mod", "edit", "-go=1.27.0",
        "-require=example.com/greeting/skeled/golang@v0.0.0",
        "-replace=example.com/greeting/skeled/golang=../../skeled/golang", cwd=server)
    run("go", "get", "go.yorun.ai/vine@" + BASELINE["vine"], cwd=server)
    run("go", "mod", "tidy", cwd=server)
    selected = run("go", "list", "-m", "-f", "{{.Path}} {{.Version}}",
                   "go.yorun.ai/vine", "go.yorun.ai/skel", "go.yorun.ai/vrpc", cwd=server, capture=True)
    expected = {"go.yorun.ai/vine": BASELINE["vine"], "go.yorun.ai/skel": BASELINE["skelc"],
                "go.yorun.ai/vrpc": BASELINE["vrpcGo"]}
    if dict(line.split() for line in selected.splitlines()) != expected:
        raise RuntimeError(f"unexpected selected dependency graph:\n{selected}")
    print(selected, flush=True)
    run("go", "test", "./...", cwd=server)
    run("go", "vet", "./...", cwd=server)

    run("skelc", "gen", "go-module", "--api", "--actor", "greeting.GreetingActor",
        "--skel-in", "./skel", "--go-out", "./skeled/go-api",
        "--go-module", "example.com/greeting/skeled/go-api",
        "--go-vrpc-version", BASELINE["vrpcGo"], cwd=fixture)
    run("go", "mod", "tidy", cwd=fixture / "skeled/go-api")
    run("go", "test", "./...", cwd=fixture / "skeled/go-api")
    run("skelc", "gen", "ts", "--api", "--actor", "greeting.GreetingActor",
        "--skel-in", "./skel", "--ts-out", "./skeled/typescript", cwd=fixture)
    (fixture / "package.json").write_text(json.dumps({
        "private": True, "dependencies": {"@yorun-ai/vrpc": BASELINE["vrpcTs"]},
        "devDependencies": {"typescript": BASELINE["typescript"]},
    }))
    run("npm", "install", "--ignore-scripts", "--no-audit", "--no-fund", cwd=fixture)
    run("node", "node_modules/typescript/bin/tsc", "--strict", "--target", "ES2022",
        "--module", "Node16", "--moduleResolution", "Node16", "--skipLibCheck",
        "--outDir", "compiled", "skeled/typescript/index.ts", cwd=fixture)
    shutil.copyfile(ROOT / "scripts/smoke_clients.cjs", fixture / "smoke_clients.cjs")
    business_port, admin_port = free_ports()
    seed = server / "seed/hub.yaml"
    seed.write_text(seed.read_text().replace("7188", str(business_port)))
    main = server / "cmd/greeting/main.go"
    main.write_text(main.read_text().replace("127.0.0.1:7299", f"127.0.0.1:{admin_port}"))
    binary = fixture / "greeting-app"
    run("go", "build", "-o", str(binary), "./cmd/greeting", cwd=server)
    smoke_process(binary, server, admin_port, ["node", str(fixture / "smoke_clients.cjs"),
        f"http://127.0.0.1:{business_port}", f"http://127.0.0.1:{admin_port}"])


def validate_domains(work: Path) -> None:
    fixture = work / "shop"
    shutil.copytree(ROOT / "references/example-domains", fixture)
    run("bash", "generate.sh", cwd=fixture)
    server = fixture / "src/server"
    selected = run("go", "list", "-m", "-f", "{{.Path}} {{.Version}}",
                   "go.yorun.ai/vine", "go.yorun.ai/skel", "go.yorun.ai/vrpc", cwd=server, capture=True)
    expected = {"go.yorun.ai/vine": BASELINE["vine"], "go.yorun.ai/skel": BASELINE["skelc"],
                "go.yorun.ai/vrpc": BASELINE["vrpcGo"]}
    if dict(line.split() for line in selected.splitlines()) != expected:
        raise RuntimeError(f"unexpected two-domain dependency graph:\n{selected}")
    print(selected, flush=True)
    run("go", "test", "-race", "./...", cwd=server)
    run("go", "vet", "./...", cwd=server)


def validate_browser_reference(work: Path) -> None:
    fixture = work / "browser"
    (fixture / "skel").mkdir(parents=True)
    (fixture / "skel/domain.skel").write_text("domain example.console\n")
    (fixture / "skel/browser.skel").write_text(block("skel", "browser.skel"))
    (fixture / "skel/probe.skel").write_text("domain example.console\napi service ProbeApiService { for BrowserActor auth optional method ping { output string } }\n")
    run("skelc", "check", "--skel-in", "./skel", cwd=fixture)
    run("skelc", "gen", "go-module", "--skel-in", "./skel", "--go-out", "./skeled/golang",
        "--go-module", "example.com/console/skeled/golang", "--go-vine-version", BASELINE["vine"], cwd=fixture)
    run("go", "mod", "tidy", cwd=fixture / "skeled/golang")
    source = fixture / "frontend"
    source.mkdir()
    run("go", "mod", "init", "example.com/console/frontend", cwd=source)
    run("go", "mod", "edit", "-require=example.com/console/skeled/golang@v0.0.0",
        "-replace=example.com/console/skeled/golang=../skeled/golang", cwd=source)
    (source / "dev.go").write_text(block("go", "dev.go"))
    shutil.copyfile(ROOT / "scripts/frontend_proxy_test.go.txt", source / "dev_test.go")
    run("go", "mod", "tidy", cwd=source)
    run("go", "test", "./...", cwd=source)
    run("go", "vet", "./...", cwd=source)
    (source / "dev.go").unlink()
    (source / "dev_test.go").unlink()
    (source / "assets.go").write_text(block("go", "assets.go"))
    (source / "hub.yaml").write_text(block("yaml", "browser-hub.yaml"))
    (source / "dist").mkdir()
    (source / "dist/index.html").write_text("<!doctype html><html><body>Vine assets fixture</body></html>")
    shutil.copyfile(ROOT / "scripts/frontend_assets_test.go.txt", source / "assets_test.go")
    run("go", "mod", "tidy", cwd=source)
    run("go", "test", "./...", cwd=source)
    run("go", "vet", "./...", cwd=source)


def main() -> None:
    version = run("go", "env", "GOVERSION", cwd=ROOT, capture=True).strip()
    if version != "go" + BASELINE["go"]:
        raise RuntimeError(f"baseline requires Go {BASELINE['go']}; found {version}")
    compiler = json.loads(run("skelc", "version", cwd=ROOT, capture=True))
    if compiler["version"] != BASELINE["skelc"]:
        raise RuntimeError(f"baseline requires skelc {BASELINE['skelc']}; found {compiler['version']}")
    with tempfile.TemporaryDirectory(prefix="vine-skill-validation-") as temporary:
        work = Path(temporary)
        validate_server(work)
        validate_domains(work)
        validate_browser_reference(work)
    print("Published baseline: Go tests/vet, Event/Task retries, two-domain Rpc/events/race, API generation, TypeScript, Portal/Admin and Web references passed.")


if __name__ == "__main__":
    main()

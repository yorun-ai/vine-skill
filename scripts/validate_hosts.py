#!/usr/bin/env python3
"""Check project installation and resource integrity without launching agent hosts."""

from __future__ import annotations

import os
from pathlib import Path
import shutil
import subprocess
import tempfile

from validate_skill import ROOT, validate_local_markdown_links, validate_skill_frontmatter


INSTALLER = "skills@1.5.23"
PROJECT_ROOTS = {
    "codex": ".agents/skills",
    "claude-code": ".claude/skills",
    "opencode": ".agents/skills",
    "cursor": ".agents/skills",
    "github-copilot": ".agents/skills",
    "gemini-cli": ".agents/skills",
    "cline": ".agents/skills",
    "qwen-code": ".qwen/skills",
}
EXCLUDED_DIRECTORIES = {".git", "__pycache__", "__pypackages__"}


def validate_bundle(destination: Path) -> None:
    for source in ROOT.rglob("*"):
        relative = source.relative_to(ROOT)
        if EXCLUDED_DIRECTORIES.intersection(relative.parts) or not source.is_file():
            continue
        installed = destination / relative
        if not installed.is_file() or source.read_bytes() != installed.read_bytes():
            raise RuntimeError(f"Missing or changed installed resource: {installed}")
    validate_skill_frontmatter(destination)
    validate_local_markdown_links(destination)


def main() -> None:
    with tempfile.TemporaryDirectory(prefix="vine-skill-hosts-") as temporary:
        fixture = Path(temporary).resolve()
        project = fixture / "project"
        project.mkdir()
        subprocess.run(["git", "init", "--quiet", str(project)], check=True)
        environment = dict(os.environ)
        environment.update({
            "XDG_STATE_HOME": str(fixture / "installer-state"),
            "DISABLE_TELEMETRY": "1",
        })
        command = [
            "npx", "--yes", INSTALLER, "add", str(ROOT),
            "--agent", *PROJECT_ROOTS,
            "--skill", "vine-skill", "--yes", "--copy",
        ]
        result = subprocess.run(
            command, cwd=project, env=environment, text=True,
            stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=180,
        )
        if result.returncode:
            raise RuntimeError(f"Installer failed:\n{result.stdout}")
        checked = set()
        for host, directory in PROJECT_ROOTS.items():
            destination = project / directory / "vine-skill"
            if destination not in checked:
                validate_bundle(destination)
                checked.add(destination)
            print(f"PASS {host}: {directory}/vine-skill (copy via {INSTALLER})")

        dsh_destination = project / ".dsh/skills/vine-skill"
        shutil.copytree(
            ROOT, dsh_destination,
            ignore=shutil.ignore_patterns(*EXCLUDED_DIRECTORIES),
        )
        validate_bundle(dsh_destination)
        print("PASS dsh: .dsh/skills/vine-skill (native directory copy)")
        print("Package installation checks passed for nine hosts; client sessions were not run.")


if __name__ == "__main__":
    main()

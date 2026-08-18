#!/usr/bin/env python3
"""Validate the vine-skill package using only the Python standard library."""

from __future__ import annotations

import re
import sys
from pathlib import Path
from urllib.parse import unquote, urlsplit


ROOT = Path(__file__).resolve().parents[1]
REQUIRED_FILES = (
    "SKILL.md",
    "README.md",
    "README.zh-CN.md",
    "LICENSE",
    "CHANGELOG.md",
    "CONTRIBUTING.md",
    "agents/openai.yaml",
    "references/capabilities.md",
    "references/data-and-testing.md",
    "references/execution-and-boundaries.md",
    "references/foundations.md",
    "references/frontend-vrpc.md",
    "references/official-sources.md",
    "references/runtime-operations.md",
    "references/example-greeting/README.md",
    "references/example-greeting/src/server/core/auth_service_test.go",
    "scripts/start_vine_app.sh",
    "scripts/start_vine_app.bat",
)
REQUIRED_DELIVERY_TEXT = {
    "SKILL.md": (
        "Build the server first",
        "Add the frontend only after confirmation",
        "src/server/seed/hub.yaml",
        "skeled/golang/",
        "skeled/typescript/",
        "src/server/app/app.go",
        "src/server/cmd/<name>/main.go",
        "src/web/",
        "replace <module>/skeled/golang => ../../skeled/golang",
    ),
    "references/foundations.md": (
        "./skeled/golang",
        "./skeled/typescript",
        "src/server/seed/hub.yaml",
        "src/server/cmd/demo/main.go",
        "src/server/app/app.go",
        "src/server/impl/",
        "src/server/repo/",
        "skeled/golang/go.mod",
        "replace example.com/demo/skeled/golang => ../../skeled/golang",
    ),
    "references/frontend-vrpc.md": (
        "web.NewReverseProxy",
        "web.NewAssetsServer",
        'router.ANY("/*path",',
        "Retry-After",
        "pathPrefix: /api",
        "pathPrefix: /",
    ),
    "README.zh-CN.md": ("src/server/", "src/web/", "服务端完成后询问"),
    "references/example-greeting/README.md": (
        "auth_service_test.go",
        "Required delivery check",
        "Only a real Portal/RPCGW call proves",
    ),
    "scripts/start_vine_app.sh": (
        "--install",
        "--check",
        "VINE_PREPARE_PACKAGE",
        "wait_for_public_page",
        "$PROJECT_ROOT/src/web",
        "$SERVER_DIR/go.mod",
        "$PROJECT_ROOT/skeled/golang/go.mod",
        "./cmd/demo",
        "$PROJECT_ROOT/src/server/seed/hub.yaml",
    ),
    "scripts/start_vine_app.bat": (
        "--install",
        "--check",
        "VINE_PREPARE_PACKAGE",
        ":wait_for_page",
        "%VINE_PROJECT_ROOT%\\src\\web",
        "%VINE_SERVER_DIR%\\go.mod",
        "%VINE_PROJECT_ROOT%\\skeled\\golang\\go.mod",
        "./cmd/demo",
        "%VINE_PROJECT_ROOT%\\src\\server\\seed\\hub.yaml",
    ),
}
MIRRORED_MARKDOWN = (
    "SKILL.md",
    "references/capabilities.md",
    "references/data-and-testing.md",
    "references/execution-and-boundaries.md",
    "references/foundations.md",
    "references/frontend-vrpc.md",
    "references/official-sources.md",
    "references/runtime-operations.md",
    "references/example-greeting/README.md",
)
EXACT_MIRROR_FILES = (
    "scripts/start_vine_app.sh",
    "scripts/start_vine_app.bat",
)
STALE_MARKDOWN_TEXT = (
    "vine version --json",
    "frontend project should be managed as a separate application",
    "前端项目应作为独立应用另行管理",
    "real Portal/vRPC end-to-end validation is optional",
    "真实 Portal/vRPC 端到端验证可选",
)
LINK_PATTERN = re.compile(r"(?<!!)\[[^\]]*\]\(([^)]+)\)")
FENCED_CODE_PATTERN = re.compile(r"^(```|~~~).*?^\1\s*$", re.MULTILINE | re.DOTALL)
INLINE_CODE_PATTERN = re.compile(r"`[^`\n]*`")
AGENT_FIELD_PATTERN = re.compile(r'^  ([a-z_]+):\s*["\']?(.*?)["\']?\s*$', re.MULTILINE)


def fail(message: str) -> None:
    print(f"error: {message}", file=sys.stderr)
    raise SystemExit(1)


def validate_required_files() -> None:
    missing = [relative for relative in REQUIRED_FILES if not (ROOT / relative).is_file()]
    if missing:
        fail(f"missing required files: {', '.join(missing)}")


def validate_skill_frontmatter(root: Path = ROOT) -> None:
    lines = (root / "SKILL.md").read_text(encoding="utf-8").splitlines()
    if not lines or lines[0] != "---":
        fail(f"{root / 'SKILL.md'} must start with YAML frontmatter")

    try:
        closing_index = lines.index("---", 1)
    except ValueError:
        fail(f"{root / 'SKILL.md'} frontmatter is not closed")

    metadata: dict[str, str] = {}
    for line in lines[1:closing_index]:
        if not line.strip() or line.startswith((" ", "\t")):
            continue
        key, separator, value = line.partition(":")
        if not separator:
            fail(f"invalid SKILL.md frontmatter line: {line}")
        metadata[key.strip()] = value.strip()

    if set(metadata) != {"name", "description"}:
        fail("SKILL.md frontmatter must contain only name and description")
    if metadata["name"] != "vine-skill":
        fail("SKILL.md name must be vine-skill")
    if not metadata["description"]:
        fail("SKILL.md description must not be empty")


def validate_agent_metadata(root: Path = ROOT) -> None:
    text = (root / "agents/openai.yaml").read_text(encoding="utf-8")
    fields = dict(AGENT_FIELD_PATTERN.findall(text))
    required = {"display_name", "short_description", "default_prompt"}
    missing = sorted(required - fields.keys())
    if missing:
        fail(f"agents/openai.yaml is missing fields: {', '.join(missing)}")
    if "$vine-skill" not in fields["default_prompt"]:
        fail("agents/openai.yaml default_prompt must invoke $vine-skill")


def validate_local_markdown_links(root: Path = ROOT) -> None:
    broken: list[str] = []
    for markdown_file in sorted(root.rglob("*.md")):
        text = markdown_file.read_text(encoding="utf-8")
        text = FENCED_CODE_PATTERN.sub("", text)
        text = INLINE_CODE_PATTERN.sub("", text)
        for raw_target in LINK_PATTERN.findall(text):
            target = raw_target.strip().strip("<>")
            parsed = urlsplit(target)
            if parsed.scheme or parsed.netloc or target.startswith(("#", "mailto:")):
                continue

            relative_path = unquote(parsed.path)
            if not relative_path:
                continue
            resolved = (markdown_file.parent / relative_path).resolve()
            if not resolved.is_relative_to(root) or not resolved.exists():
                broken.append(f"{markdown_file.relative_to(root)} -> {target}")

    if broken:
        fail("broken local Markdown links:\n  " + "\n  ".join(broken))


def validate_browser_delivery_contract() -> None:
    missing: list[str] = []
    for relative, snippets in REQUIRED_DELIVERY_TEXT.items():
        text = (ROOT / relative).read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{relative}: {snippet}")
    if missing:
        fail("browser delivery contract is incomplete:\n  " + "\n  ".join(missing))


def validate_no_stale_markdown(root: Path = ROOT) -> None:
    stale: list[str] = []
    for markdown_file in sorted(root.rglob("*.md")):
        text = markdown_file.read_text(encoding="utf-8")
        for snippet in STALE_MARKDOWN_TEXT:
            if snippet in text:
                stale.append(f"{markdown_file.relative_to(root)}: {snippet}")
    if stale:
        fail("stale documentation contract found:\n  " + "\n  ".join(stale))


def markdown_shape(path: Path) -> tuple[tuple[int, ...], tuple[str, ...], int, int, int]:
    heading_levels: list[int] = []
    fence_languages: list[str] = []
    bullets = 0
    numbered = 0
    table_rows = 0
    in_fence = False

    for line in path.read_text(encoding="utf-8").splitlines():
        fence = re.match(r"^\s*(```|~~~)(.*)$", line)
        if fence:
            if in_fence:
                in_fence = False
            else:
                in_fence = True
                fence_languages.append(fence.group(2).strip())
            continue
        if in_fence:
            continue
        heading = re.match(r"^(#{1,6})\s+", line)
        if heading:
            heading_levels.append(len(heading.group(1)))
        if re.match(r"^\s*[-*]\s+", line):
            bullets += 1
        if re.match(r"^\s*\d+\.\s+", line):
            numbered += 1
        if line.startswith("|"):
            table_rows += 1

    return tuple(heading_levels), tuple(fence_languages), bullets, numbered, table_rows


def validate_markdown_shape(left: Path, right: Path, label: str) -> None:
    if markdown_shape(left) != markdown_shape(right):
        fail(f"bilingual Markdown structure differs: {label}")


def uncommented_yaml(path: Path) -> str:
    lines = []
    for line in path.read_text(encoding="utf-8").splitlines():
        if line.lstrip().startswith("#") or not line.strip():
            continue
        lines.append(line.rstrip())
    return "\n".join(lines)


def find_paired_chinese_root(root: Path = ROOT) -> Path | None:
    """Return the Chinese tree only for the expected local bilingual layout."""
    if root.parent.name != "english":
        return None

    candidate = root.parent.parent / "vine-skill"
    if not (candidate / "SKILL.md").is_file():
        return None
    return candidate


def validate_paired_chinese_tree() -> None:
    paired_zh_root = find_paired_chinese_root()
    if paired_zh_root is None:
        return

    validate_skill_frontmatter(paired_zh_root)
    validate_agent_metadata(paired_zh_root)
    validate_local_markdown_links(paired_zh_root)
    validate_no_stale_markdown(paired_zh_root)

    for relative in MIRRORED_MARKDOWN:
        english_file = ROOT / relative
        chinese_file = paired_zh_root / relative
        if not chinese_file.is_file():
            fail(f"paired Chinese tree is missing: {relative}")
        validate_markdown_shape(english_file, chinese_file, relative)

    for relative in EXACT_MIRROR_FILES:
        if (ROOT / relative).read_bytes() != (paired_zh_root / relative).read_bytes():
            fail(f"paired executable artifact differs: {relative}")

    english_example = ROOT / "references/example-greeting"
    chinese_example = paired_zh_root / "references/example-greeting"
    for english_file in sorted(english_example.rglob("*")):
        if not english_file.is_file() or english_file.name == "README.md":
            continue
        relative = english_file.relative_to(english_example)
        chinese_file = chinese_example / relative
        if not chinese_file.is_file():
            fail(f"paired Chinese example is missing: {relative}")
        if relative.as_posix() == "src/server/seed/test-hub.yaml":
            if uncommented_yaml(english_file) != uncommented_yaml(chinese_file):
                fail(f"paired example behavior differs: {relative}")
        elif english_file.read_bytes() != chinese_file.read_bytes():
            fail(f"paired example artifact differs: {relative}")


def main() -> None:
    validate_required_files()
    validate_skill_frontmatter()
    validate_agent_metadata()
    validate_local_markdown_links()
    validate_browser_delivery_contract()
    validate_no_stale_markdown()
    validate_markdown_shape(ROOT / "README.md", ROOT / "README.zh-CN.md", "repository README")
    validate_paired_chinese_tree()
    print("vine-skill validation passed")


if __name__ == "__main__":
    main()

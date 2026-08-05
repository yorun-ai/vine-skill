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
    "references/frontend-vrpc.md",
    "scripts/start_vine_app.sh",
    "scripts/start_vine_app.bat",
)
REQUIRED_DELIVERY_TEXT = {
    "SKILL.md": ("app.WebberEnabled", "WEBGW", "7288", "webgw endpoint is unavailable"),
    "references/frontend-vrpc.md": (
        "web.NewReverseProxy",
        "web.NewAssetsServer",
        'router.ANY("/*path",',
        "Retry-After",
        "pathPrefix: /api",
        "pathPrefix: /",
    ),
    "README.zh-CN.md": ("7288", "WEBGW", "RPCGW"),
    "scripts/start_vine_app.sh": (
        "--install",
        "--check",
        "VINE_PREPARE_PACKAGE",
        "wait_for_public_page",
    ),
    "scripts/start_vine_app.bat": (
        "--install",
        "--check",
        "VINE_PREPARE_PACKAGE",
        ":wait_for_page",
    ),
}
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


def validate_skill_frontmatter() -> None:
    lines = (ROOT / "SKILL.md").read_text(encoding="utf-8").splitlines()
    if not lines or lines[0] != "---":
        fail("SKILL.md must start with YAML frontmatter")

    try:
        closing_index = lines.index("---", 1)
    except ValueError:
        fail("SKILL.md frontmatter is not closed")

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


def validate_agent_metadata() -> None:
    text = (ROOT / "agents/openai.yaml").read_text(encoding="utf-8")
    fields = dict(AGENT_FIELD_PATTERN.findall(text))
    required = {"display_name", "short_description", "default_prompt"}
    missing = sorted(required - fields.keys())
    if missing:
        fail(f"agents/openai.yaml is missing fields: {', '.join(missing)}")
    if "$vine-skill" not in fields["default_prompt"]:
        fail("agents/openai.yaml default_prompt must invoke $vine-skill")


def validate_local_markdown_links() -> None:
    broken: list[str] = []
    for markdown_file in sorted(ROOT.rglob("*.md")):
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
            if not resolved.is_relative_to(ROOT) or not resolved.exists():
                broken.append(f"{markdown_file.relative_to(ROOT)} -> {target}")

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


def main() -> None:
    validate_required_files()
    validate_skill_frontmatter()
    validate_agent_metadata()
    validate_local_markdown_links()
    validate_browser_delivery_contract()
    print("vine-skill validation passed")


if __name__ == "__main__":
    main()

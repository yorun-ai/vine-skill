#!/usr/bin/env python3
"""Regression tests for the vine-skill package validator."""

from contextlib import redirect_stderr
import io
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest

import validate_skill


class PortableSkillContractTests(unittest.TestCase):
    def write_skill(self, root: Path, description: str, body: str = "# Workflow\n") -> None:
        root.mkdir(parents=True, exist_ok=True)
        (root / "SKILL.md").write_text(
            f"---\nname: vine-skill\ndescription: {description}\n---\n\n{body}",
            encoding="utf-8",
        )

    def test_accepts_host_neutral_core(self) -> None:
        with TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            self.write_skill(root, "Develop and validate Yorun Vine services.")

            validate_skill.validate_skill_frontmatter(root)
            validate_skill.validate_host_neutral_core(root)

    def test_rejects_description_over_portable_limit(self) -> None:
        with TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            self.write_skill(root, "x" * 1025)

            with redirect_stderr(io.StringIO()), self.assertRaises(SystemExit):
                validate_skill.validate_skill_frontmatter(root)

    def test_rejects_host_specific_core_instruction(self) -> None:
        with TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory)
            self.write_skill(
                root,
                "Develop and validate Yorun Vine services.",
                "# Workflow\n\nRead AGENTS.md before editing.\n",
            )

            with redirect_stderr(io.StringIO()), self.assertRaises(SystemExit):
                validate_skill.validate_host_neutral_core(root)


class PairedChineseRootTests(unittest.TestCase):
    def test_ignores_github_actions_checkout_container(self) -> None:
        with TemporaryDirectory() as temporary_directory:
            root = Path(temporary_directory) / "work" / "vine-skill" / "vine-skill"
            root.mkdir(parents=True)

            self.assertIsNone(validate_skill.find_paired_chinese_root(root))

    def test_detects_complete_local_bilingual_layout(self) -> None:
        with TemporaryDirectory() as temporary_directory:
            workspace = Path(temporary_directory) / "workspace"
            root = workspace / "english" / "vine-skill"
            paired_root = workspace / "vine-skill"
            root.mkdir(parents=True)
            paired_root.mkdir(parents=True)
            (paired_root / "SKILL.md").write_text("---\n", encoding="utf-8")

            self.assertEqual(
                validate_skill.find_paired_chinese_root(root),
                paired_root,
            )

    def test_ignores_incomplete_local_bilingual_layout(self) -> None:
        with TemporaryDirectory() as temporary_directory:
            workspace = Path(temporary_directory) / "workspace"
            root = workspace / "english" / "vine-skill"
            (workspace / "vine-skill").mkdir(parents=True)
            root.mkdir(parents=True)

            self.assertIsNone(validate_skill.find_paired_chinese_root(root))


if __name__ == "__main__":
    unittest.main()

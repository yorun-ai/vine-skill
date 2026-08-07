#!/usr/bin/env python3
"""Regression tests for the vine-skill package validator."""

from pathlib import Path
from tempfile import TemporaryDirectory
import unittest

import validate_skill


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

# Vine Skill Maintenance

This repository owns a portable Vine application Skill, its references, launchers,
and disposable validation fixtures. Read README.md and CONTRIBUTING.md before changes.
State guidance as direct responsibilities, actions and conditions, with explicit technical limits.

- Keep the workflow host-neutral; host discovery belongs in references/agent-hosts.md.
- Preserve target-project pins and user scope. Verify examples against released source,
  generated output and the actual dependency graph.
- Keep README.md and README.zh-CN.md aligned. Generate fixtures outside the checkout;
  keep generated applications, databases, dependencies and test output in disposable directories.
- Run `python3 scripts/validate_skill.py`, the Python regression tests,
  `bash -n scripts/start_vine_app.sh`, and `git diff --check`.
- For baseline, contract, example or API changes, run `bash scripts/validate_example.sh`.
  It regenerates and compiles the example and validates real Portal admission.
- Change CHANGELOG.md only in release preparation. Ordinary updates describe their
  effects in the review/PR description. Commit, push and release require explicit user authorization.

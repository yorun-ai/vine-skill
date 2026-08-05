# Contributing to vine-skill

Thank you for improving vine-skill. Contributions are licensed under Apache-2.0 as described in `LICENSE`.

## Repository scope

The maintained deliverable is this Skill package:

- `SKILL.md`: core workflow and routing rules
- `references/`: detailed, version-sensitive guidance
- `scripts/`: native launcher templates copied into generated projects
- `agents/openai.yaml`: Skill metadata
- `.github/workflows/ci.yml`: required package validation

Generated Vine applications used for manual validation are disposable fixtures, not the primary product. Do not commit business databases, installed dependencies, build output, credentials, or other runtime artifacts as repository documentation or release content.

## Contribution principles

- Preserve version-first behavior. Never silently install or upgrade Vine, skelc, Go, or infrastructure while inspecting a target project.
- Keep Skel-generated Go and TypeScript boundaries authoritative.
- Keep structured browser calls on generated vRPC clients; reserve Web/HTTP for binary data and static assets.
- Match validation claims to the actual runtime topology and commands executed.
- Keep the core workflow concise and route detailed material into the appropriate reference file.
- Keep `README.md` and `README.zh-CN.md` aligned.
- Keep launcher templates dependency-free and behaviorally aligned across Linux and Windows.

## Making a change

1. Identify whether the change affects core routing, a version-specific reference, launcher behavior, metadata, or translation only.
2. Update `SKILL.md` when the core workflow or routing table changes.
3. Update only the relevant file under `references/` for detailed guidance.
4. Update both `README.md` and `README.zh-CN.md` when repository-level documentation changes.
5. Record notable user-visible changes under `Unreleased` in `CHANGELOG.md`.
6. Run validation that matches the affected files.

## Updating version facts

The current research baseline is Vine `v0.12.0`, skelc `v0.11.1`, and Go `1.26.5` or later. A version update must include:

- primary-source evidence from the matching release, source revision, GoDoc, tests, or release notes;
- synchronized version claims across `SKILL.md`, relevant references, and both repository READMEs;
- review of affected CLI flags, generated code, lifecycle behavior, and runtime topology;
- updates to `references/official-sources.md` when source priority or links change;
- an entry in `CHANGELOG.md`.

Do not use the website's `next` documentation as proof for an older pinned target revision.

## Validation

At minimum, validate the files affected by the contribution:

```bash
python3 scripts/validate_skill.py
bash -n scripts/start_vine_app.sh
```

Also verify:

- `SKILL.md` has valid frontmatter and the expected Skill name;
- every relative Markdown link resolves from the file that contains it;
- English and Simplified Chinese repository documentation stays aligned;
- Linux and Windows launcher environment variables and lifecycle behavior remain aligned;
- realistic prompts route to the correct reference material;
- no generated fixture, database, dependency directory, build output, token, or private key is included.

When launcher behavior changes, test `--check`, first-run installation, readiness, occupied-port refusal, and cleanup in a disposable Vine browser-app fixture. Review Windows Batch changes on Windows or in Windows CI when available; otherwise report that limitation explicitly.

## Pull requests

- Keep each pull request focused on one coherent Skill change.
- Explain the behavior being changed and why it belongs in the core workflow or a specific reference.
- List the exact validation commands run and their results.
- Identify version, platform, topology, or network behavior that was not verified.
- Avoid unrelated formatting or generated-file churn.

Unless explicitly stated otherwise, submitted contributions are licensed under Apache-2.0.

# Agent Host Compatibility

Read this reference only when installing, discovering, invoking, or troubleshooting `vine-skill` in an agent host. The Vine workflow itself remains host-neutral in `SKILL.md`.

## Preserve the Portable Contract

- Keep exactly `name` and `description` in the `SKILL.md` frontmatter. They are the shared discovery contract across the supported hosts.
- Keep host-specific paths, invocation syntax, permissions, and UI metadata out of the core workflow.
- Resolve `references/` and `scripts/` relative to the installed skill directory; never assume the repository remains the current working directory.
- Use only tools exposed by the active host and obey its permission model. If shell, network, editing, or approval capabilities are unavailable, report the missing validation instead of fabricating evidence.
- Keep `agents/openai.yaml` as the Codex and ChatGPT presentation adapter. The Agent Skills standard makes it optional, but this package ships and validates it for Codex; Claude Code and OpenCode consume the same `SKILL.md` without requiring that metadata.

## Supported Hosts

These contracts were checked against the official documentation on 2026-08-20.

| Host | Project discovery | User discovery | Explicit use | Host-specific notes |
| --- | --- | --- | --- | --- |
| Codex | `.agents/skills/vine-skill/SKILL.md` | `$HOME/.agents/skills/vine-skill/SKILL.md` | Mention `$vine-skill`; `/skills` lists skills | `agents/openai.yaml` optionally controls presentation and invocation policy |
| Claude Code | `.claude/skills/vine-skill/SKILL.md` | `$HOME/.claude/skills/vine-skill/SKILL.md` | Run `/vine-skill` | The common package does not use Claude-only frontmatter or dynamic context injection |
| OpenCode | `.opencode/skills/vine-skill/SKILL.md`, `.claude/skills/...`, or `.agents/skills/...` | `$HOME/.config/opencode/skills/...`, `$HOME/.claude/skills/...`, or `$HOME/.agents/skills/...` | Ask the agent to use `vine-skill`; it loads the skill through its `skill` tool | Ensure the `skill` permission is not `deny`; `ask` requires approval |

Official references:

- [Codex: Build skills](https://learn.chatgpt.com/docs/build-skills)
- [Claude Code: Extend Claude with skills](https://code.claude.com/docs/en/skills)
- [OpenCode: Skills](https://opencode.ai/docs/skills/)

## Install for All Supported Hosts

The verified installer command pins `skills` `1.5.23`, whose declared Node.js requirement is `>=22.20.0`:

```bash
npx --yes skills@1.5.23 add yorun-ai/vine-skill --global --agent codex claude-code opencode --skill vine-skill --yes
```

Remove `--global` for a project-scoped installation. Replace the three values after `--agent` with one host name to install for only that host. Prefer the installer-managed links or copies; do not maintain divergent host-specific copies of `SKILL.md`.

## Run Compatibility Smoke Tests

For every supported host and recorded host version:

1. Confirm that `vine-skill` appears in the host's skill list or available-skill context.
2. Invoke it explicitly, then use a matching natural-language request to test implicit discovery where the host supports it.
3. Ask for a read-only Vine version and topology diagnosis; confirm that no files change.
4. Ask it to locate the generation gate and load one linked reference; confirm that relative resources resolve from the installed skill.
5. In a disposable Vine fixture, request one implementation-only change and confirm that the agent preserves generated files and runs impact-matched validation.
6. Deny or remove one required capability and confirm that the agent reports the missing evidence instead of claiming success.

Record the host version, installer version, installation scope, prompt, raw result, and any unverified capability. Passing static package validation alone proves format compatibility, not end-to-end host behavior.

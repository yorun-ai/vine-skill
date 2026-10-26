# Agent Host Compatibility

Read this reference only when installing, discovering, invoking, or troubleshooting `vine-skill` in an agent host. The Vine workflow itself remains host-neutral in `SKILL.md`.

## Preserve the Portable Contract

- Keep exactly `name` and `description` in the `SKILL.md` frontmatter. They are the shared discovery contract across the supported hosts.
- Keep host-specific paths, invocation syntax and permissions in this reference, and UI metadata in the host adapter.
- Resolve `references/` and `scripts/` relative to the installed skill directory.
- Use the active host's tools and permission model. Report completed checks and any validation gaps caused by unavailable capabilities.
- Keep `agents/openai.yaml` as the Codex and ChatGPT presentation adapter. The Agent Skills standard makes it optional; this package ships and validates it for Codex. Claude Code and OpenCode read the shared `SKILL.md` directly.

## Supported Hosts

These contracts were checked against the official documentation on 2026-08-20.

| Host | Project discovery | User discovery | Explicit use | Host-specific notes |
| --- | --- | --- | --- | --- |
| Codex | `.agents/skills/vine-skill/SKILL.md` | `$HOME/.agents/skills/vine-skill/SKILL.md` | Mention `$vine-skill`; `/skills` lists skills | `agents/openai.yaml` optionally controls presentation and invocation policy |
| Claude Code | `.claude/skills/vine-skill/SKILL.md` | `$HOME/.claude/skills/vine-skill/SKILL.md` | Run `/vine-skill` | The common package uses portable frontmatter and static linked references |
| OpenCode | `.opencode/skills/vine-skill/SKILL.md`, `.claude/skills/...`, or `.agents/skills/...` | `$HOME/.config/opencode/skills/...`, `$HOME/.claude/skills/...`, or `$HOME/.agents/skills/...` | Ask the agent to use `vine-skill`; it loads the skill through its `skill` tool | Use `allow` for direct skill access or `ask` for access with approval |

Official references:

- [Codex: Build skills](https://learn.chatgpt.com/docs/build-skills)
- [Claude Code: Extend Claude with skills](https://code.claude.com/docs/en/skills)
- [OpenCode: Skills](https://opencode.ai/docs/skills/)

## Install for All Supported Hosts

The verified installer command pins `skills` `1.5.23`, whose declared Node.js requirement is `>=22.20.0`:

```bash
npx --yes skills@1.5.23 add yorun-ai/vine-skill --global --agent codex claude-code opencode --skill vine-skill --yes
```

Remove `--global` for a project-scoped installation. Replace the three values after `--agent` with one host name to install for only that host. Use installer-managed links or synchronized copies of the shared `SKILL.md`.

## Run Compatibility Smoke Tests

For every supported host and recorded host version:

1. Confirm that `vine-skill` appears in the host's skill list or available-skill context.
2. Invoke it explicitly, then use a matching natural-language request to test implicit discovery where the host supports it.
3. Ask for a read-only Vine version and topology diagnosis; confirm that working-tree contents are preserved.
4. Ask it to locate the generation gate and load one linked reference; confirm that relative resources resolve from the installed skill.
5. In a disposable Vine fixture, request one implementation-only change and confirm that the agent preserves generated files and runs impact-matched validation.
6. Deny or remove one required capability and confirm that the agent records the affected check as pending and identifies the required capability.

Record the host version, installer version, installation scope, prompt, raw result, and any unverified capability. Static package validation establishes format compatibility; the smoke tests above establish host behavior.

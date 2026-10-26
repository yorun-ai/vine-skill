# Agent Host Compatibility

Read this reference only when installing, discovering, invoking, or troubleshooting `vine-skill` in an agent host. The Vine workflow itself remains host-neutral in `SKILL.md`.

## Preserve the Portable Contract

- Keep exactly `name` and `description` in the `SKILL.md` frontmatter. They are the shared discovery contract across the supported hosts.
- Keep host-specific paths, invocation syntax and permissions in this reference, and UI metadata in the host adapter.
- Resolve `references/` and `scripts/` relative to the installed skill directory.
- Use the active host's tools and permission model. Report completed checks and any validation gaps caused by unavailable capabilities.
- Keep `agents/openai.yaml` as the Codex and ChatGPT presentation adapter. The Agent Skills standard makes it optional; this package ships and validates it for Codex. The other hosts read the shared `SKILL.md` directly.

## Supported Hosts

The original Codex, Claude Code and OpenCode contracts were checked on 2026-08-20.
The six additional hosts were checked against official documentation and source on
2026-10-27. Paths below are local discovery roots; each contains
`vine-skill/SKILL.md` together with the rest of the package.

| Host | Project discovery | User discovery | Explicit use | Host-specific notes |
| --- | --- | --- | --- | --- |
| Codex | `.agents/skills/` | `~/.agents/skills/` | Mention `$vine-skill`; `/skills` lists skills | `agents/openai.yaml` optionally controls presentation and invocation policy |
| Claude Code | `.claude/skills/` | `~/.claude/skills/` | Run `/vine-skill` | The common package uses portable frontmatter and static linked references |
| OpenCode | `.opencode/skills/`, `.claude/skills/`, or `.agents/skills/` | `~/.config/opencode/skills/`, `~/.claude/skills/`, or `~/.agents/skills/` | Ask the agent to use `vine-skill`; it loads the skill through its `skill` tool | Use `allow` for direct skill access or `ask` for access with approval |
| Cursor | `.agents/skills/` or `.cursor/skills/` | `~/.agents/skills/` or `~/.cursor/skills/` | Select `/vine-skill` in Agent chat | Local user skills need explicit synchronization or deployment for remote agents |
| GitHub Copilot (CLI / VS Code) | `.github/skills/`, `.claude/skills/`, or `.agents/skills/` | `~/.copilot/skills/` or `~/.agents/skills/` | Run `/vine-skill` | Use a client with Agent Skills enabled; deployment to cloud agents has separate discovery and tool availability |
| Gemini CLI | `.gemini/skills/` or `.agents/skills/` | `~/.gemini/skills/` or `~/.agents/skills/` | Ask Gemini to use `vine-skill`; `/skills list` shows discovery | Trust the workspace for project skills; handle activation consent through the host |
| DSH (DeepSeek Harness) | `.dsh/skills/` or `.agents/skills/` | `$DSH_HOME/skills/` (default `~/.dsh/skills/`) or `~/.agents/skills/` | Ask DSH to use `vine-skill` | Requires `@deepseek-ai/dsh-skill` and `@deepseek-ai/dsh-skill-filesystem` in the active configuration |
| Cline | `.cline/skills/`, `.clinerules/skills/`, `.claude/skills/`, or `.agents/skills/` | `~/.cline/skills/` or `~/.agents/skills/` | Run `/vine-skill` or ask Cline to use it | Enable the skill in the Skills tab; shared `.agents` roots are confirmed in `v4.1.26` source |
| Qwen Code | `.qwen/skills/` or `.agents/skills/` | `~/.qwen/skills/` or `~/.agents/skills/` | Run `/vine-skill`; `/skills` lists skills | Within a scope, `.qwen` takes precedence over `.agents` for the same name |

For DSH, project discovery starts at the nearest ancestor containing `.git`, or the
current directory when none exists. Put the bundle directly at
`<skill-root>/vine-skill/SKILL.md`. Custom roots and `includeDefaultRoots: false`
change discovery; `DSH_AGENTS_HOME` can replace the shared user root.

For a Cline version whose discovery list omits `.agents/skills`, place the complete
bundle in `.cline/skills/vine-skill/` or `~/.cline/skills/vine-skill/` instead.

Official references:

- [Codex: Build skills](https://learn.chatgpt.com/docs/build-skills)
- [Claude Code: Extend Claude with skills](https://code.claude.com/docs/en/skills)
- [OpenCode: Skills](https://opencode.ai/docs/skills/)
- [Cursor: Agent Skills](https://cursor.com/docs/skills)
- [GitHub Copilot: Agent Skills](https://docs.github.com/en/copilot/concepts/customizations/agent-skills)
- [Copilot CLI: skill invocation](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-command-reference)
- [VS Code: Agent Skills](https://code.visualstudio.com/docs/agent-customization/agent-skills)
- [Gemini CLI: Agent Skills](https://geminicli.com/docs/cli/skills/)
- [DSH: filesystem skill provider](https://github.com/deepseek-ai/deepseek-harness/blob/master/packages/skill/skill-filesystem/README.md)
- [Cline: Skills](https://docs.cline.bot/customization/skills)
- [Cline v4.1.26: discovery roots](https://github.com/cline/cline/blob/v4.1.26/apps/vscode/src/core/storage/skill-directories.ts)
- [Qwen Code: Agent Skills](https://qwenlm.github.io/qwen-code-docs/en/users/features/skills/)

## Install for All Supported Hosts

The installer command pins `skills` `1.5.23`, whose declared Node.js requirement is
`>=22.20.0`. It recognizes eight of the supported hosts:

```bash
npx --yes skills@1.5.23 add yorun-ai/vine-skill --global \
  --agent codex claude-code opencode cursor github-copilot gemini-cli cline qwen-code \
  --skill vine-skill --yes --copy
```

Remove `--global` for a project-scoped installation. Keep one identifier after
`--agent` to install for one host. The pinned installer's copy destinations are:

| Installer identifier | Project root | Default user root |
| --- | --- | --- |
| `codex` | `.agents/skills/` | `~/.codex/skills/` (or `$CODEX_HOME/skills/`) |
| `claude-code` | `.claude/skills/` | `~/.claude/skills/` |
| `opencode` | `.agents/skills/` | `~/.config/opencode/skills/` |
| `cursor` | `.agents/skills/` | `~/.cursor/skills/` |
| `github-copilot` | `.agents/skills/` | `~/.copilot/skills/` |
| `gemini-cli` | `.agents/skills/` | `~/.gemini/skills/` |
| `cline` | `.agents/skills/` | `~/.agents/skills/` |
| `qwen-code` | `.qwen/skills/` | `~/.qwen/skills/` |

`--copy` installs the complete directory at these paths, including references and
scripts. Rerun the command to update the copies. The installer also offers symlinks;
their canonical directory and host links can differ from these copy destinations.

For DSH, use a complete bundle at its native discovery root. For a new user
installation (Bash):

```bash
git clone --depth 1 https://github.com/yorun-ai/vine-skill.git \
  "${DSH_HOME:-$HOME/.dsh}/skills/vine-skill"
```

For project scope, use `.dsh/skills/vine-skill` as the clone destination, or copy the
complete package there from a checkout. Update an existing clone with
`git -C "${DSH_HOME:-$HOME/.dsh}/skills/vine-skill" pull --ff-only`.
An existing `.agents/skills/vine-skill/` installation is also discoverable by DSH
with its default filesystem roots. The pinned installer has no `dsh` identifier;
DSH installation uses these filesystem paths.

## Validation Coverage

Run `python3 scripts/validate_hosts.py` for disposable project installation checks.
It exercises all eight installer identifiers in copy mode, checks each resulting
bundle against the source, and validates linked resources. It also checks a full
directory copy at DSH's native project path. Fixtures and installer state are
isolated from the checkout and user skill directories.

| Evidence | Coverage |
| --- | --- |
| Discovery contract | Official documentation for all nine hosts; Cline shared paths additionally confirmed at `v4.1.26` |
| Project installation | `skills@1.5.23` for eight hosts; native directory copy for DSH |
| Package integrity | File contents, portable frontmatter and relative Markdown links at each distinct installation path |
| Global installation | Default paths inspected in the pinned installer and host documentation; user-profile installation remains untested |
| Client sessions | Discovery, invocation, reference loading, implementation and capability-denial smoke tests remain unrecorded for this matrix |

The installation validator runs without starting a client or a model session.
Record client versions and the results below when those environments are available.

Local installation validation passed on 2026-10-27 with macOS, Node.js `v26.12.1`
and `skills@1.5.23`: all eight installer identifiers and the native DSH copy passed.

## Run Compatibility Smoke Tests

For every supported host and recorded host version:

1. Confirm that `vine-skill` appears in the host's skill list or available-skill context.
2. Invoke it explicitly, then use a matching natural-language request to test implicit discovery where the host supports it.
3. Ask for a read-only Vine version and topology diagnosis; confirm that working-tree contents are preserved.
4. Ask it to locate the generation gate and load one linked reference; confirm that relative resources resolve from the installed skill.
5. In a disposable Vine fixture, request one implementation-only change and confirm that the agent preserves generated files and runs impact-matched validation.
6. Deny or remove one required capability and confirm that the agent records the affected check as pending and identifies the required capability.

Record the host version, installer version, installation scope, prompt, raw result, and any unverified capability. Static package validation establishes format compatibility; the smoke tests above establish host behavior.

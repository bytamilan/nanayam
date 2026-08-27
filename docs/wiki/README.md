# Wiki source

These pages are the source for the [Nanayam GitHub Wiki](https://github.com/bytamilan/nanayam/wiki). Editing them here — not in the wiki UI — keeps documentation in the same review flow as code. `.github/workflows/publish-wiki.yml` publishes on every push to `main` that touches this directory.

## Structure

| English | தமிழ் | Covers |
|---|---|---|
| `Home.md` | `Home-ta.md` | Landing page and navigation |
| `Getting-Started.md` | `Getting-Started-ta.md` | Local setup from scratch |
| `Architecture.md` | `Architecture-ta.md` | How the components fit together |
| `CLI-Reference.md` | `CLI-Reference-ta.md` | Every command and flag |
| `Cloud-Deployment.md` | `Cloud-Deployment-ta.md` | Kubernetes deployment |
| `API-Reference.md` | `API-Reference-ta.md` | REST and gRPC endpoints |
| `Testing.md` | `Testing-ta.md` | Running and writing tests |
| `Troubleshooting.md` | `Troubleshooting-ta.md` | Common failures and fixes |
| `Contributing.md` | `Contributing-ta.md` | How to contribute |

`_Sidebar.md` and `_Footer.md` are rendered by GitHub on every wiki page.

## Conventions

**Pages come in pairs.** Every English page has a Tamil counterpart with a `-ta` suffix. Change one, change the other. A Tamil page that silently drifts out of date is worse than an obviously missing one — if you cannot write the Tamil, say so in the pull request.

**Cross-link at the top.** Every page opens with a language switcher:

```markdown
**Languages:** **English** · [தமிழ்](Page-Name-ta)
**மொழிகள்:** [English](Page-Name) · **தமிழ்**
```

**Wiki links have no extension.** Link to `Getting-Started`, not `Getting-Started.md` — that is how GitHub Wiki resolves pages.

**Diagrams are Mermaid.** Use fenced ```mermaid blocks; GitHub renders them natively, so there are no image files to keep in sync with the text.

## Local preview

```bash
# Any markdown previewer works; this one renders Mermaid.
npx -y @mermaid-js/mermaid-cli --help    # optional, for diagram export
```

Or clone the published wiki directly:

```bash
git clone https://github.com/bytamilan/nanayam.wiki.git
```

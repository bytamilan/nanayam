#!/usr/bin/env python3
"""Validate the wiki source in docs/wiki/.

These files serve two surfaces with different rules, so a mistake in one is
easy to miss from the other:

  GitHub Pages   Jekyll renders a markdown file only if it has YAML front
                 matter — without it the file is copied out as raw markdown —
                 and it emits Page.md as Page.html, so internal links need the
                 .html suffix to resolve.
  GitHub Wiki    scripts/wiki-to-github-wiki.py strips the front matter and the
                 .html on the way out.

Checked here: every English page has a Tamil counterpart, both carry a language
switcher, every served page has front matter, and no internal link points at a
page that does not exist.
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

WIKI_DIR = Path(__file__).resolve().parent.parent / "docs" / "wiki"

# Rendered by GitHub as chrome on every wiki page, plus the directory's own
# README. Jekyll skips underscore-prefixed files, so these are wiki-only and
# keep the extensionless links the wiki resolves.
SIDEBAR_PAGES = {"_Sidebar", "_Footer"}
SPECIAL_PAGES = SIDEBAR_PAGES | {"README"}

INTERNAL_LINK = re.compile(r"\[[^\]]*\]\((?!https?://|#|mailto:)([^)]+)\)")
CODE_BLOCK = re.compile(r"^```.*?^```", re.MULTILINE | re.DOTALL)
FRONT_MATTER = re.compile(r"\A---\r?\n.*?\r?\n---\r?\n", re.DOTALL)


def strip_code_blocks(text: str) -> str:
    """Blank out fenced code blocks, keeping the line count stable.

    Links inside a code block are examples, not navigation: Contributing shows
    the language-switcher markup with a Page-Name placeholder, which must not
    be checked as a real link.
    """
    return CODE_BLOCK.sub(lambda match: "\n" * match.group(0).count("\n"), text)


def main() -> int:
    if not WIKI_DIR.is_dir():
        print(f"wiki directory not found: {WIKI_DIR}", file=sys.stderr)
        return 1

    pages = {path.stem: path for path in sorted(WIKI_DIR.glob("*.md"))}
    content_pages = {name for name in pages if name not in SPECIAL_PAGES}
    english = {name for name in content_pages if not name.endswith("-ta")}

    errors: list[str] = []

    # 1. Every English page needs a Tamil counterpart, and vice versa.
    for name in sorted(english):
        if f"{name}-ta" not in content_pages:
            errors.append(f"{name}.md has no Tamil counterpart ({name}-ta.md)")

    for name in sorted(content_pages - english):
        if name.removesuffix("-ta") not in english:
            errors.append(f"{name}.md has no English counterpart")

    for name in sorted(content_pages):
        text = pages[name].read_text(encoding="utf-8")

        # 2. Front matter, or GitHub Pages serves the file as raw markdown.
        if not FRONT_MATTER.match(text):
            errors.append(f"{name}.md has no Jekyll front matter; Pages would serve it unrendered")

        # 3. A language switcher near the top, so a reader can move between the
        #    two versions from any page rather than only from Home.
        head = "\n".join(text.splitlines()[:16])
        if "**Languages:**" not in head and "**மொழிகள்:**" not in head:
            errors.append(f"{name}.md is missing the language switcher near the top")

    # 4. Internal links must resolve to a page that exists, in the form the
    #    surface actually serves.
    for name in sorted(pages):
        text = strip_code_blocks(pages[name].read_text(encoding="utf-8"))
        wiki_only = name in SIDEBAR_PAGES

        for target in INTERNAL_LINK.findall(text):
            target = target.split("#", 1)[0].strip()
            if not target or target.startswith(("/", "./", "../")):
                continue

            if target.endswith(".md"):
                errors.append(f"{name}.md links to {target}: use the .html form Jekyll emits")
                continue

            if target.endswith(".html"):
                if wiki_only:
                    errors.append(
                        f"{name}.md is wiki-only chrome and must link without .html: {target}"
                    )
                elif target.removesuffix(".html") not in pages:
                    errors.append(f"{name}.md links to a page that does not exist: {target}")
                continue

            # Extensionless: correct for the sidebar/footer, stale anywhere else.
            if target in pages:
                if not wiki_only:
                    errors.append(
                        f"{name}.md links to {target} without .html; Pages would 404 on it"
                    )
            else:
                errors.append(f"{name}.md links to a page that does not exist: {target}")

    if errors:
        print("Wiki validation failed:", file=sys.stderr)
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        return 1

    print(f"Wiki OK: {len(english)} pages, each in English and Tamil")
    return 0


if __name__ == "__main__":
    sys.exit(main())

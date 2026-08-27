#!/usr/bin/env python3
"""Validate the wiki source in docs/wiki/.

Checks that every English page has a Tamil counterpart, that both carry a
language switcher, and that no internal wiki link points at a page that does
not exist. A broken link or a missing translation is invisible until a reader
hits it, so this runs in CI instead.
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

WIKI_DIR = Path(__file__).resolve().parent.parent / "docs" / "wiki"

# Pages GitHub renders as chrome on every wiki page, plus the directory's own
# README, which documents the source and is not published.
SPECIAL_PAGES = {"_Sidebar", "_Footer", "README"}

# A markdown link that is not a URL, an anchor, or a mail link.
INTERNAL_LINK = re.compile(r"\[[^\]]*\]\((?!https?://|#|mailto:)([^)]+)\)")

# A fenced code block, including its fence lines.
CODE_BLOCK = re.compile(r"^```.*?^```", re.MULTILINE | re.DOTALL)


def strip_code_blocks(text: str) -> str:
    """Blank out fenced code blocks, keeping line count stable.

    Links inside a code block are examples, not navigation: the Contributing
    page shows the language-switcher markup with a Page-Name placeholder, which
    must not be checked as a real link.
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

    # 2. Both versions must carry a language switcher so a reader can move
    #    between them from any page, not only from Home.
    for name in sorted(content_pages):
        text = pages[name].read_text(encoding="utf-8")
        head = "\n".join(text.splitlines()[:12])
        if "**Languages:**" not in head and "**மொழிகள்:**" not in head:
            errors.append(f"{name}.md is missing the language switcher near the top")

    # 3. Internal links must resolve to a page that exists.
    for name in sorted(pages):
        text = strip_code_blocks(pages[name].read_text(encoding="utf-8"))
        for target in INTERNAL_LINK.findall(text):
            target = target.split("#", 1)[0].strip()
            if not target:
                continue
            if target.endswith(".md"):
                errors.append(
                    f"{name}.md links to {target}: wiki links carry no .md extension"
                )
                continue
            if target not in pages:
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

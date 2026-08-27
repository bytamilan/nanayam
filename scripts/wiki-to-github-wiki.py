#!/usr/bin/env python3
"""Convert the Pages-flavoured wiki source into GitHub Wiki pages.

docs/wiki/ is the single source for two surfaces that want different things:

  GitHub Pages   Jekyll needs YAML front matter to render a page at all, and
                 emits Page.md as Page.html, so links must carry .html.
  GitHub Wiki    Renders front matter as visible junk, and resolves links
                 without any extension.

The source is written for Pages, because that is the surface that is actually
live. This script strips the front matter and removes the .html from internal
links on the way to the wiki, so one set of files serves both correctly.
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

FRONT_MATTER = re.compile(r"\A---\r?\n.*?\r?\n---\r?\n+", re.DOTALL)
CODE_BLOCK = re.compile(r"^```.*?^```", re.MULTILINE | re.DOTALL)


def strip_front_matter(text: str) -> str:
    return FRONT_MATTER.sub("", text, count=1)


def delink_html(text: str, pages: set[str]) -> str:
    """Rewrite [label](Page.html) to [label](Page) outside fenced code."""
    blocks: list[str] = []

    def stash(match: re.Match[str]) -> str:
        blocks.append(match.group(0))
        return f"\x00BLOCK{len(blocks) - 1}\x00"

    body = CODE_BLOCK.sub(stash, text)

    def fix(match: re.Match[str]) -> str:
        label, target = match.group(1), match.group(2)
        if target.endswith(".html") and target[: -len(".html")] in pages:
            return f"[{label}]({target[: -len('.html')]})"
        return match.group(0)

    body = re.sub(r"\[([^\]]*)\]\(([^)\s]+)\)", fix, body)

    for index, block in enumerate(blocks):
        body = body.replace(f"\x00BLOCK{index}\x00", block)
    return body


def main(argv: list[str]) -> int:
    if len(argv) != 3:
        print(f"usage: {argv[0]} <source-dir> <wiki-dir>", file=sys.stderr)
        return 2

    source, dest = Path(argv[1]), Path(argv[2])
    if not source.is_dir():
        print(f"source directory not found: {source}", file=sys.stderr)
        return 1
    if not dest.is_dir():
        print(f"wiki directory not found: {dest}", file=sys.stderr)
        return 1

    pages = {path.stem for path in source.glob("*.md")}

    # Clear the published pages first so a page deleted from the source also
    # disappears from the wiki; the wiki's own git history is preserved.
    for stale in dest.glob("*.md"):
        stale.unlink()

    written = 0
    for path in sorted(source.glob("*.md")):
        # README documents the source directory itself and is not a wiki page.
        if path.name == "README.md":
            continue

        text = delink_html(strip_front_matter(path.read_text(encoding="utf-8")), pages)
        (dest / path.name).write_text(text, encoding="utf-8")
        written += 1

    print(f"prepared {written} wiki pages in {dest}")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))

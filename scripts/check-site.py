#!/usr/bin/env python3
"""Verify a built Jekyll site: an index exists and every internal link resolves.

GitHub Pages deploys whatever is on main without telling anyone a link broke,
so this runs in CI against a local build instead.
"""
from __future__ import annotations

import re
import sys
import urllib.parse
from pathlib import Path

LINK = re.compile(r'(?:href|src)="([^"]+)"')
EXTERNAL = ("http://", "https://", "mailto:", "#", "//", "data:")


def main(argv: list[str]) -> int:
    site = Path(argv[1] if len(argv) > 1 else "docs/_site")
    baseurl = argv[2] if len(argv) > 2 else "/nanayam"

    if not site.is_dir():
        print(f"built site not found: {site}", file=sys.stderr)
        return 1

    index = site / "index.html"
    if not index.is_file():
        print(f"no index.html at the site root ({index}) — the site would 404", file=sys.stderr)
        return 1

    broken: list[str] = []
    checked = 0

    for page in sorted(site.rglob("*.html")):
        for raw in LINK.findall(page.read_text(encoding="utf-8", errors="replace")):
            if raw.startswith(EXTERNAL):
                continue
            target = urllib.parse.unquote(raw.split("#")[0].split("?")[0])
            if not target:
                continue
            checked += 1

            if target.startswith("/"):
                rel = target[len(baseurl) + 1:] if target.startswith(baseurl + "/") else target.lstrip("/")
                dest = site / rel
            else:
                dest = page.parent / target

            if not (dest.exists() or (dest / "index.html").exists()):
                broken.append(f"{page.relative_to(site)} -> {raw}")

    if broken:
        print(f"{len(broken)} broken internal link(s):", file=sys.stderr)
        for entry in broken:
            print(f"  - {entry}", file=sys.stderr)
        return 1

    pages = sum(1 for _ in site.rglob("*.html"))
    print(f"site OK: {pages} pages, {checked} internal links all resolve")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))

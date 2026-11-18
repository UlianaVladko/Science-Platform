#!/usr/bin/env python3
"""Lightweight Go sanity checker: unused imports, unused top-level funcs, brace balance.

Not a compiler — it catches the mistakes that are cheap to make when editing
without a toolchain available (this environment has no Go).
"""
import re
import sys
from pathlib import Path

GO_KEYWORDS = {
    "break", "case", "chan", "const", "continue", "default", "defer", "else",
    "fallthrough", "for", "func", "go", "goto", "if", "import", "interface",
    "map", "package", "range", "return", "select", "struct", "switch", "type", "var",
}


def strip_strings(src: str) -> str:
    out = []
    i = 0
    n = len(src)
    while i < n:
        ch = src[i]
        if ch == '"':
            i += 1
            while i < n and src[i] != '"':
                if src[i] == '\\':
                    i += 1
                i += 1
            i += 1
            out.append('""')
        elif ch == '`':
            i += 1
            while i < n and src[i] != '`':
                i += 1
            i += 1
            out.append('``')
        elif ch == "'":
            i += 1
            while i < n and src[i] != "'":
                if src[i] == '\\':
                    i += 1
                i += 1
            i += 1
            out.append("''")
        elif src.startswith("//", i):
            while i < n and src[i] != "\n":
                i += 1
        elif src.startswith("/*", i):
            end = src.find("*/", i + 2)
            i = n if end < 0 else end + 2
        else:
            out.append(ch)
            i += 1
    return "".join(out)


def check_file(path: Path) -> list[str]:
    errors = []
    raw = path.read_text(encoding="utf-8")
    src = strip_strings(raw)

    opens = src.count("(") + src.count("{") + src.count("[")
    closes = src.count(")") + src.count("}") + src.count("]")
    if opens != closes:
        errors.append(f"{path}: unbalanced brackets ({opens} open vs {closes} close)")

    # Unused imports
    header = src.split(")", 1)[0] if "import (" in src else ""
    block = None
    m = re.search(r"import\s*\((.*?)\)", src, re.S)
    if m:
        block = m.group(1)
    if block:
        body = src[m.end():]
        for line in block.splitlines():
            line = line.strip()
            if not line:
                continue
            alias = None
            mm = re.match(r'^(\w+|\.|_)\s+"([^"]+)"', line)
            if mm:
                alias = mm.group(1)
                pkg = mm.group(2).rstrip("/").split("/")[-1]
            else:
                mm = re.match(r'^"([^"]+)"', line)
                if not mm:
                    continue
                pkg = mm.group(1).rstrip("/").split("/")[-1]
            if alias == "_":
                continue
            name = alias or pkg
            if not re.search(r"\b" + re.escape(name) + r"\.", body):
                errors.append(f"{path}: import \"{pkg}\" appears unused")

    # Unused top-level functions are checked across the whole package below,
    # never per file: exported helpers are called from sibling files.

    return errors


def check_unexported_funcs(root: Path) -> list[str]:
    """Report unexported functions with no in-package reference.

    Exported identifiers can be used by another package, so only lowercase
    names are safe to flag.
    """
    problems = []
    by_pkg: dict[Path, list[Path]] = {}
    for f in root.rglob("*.go"):
        by_pkg.setdefault(f.parent, []).append(f)

    for pkg_dir, files in by_pkg.items():
        texts = {}
        for f in files:
            texts[f] = strip_strings(f.read_text(encoding="utf-8"))
        merged = "\n".join(texts.values())
        for f, src in texts.items():
            for fm in re.finditer(r"^func\s+(?:\([^)]*\)\s*)?([a-z]\w*)\s*\(", src, re.M):
                name = fm.group(1)
                decl_re = re.compile(r"func\s+(?:\([^)]*\)\s*)?" + re.escape(name) + r"\s*\(")
                decls = len(decl_re.findall(merged))
                refs = len(re.findall(r"\b" + re.escape(name) + r"\b", merged)) - decls
                if refs <= 0:
                    problems.append(f"{f}: unexported func {name} never referenced in package {pkg_dir.name}")
    return problems


def check_import_cycles(root: Path) -> list[str]:
    """Detect import cycles between local packages.

    `go build` rejects these with "import cycle not allowed"; catching them here
    keeps the Docker build from failing on a mistake that needs no compiler.
    """
    problems = []
    pkg_imports: dict[str, set[str]] = {}
    local_dirs = {p.parent.name: p.parent for p in root.rglob("*.go")}
    for f in root.rglob("*.go"):
        # Import paths live inside string literals, so parse the raw source here
        # (strip_strings blanks literals out).
        raw = f.read_text(encoding="utf-8")
        if not re.search(r"^package\s+\w+", raw, re.M):
            continue
        pkg = f.parent.as_posix()
        pkg_imports.setdefault(pkg, set())
        for m in re.finditer(r'"([\w./-]*/internal/(\w+))"', raw):
            target = local_dirs.get(m.group(2))
            if target is not None:
                pkg_imports[pkg].add(target.as_posix())

    seen: set[tuple[str, str]] = set()
    for start in sorted(pkg_imports):
        stack = [(start, [start])]
        while stack:
            node, path = stack.pop()
            for nxt in sorted(pkg_imports.get(node, ())):
                if nxt == start:
                    pair = (start, nxt)
                    cycle = " -> ".join(
                        Path(p).name for p in path + [nxt]
                    )
                    if pair not in seen:
                        seen.add(pair)
                        problems.append(f"import cycle not allowed: {cycle}")
                elif nxt in path:
                    continue
                elif len(path) < 8:
                    stack.append((nxt, path + [nxt]))
    return problems


def main(root: str) -> int:
    base = Path(root)
    problems = []
    files = sorted(base.rglob("*.go"))
    for f in files:
        problems.extend(check_file(f))
    problems.extend(check_unexported_funcs(base))
    problems.extend(check_import_cycles(base))
    if problems:
        print("\n".join(problems))
        return 1
    print(f"OK: {len(files)} Go files, no obvious issues")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1] if len(sys.argv) > 1 else "."))

"""Run from project root after installing the package, e.g. with uv run."""

import json
from pathlib import Path
from app.main import create_app

if __name__ == "__main__":
    target = Path(__file__).resolve().parents[1] / "openapi.json"
    target.write_text(
        json.dumps(create_app().openapi(), ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )
    print(target)

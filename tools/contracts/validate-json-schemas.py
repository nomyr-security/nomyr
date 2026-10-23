#!/usr/bin/env python3
import json
import sys
from pathlib import Path

from jsonschema import Draft202012Validator, FormatChecker


ROOT = Path(__file__).resolve().parents[2]
SCHEMAS = ROOT / "schemas"
FIXTURES = ROOT / "tools/contracts/fixtures"


def validate_schema(path: Path) -> dict:
    schema = json.loads(path.read_text())
    Draft202012Validator.check_schema(schema)
    return schema


def main() -> int:
    failed = False
    for schema_path in sorted(SCHEMAS.glob("*.schema.json")):
        schema = validate_schema(schema_path)
        validator = Draft202012Validator(schema, format_checker=FormatChecker())
        fixture_directory = schema_path.name.removesuffix(".schema.json")
        for fixture_path in sorted((FIXTURES / fixture_directory).glob("*.json")):
            fixture = json.loads(fixture_path.read_text())
            errors = sorted(validator.iter_errors(fixture), key=lambda error: list(error.path))
            expects_valid = not fixture_path.name.startswith("invalid-")
            if bool(errors) == (not expects_valid):
                continue
            failed = True
            if errors:
                print(f"{fixture_path.relative_to(ROOT)}: {errors[0].message}", file=sys.stderr)
            else:
                print(f"{fixture_path.relative_to(ROOT)}: invalid fixture was accepted", file=sys.stderr)
    if failed:
        return 1
    print("JSON Schemas are valid and positive/negative compatibility fixtures behave as expected.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

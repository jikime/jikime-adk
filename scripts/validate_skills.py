#!/usr/bin/env python3
"""
Skill Metadata Validator for JikiME-ADK

Validates all SKILL.md files against the frontmatter schema.
Uses only standard library (no jsonschema dependency).

Usage:
    python scripts/validate_skills.py
    python scripts/validate_skills.py --verbose
    python scripts/validate_skills.py --skill jikime-marketing-seo
"""

import os
import re
import json
import argparse
from pathlib import Path
from typing import Any, Optional
from dataclasses import dataclass


@dataclass
class ValidationError:
    """Represents a validation error."""
    skill: str
    field: str
    message: str
    severity: str = "error"  # error, warning


@dataclass
class ValidationResult:
    """Validation result for a single skill."""
    skill: str
    valid: bool
    errors: list[ValidationError]
    warnings: list[ValidationError]


def parse_yaml_value(value: str) -> Any:
    """Parse a simple YAML value."""
    value = value.strip()

    if value.lower() in ('true', 'yes'):
        return True
    if value.lower() in ('false', 'no'):
        return False
    if value.lower() in ('null', '~', ''):
        return None
    if value.startswith('~'):
        try:
            return int(value[1:])
        except ValueError:
            return value

    try:
        return int(value)
    except ValueError:
        pass

    try:
        return float(value)
    except ValueError:
        pass

    if value.startswith('[') and value.endswith(']'):
        items = value[1:-1].split(',')
        return [item.strip().strip('"\'') for item in items if item.strip()]

    if (value.startswith('"') and value.endswith('"')) or \
       (value.startswith("'") and value.endswith("'")):
        return value[1:-1]

    return value


def extract_frontmatter(content: str) -> Optional[dict[str, Any]]:
    """Extract YAML frontmatter from SKILL.md content."""
    match = re.match(r'^---\s*\n(.*?)\n---', content, re.DOTALL)
    if not match:
        return None

    yaml_content = match.group(1)
    result = {}
    current_key = None
    nested_obj = None
    nested_key = None

    for line in yaml_content.split('\n'):
        if not line.strip() or line.strip().startswith('#'):
            continue

        indent = len(line) - len(line.lstrip())

        if indent > 0 and current_key:
            stripped = line.strip()
            if ':' in stripped:
                key, _, val = stripped.partition(':')
                key = key.strip()
                val = val.strip()

                if nested_obj is None:
                    nested_obj = {}

                if val:
                    nested_obj[key] = parse_yaml_value(val)
                else:
                    nested_obj[key] = []
                    nested_key = key
            elif stripped.startswith('- '):
                item = stripped[2:].strip().strip('"\'')
                if nested_key and isinstance(nested_obj.get(nested_key), list):
                    nested_obj[nested_key].append(item)
                elif isinstance(result.get(current_key), list):
                    result[current_key].append(item)
                elif result.get(current_key) is None:
                    result[current_key] = [item]
            continue
        else:
            if nested_obj is not None and current_key:
                result[current_key] = nested_obj
                nested_obj = None
                nested_key = None

        if ':' in line:
            key, _, val = line.partition(':')
            key = key.strip()
            val = val.strip()
            current_key = key

            if val:
                result[key] = parse_yaml_value(val)
            else:
                result[key] = None

    if nested_obj is not None and current_key:
        result[current_key] = nested_obj

    return result


def validate_skill(skill_path: Path, verbose: bool = False) -> ValidationResult:
    """Validate a single skill's frontmatter."""
    skill_name = skill_path.parent.name
    errors = []
    warnings = []

    # Check SKILL.md exists
    if not skill_path.exists():
        errors.append(ValidationError(skill_name, "file", "SKILL.md not found"))
        return ValidationResult(skill_name, False, errors, warnings)

    content = skill_path.read_text(encoding='utf-8')
    frontmatter = extract_frontmatter(content)

    if not frontmatter:
        errors.append(ValidationError(skill_name, "frontmatter", "No valid YAML frontmatter found"))
        return ValidationResult(skill_name, False, errors, warnings)

    # Required fields
    required_fields = ['name', 'description', 'version']
    for field in required_fields:
        if field not in frontmatter:
            errors.append(ValidationError(skill_name, field, f"Required field '{field}' is missing"))

    # Name validation
    name = frontmatter.get('name', '')
    if name:
        if not re.match(r'^jikime-[a-z]+-[a-z0-9@.-]+$', name):
            errors.append(ValidationError(
                skill_name, 'name',
                f"Name '{name}' doesn't match pattern 'jikime-{{domain}}-{{name}}'"
            ))
        if name != skill_name:
            warnings.append(ValidationError(
                skill_name, 'name',
                f"Name '{name}' doesn't match folder name '{skill_name}'",
                severity="warning"
            ))

    # Description validation
    desc = frontmatter.get('description', '')
    if desc:
        if len(desc) < 10:
            errors.append(ValidationError(skill_name, 'description', "Description is too short (min 10 chars)"))
        if len(desc) > 500:
            warnings.append(ValidationError(
                skill_name, 'description',
                f"Description is very long ({len(desc)} chars, recommended max 500)",
                severity="warning"
            ))

    # Version validation
    version = frontmatter.get('version', '')
    if version and not re.match(r'^\d+\.\d+\.\d+$', str(version)):
        # Allow non-semver for framework version skills (e.g., nextjs@14)
        if '@' in skill_name or 'framework' in skill_name:
            warnings.append(ValidationError(
                skill_name, 'version',
                f"Version '{version}' is not semver (allowed for framework skills)",
                severity="warning"
            ))
        else:
            errors.append(ValidationError(skill_name, 'version', f"Version '{version}' is not valid semver"))

    # Tags validation
    tags = frontmatter.get('tags', [])
    if not tags:
        warnings.append(ValidationError(skill_name, 'tags', "No tags defined", severity="warning"))
    elif not isinstance(tags, list):
        errors.append(ValidationError(skill_name, 'tags', "Tags must be an array"))

    # Triggers validation
    triggers = frontmatter.get('triggers', {})
    if isinstance(triggers, dict):
        keywords = triggers.get('keywords', [])
        if not keywords:
            warnings.append(ValidationError(
                skill_name, 'triggers.keywords',
                "No trigger keywords defined",
                severity="warning"
            ))

        phases = triggers.get('phases', [])
        if phases:
            valid_phases = ['plan', 'run', 'sync', 'implement', 'review', 'test', 'debug']
            for phase in phases:
                if phase not in valid_phases:
                    errors.append(ValidationError(
                        skill_name, 'triggers.phases',
                        f"Invalid phase '{phase}'. Valid: {valid_phases}"
                    ))

    # Progressive disclosure validation
    pd = frontmatter.get('progressive_disclosure', {})
    if isinstance(pd, dict):
        if pd.get('enabled', True):
            l1 = pd.get('level1_tokens')
            l2 = pd.get('level2_tokens')
            if l1 is None:
                warnings.append(ValidationError(
                    skill_name, 'progressive_disclosure.level1_tokens',
                    "level1_tokens not specified",
                    severity="warning"
                ))
            if l2 is None:
                warnings.append(ValidationError(
                    skill_name, 'progressive_disclosure.level2_tokens',
                    "level2_tokens not specified",
                    severity="warning"
                ))

    # Context validation
    context = frontmatter.get('context')
    if context and context not in ['fork', 'main', 'isolated']:
        errors.append(ValidationError(
            skill_name, 'context',
            f"Invalid context '{context}'. Valid: fork, main, isolated"
        ))

    # Check for content after frontmatter
    content_after_frontmatter = re.sub(r'^---\s*\n.*?\n---', '', content, flags=re.DOTALL).strip()
    if not content_after_frontmatter:
        warnings.append(ValidationError(
            skill_name, 'content',
            "No content after frontmatter",
            severity="warning"
        ))
    elif len(content_after_frontmatter) < 100:
        warnings.append(ValidationError(
            skill_name, 'content',
            "Very little content after frontmatter",
            severity="warning"
        ))

    is_valid = len(errors) == 0
    return ValidationResult(skill_name, is_valid, errors, warnings)


def validate_all_skills(skills_dir: Path, verbose: bool = False, specific_skill: Optional[str] = None) -> tuple[int, int, int]:
    """Validate all skills in the directory."""
    total = 0
    passed = 0
    failed = 0
    all_errors = []
    all_warnings = []

    for skill_folder in sorted(skills_dir.iterdir()):
        if not skill_folder.is_dir() or skill_folder.name.startswith('.'):
            continue

        if specific_skill and skill_folder.name != specific_skill:
            continue

        total += 1
        skill_file = skill_folder / "SKILL.md"
        result = validate_skill(skill_file, verbose)

        if result.valid:
            passed += 1
            if verbose:
                print(f"✓ {result.skill}")
        else:
            failed += 1
            print(f"✗ {result.skill}")
            for err in result.errors:
                print(f"    ERROR: [{err.field}] {err.message}")
                all_errors.append(err)

        if verbose or not result.valid:
            for warn in result.warnings:
                print(f"    WARN:  [{warn.field}] {warn.message}")
                all_warnings.append(warn)

    return total, passed, failed


def main():
    parser = argparse.ArgumentParser(description='Validate JikiME-ADK skill metadata')
    parser.add_argument('--skills-dir', type=Path,
                        default=Path(__file__).parent.parent / 'templates/.claude/skills',
                        help='Skills directory path')
    parser.add_argument('--verbose', '-v', action='store_true',
                        help='Show all results including passed skills')
    parser.add_argument('--skill', type=str,
                        help='Validate specific skill only')
    args = parser.parse_args()

    print(f"Validating skills in: {args.skills_dir}")
    print()

    if not args.skills_dir.exists():
        print(f"Error: Skills directory not found: {args.skills_dir}")
        return 1

    total, passed, failed = validate_all_skills(args.skills_dir, args.verbose, args.skill)

    print()
    print("=" * 50)
    print(f"Total: {total} | Passed: {passed} | Failed: {failed}")

    if failed > 0:
        print("\nValidation FAILED")
        return 1
    else:
        print("\nValidation PASSED")
        return 0


if __name__ == '__main__':
    exit(main())

#!/usr/bin/env python3
"""
Skill Catalog Generator for JikiME-ADK

Scans all skills in templates/.claude/skills/ and generates:
1. skills-catalog.yaml - Machine-readable catalog
2. docs/skills-catalog.md - Human-readable documentation

Usage:
    python scripts/generate_skill_catalog.py
    python scripts/generate_skill_catalog.py --output-dir ./custom-output

Note: Uses only standard library (no PyYAML dependency)
"""

import os
import re
import json
import argparse
from pathlib import Path
from datetime import datetime
from collections import defaultdict
from typing import Any, Optional


def parse_yaml_value(value: str) -> Any:
    """Parse a simple YAML value (no PyYAML dependency)."""
    value = value.strip()

    # Boolean
    if value.lower() in ('true', 'yes'):
        return True
    if value.lower() in ('false', 'no'):
        return False

    # None
    if value.lower() in ('null', '~', ''):
        return None

    # Number
    if value.startswith('~'):
        # Approximate number like ~100
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

    # Array (inline)
    if value.startswith('[') and value.endswith(']'):
        items = value[1:-1].split(',')
        return [item.strip().strip('"\'') for item in items if item.strip()]

    # String (remove quotes)
    if (value.startswith('"') and value.endswith('"')) or \
       (value.startswith("'") and value.endswith("'")):
        return value[1:-1]

    return value


def extract_frontmatter(content: str) -> Optional[dict[str, Any]]:
    """Extract YAML frontmatter from SKILL.md content (simple parser)."""
    match = re.match(r'^---\s*\n(.*?)\n---', content, re.DOTALL)
    if not match:
        return None

    yaml_content = match.group(1)
    result = {}
    current_key = None
    current_indent = 0
    nested_obj = None
    nested_key = None

    for line in yaml_content.split('\n'):
        if not line.strip() or line.strip().startswith('#'):
            continue

        # Check indentation
        indent = len(line) - len(line.lstrip())

        # Handle nested objects
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
                    nested_obj[key] = {}
                    nested_key = key
            elif stripped.startswith('- '):
                # List item in nested object
                item = stripped[2:].strip().strip('"\'')
                if nested_key and isinstance(nested_obj.get(nested_key), dict):
                    if not isinstance(nested_obj[nested_key], list):
                        nested_obj[nested_key] = []
                    nested_obj[nested_key].append(item)
                elif isinstance(result.get(current_key), list):
                    result[current_key].append(item)
                elif result.get(current_key) is None:
                    result[current_key] = [item]
            continue
        else:
            # Save previous nested object
            if nested_obj is not None and current_key:
                result[current_key] = nested_obj
                nested_obj = None
                nested_key = None

        # Top-level key
        if ':' in line:
            key, _, val = line.partition(':')
            key = key.strip()
            val = val.strip()
            current_key = key
            current_indent = indent

            if val:
                result[key] = parse_yaml_value(val)
            else:
                result[key] = None

    # Save last nested object
    if nested_obj is not None and current_key:
        result[current_key] = nested_obj

    return result


def scan_skills(skills_dir: Path) -> list[dict[str, Any]]:
    """Scan all skills and extract metadata."""
    skills = []

    for skill_folder in sorted(skills_dir.iterdir()):
        if not skill_folder.is_dir() or skill_folder.name.startswith('.'):
            continue

        skill_file = skill_folder / "SKILL.md"
        if not skill_file.exists():
            print(f"  Warning: No SKILL.md in {skill_folder.name}")
            continue

        content = skill_file.read_text(encoding='utf-8')
        frontmatter = extract_frontmatter(content)

        if not frontmatter:
            print(f"  Warning: No valid frontmatter in {skill_folder.name}")
            continue

        # Extract domain from naming convention
        name = frontmatter.get('name', skill_folder.name)
        parts = name.replace('jikime-', '').split('-')
        domain = parts[0] if parts else 'unknown'

        # Handle triggers
        triggers = frontmatter.get('triggers', {})
        if not isinstance(triggers, dict):
            triggers = {}

        skill_info = {
            'name': name,
            'folder': skill_folder.name,
            'domain': domain,
            'description': frontmatter.get('description', ''),
            'version': frontmatter.get('version', '1.0.0'),
            'tags': frontmatter.get('tags', []),
            'triggers': triggers,
            'progressive_disclosure': frontmatter.get('progressive_disclosure', {}),
            'user_invocable': frontmatter.get('user-invocable', False),
            'context': frontmatter.get('context', 'fork'),
            'agent': frontmatter.get('agent', 'general-purpose'),
            'allowed_tools': frontmatter.get('allowed-tools', []),
        }

        skills.append(skill_info)

    return skills


def generate_yaml_catalog(skills: list[dict], output_path: Path) -> None:
    """Generate machine-readable YAML catalog (simple format)."""
    # Group by domain
    by_domain = defaultdict(list)
    for skill in skills:
        by_domain[skill['domain']].append(skill['name'])

    lines = [
        "# JikiME-ADK Skills Catalog",
        f"# Auto-generated on {datetime.now().strftime('%Y-%m-%d %H:%M')}",
        "",
        f"version: '1.0.0'",
        f"generated_at: '{datetime.now().isoformat()}'",
        f"total_skills: {len(skills)}",
        "",
        "# Skills by domain",
        "domains:",
    ]

    for domain in sorted(by_domain.keys()):
        lines.append(f"  {domain}:")
        lines.append(f"    count: {len(by_domain[domain])}")
        lines.append(f"    skills:")
        for skill_name in sorted(by_domain[domain]):
            lines.append(f"      - {skill_name}")
        lines.append("")

    lines.append("# All skills with metadata")
    lines.append("skills:")

    for skill in sorted(skills, key=lambda x: x['name']):
        lines.append(f"  - name: {skill['name']}")
        lines.append(f"    domain: {skill['domain']}")
        lines.append(f"    description: \"{skill['description']}\"")
        lines.append(f"    version: {skill['version']}")
        lines.append(f"    user_invocable: {str(skill['user_invocable']).lower()}")
        if skill['tags']:
            tags_str = ', '.join(f'"{t}"' for t in skill['tags'][:5])
            lines.append(f"    tags: [{tags_str}]")
        lines.append("")

    output_path.write_text("\n".join(lines), encoding='utf-8')
    print(f"Generated: {output_path}")


def generate_markdown_catalog(skills: list[dict], output_path: Path) -> None:
    """Generate human-readable Markdown catalog."""
    # Group by domain
    by_domain = defaultdict(list)
    for skill in skills:
        by_domain[skill['domain']].append(skill)

    lines = [
        "# JikiME-ADK Skills Catalog",
        "",
        f"> Auto-generated on {datetime.now().strftime('%Y-%m-%d %H:%M')}",
        f"> Total Skills: **{len(skills)}**",
        "",
        "## Overview",
        "",
        "| Domain | Count | Skills |",
        "|--------|-------|--------|",
    ]

    for domain in sorted(by_domain.keys()):
        domain_skills = by_domain[domain]
        skill_names = ", ".join(f"`{s['name']}`" for s in domain_skills[:3])
        if len(domain_skills) > 3:
            skill_names += f" +{len(domain_skills) - 3} more"
        lines.append(f"| **{domain}** | {len(domain_skills)} | {skill_names} |")

    lines.append("")
    lines.append("---")
    lines.append("")

    # Detail by domain
    for domain in sorted(by_domain.keys()):
        domain_skills = by_domain[domain]
        lines.append(f"## {domain.title()} Skills ({len(domain_skills)})")
        lines.append("")
        lines.append("| Skill | Description | Tags | Invocable |")
        lines.append("|-------|-------------|------|-----------|")

        for skill in sorted(domain_skills, key=lambda x: x['name']):
            tags = ", ".join(str(t) for t in skill['tags'][:3]) if skill['tags'] else "-"
            invocable = "Yes" if skill['user_invocable'] else "No"
            desc = skill['description']
            if len(desc) > 60:
                desc = desc[:57] + "..."
            lines.append(f"| `{skill['name']}` | {desc} | {tags} | {invocable} |")

        lines.append("")

    # Keyword index
    lines.append("---")
    lines.append("")
    lines.append("## Keyword Index")
    lines.append("")

    keyword_map = defaultdict(list)
    for skill in skills:
        triggers = skill.get('triggers', {})
        if isinstance(triggers, dict):
            keywords = triggers.get('keywords', [])
            if isinstance(keywords, list):
                for kw in keywords:
                    keyword_map[str(kw).lower()].append(skill['name'])

    lines.append("| Keyword | Skills |")
    lines.append("|---------|--------|")

    sorted_keywords = sorted(keyword_map.keys())[:50]  # Top 50 keywords
    for kw in sorted_keywords:
        skill_list = ", ".join(f"`{s}`" for s in keyword_map[kw][:3])
        if len(keyword_map[kw]) > 3:
            skill_list += f" +{len(keyword_map[kw]) - 3}"
        lines.append(f"| {kw} | {skill_list} |")

    if len(keyword_map) > 50:
        lines.append(f"| ... | +{len(keyword_map) - 50} more keywords |")

    lines.append("")
    lines.append("---")
    lines.append("")
    lines.append("*Generated by `scripts/generate_skill_catalog.py`*")

    output_path.write_text("\n".join(lines), encoding='utf-8')
    print(f"Generated: {output_path}")


def main():
    parser = argparse.ArgumentParser(description='Generate JikiME-ADK skill catalog')
    parser.add_argument('--skills-dir', type=Path,
                        default=Path(__file__).parent.parent / 'templates/.claude/skills',
                        help='Skills directory path')
    parser.add_argument('--output-dir', type=Path,
                        default=Path(__file__).parent.parent,
                        help='Output directory for generated files')
    args = parser.parse_args()

    print(f"Scanning skills in: {args.skills_dir}")

    if not args.skills_dir.exists():
        print(f"Error: Skills directory not found: {args.skills_dir}")
        return 1

    skills = scan_skills(args.skills_dir)
    print(f"Found {len(skills)} skills")

    # Generate YAML catalog
    yaml_output = args.output_dir / 'skills-catalog.yaml'
    generate_yaml_catalog(skills, yaml_output)

    # Generate Markdown catalog
    md_output = args.output_dir / 'docs' / 'skills-catalog.md'
    md_output.parent.mkdir(parents=True, exist_ok=True)
    generate_markdown_catalog(skills, md_output)

    print("\nDone!")
    return 0


if __name__ == '__main__':
    exit(main())

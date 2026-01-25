#!/usr/bin/env python3
"""
Add progressive_disclosure settings to SKILL.md files.
Based on Alfred's Progressive Disclosure pattern from MoAI-ADK.
"""

import os
import re
from pathlib import Path


def add_progressive_disclosure(skill_path: Path) -> bool:
    """Add progressive_disclosure section to SKILL.md frontmatter if not present."""

    content = skill_path.read_text(encoding='utf-8')

    # Check if already has progressive_disclosure
    if 'progressive_disclosure:' in content:
        print(f"  SKIP: {skill_path.name} (already has progressive_disclosure)")
        return False

    # Find the frontmatter section
    frontmatter_match = re.match(r'^---\n(.*?)\n---', content, re.DOTALL)
    if not frontmatter_match:
        print(f"  ERROR: {skill_path.name} (no frontmatter found)")
        return False

    frontmatter = frontmatter_match.group(1)
    rest_of_file = content[frontmatter_match.end():]

    # Calculate approximate token counts based on file size
    file_size = len(content)
    level2_tokens = min(10000, max(2000, file_size // 4))  # Estimate ~4 chars per token

    # Add progressive_disclosure after triggers section
    progressive_disclosure_section = f"""
# Progressive Disclosure Configuration
progressive_disclosure:
  enabled: true
  level1_tokens: ~100
  level2_tokens: ~{level2_tokens}"""

    # Find position after triggers section
    triggers_match = re.search(r'(triggers:\s*\n(?:  .*\n)*)', frontmatter)
    if triggers_match:
        insert_pos = triggers_match.end()
        new_frontmatter = (
            frontmatter[:insert_pos].rstrip() +
            progressive_disclosure_section + '\n' +
            frontmatter[insert_pos:].lstrip('\n')
        )
    else:
        # Add at the end of frontmatter
        new_frontmatter = frontmatter.rstrip() + progressive_disclosure_section + '\n'

    # Reconstruct the file
    new_content = f"---\n{new_frontmatter}\n---{rest_of_file}"

    skill_path.write_text(new_content, encoding='utf-8')
    print(f"  UPDATED: {skill_path.parent.name}/SKILL.md")
    return True


def main():
    # Find all SKILL.md files
    skills_dir = Path(__file__).parent.parent / "templates" / ".claude" / "skills"

    if not skills_dir.exists():
        print(f"Skills directory not found: {skills_dir}")
        return

    skill_files = list(skills_dir.glob("*/SKILL.md"))
    print(f"Found {len(skill_files)} SKILL.md files\n")

    updated_count = 0
    for skill_path in sorted(skill_files):
        if add_progressive_disclosure(skill_path):
            updated_count += 1

    print(f"\nDone! Updated {updated_count} of {len(skill_files)} files.")


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""
Skill Test Runner for JikiME-ADK

Validates skill test examples and trigger configurations.
Uses only standard library (no external dependencies).

Usage:
    python scripts/test_skills.py
    python scripts/test_skills.py --skill jikime-marketing-seo
    python scripts/test_skills.py --verbose
"""

import os
import re
import argparse
from pathlib import Path
from typing import Any, Optional
from dataclasses import dataclass


def parse_simple_yaml(content: str) -> dict[str, Any]:
    """Parse simple flat YAML format."""
    result = {}
    current_list_key = None
    current_list = []

    for line in content.split('\n'):
        line = line.rstrip()

        # Skip comments and empty lines
        if not line or line.strip().startswith('#'):
            continue

        # List item
        if line.startswith('  - '):
            item = line[4:].strip()
            current_list.append(item)
            continue

        # End of list - save it
        if current_list_key and current_list:
            result[current_list_key] = current_list
            current_list = []
            current_list_key = None

        # Key-value pair
        if ':' in line and not line.startswith(' '):
            key, _, val = line.partition(':')
            key = key.strip()
            val = val.strip()

            if val:
                # Simple value
                result[key] = val
            else:
                # Start of list
                current_list_key = key
                current_list = []

    # Save last list
    if current_list_key and current_list:
        result[current_list_key] = current_list

    return result


@dataclass
class TestResult:
    """Result of a single test."""
    name: str
    passed: bool
    message: str


@dataclass
class SkillTestResult:
    """Result of testing a skill."""
    skill: str
    has_tests: bool
    tests_run: int
    tests_passed: int
    trigger_tests_run: int
    trigger_tests_passed: int
    results: list[TestResult]


def extract_keywords_from_skill(skill_path: Path) -> list[str]:
    """Extract trigger keywords from SKILL.md frontmatter."""
    skill_file = skill_path / 'SKILL.md'
    if not skill_file.exists():
        return []

    content = skill_file.read_text(encoding='utf-8')
    match = re.match(r'^---\s*\n(.*?)\n---', content, re.DOTALL)
    if not match:
        return []

    yaml_content = match.group(1)
    keywords = []
    in_keywords = False

    for line in yaml_content.split('\n'):
        stripped = line.strip()
        if stripped.startswith('keywords:'):
            in_keywords = True
            # Check for inline array
            if '[' in stripped:
                match = re.search(r'\[(.*?)\]', stripped)
                if match:
                    items = match.group(1).split(',')
                    keywords = [i.strip().strip('"\'') for i in items if i.strip()]
                in_keywords = False
            continue

        if in_keywords:
            if stripped.startswith('- '):
                keywords.append(stripped[2:].strip().strip('"\''))
            elif stripped and not stripped.startswith('#'):
                # End of keywords section
                in_keywords = False

    return keywords


def validate_trigger(input_text: str, keywords: list[str]) -> bool:
    """Check if input would trigger any of the keywords."""
    input_lower = input_text.lower()
    for keyword in keywords:
        if keyword.lower() in input_lower:
            return True
    return False


def test_skill(skill_path: Path, verbose: bool = False) -> SkillTestResult:
    """Test a single skill."""
    skill_name = skill_path.name
    results = []

    # Check for test files
    tests_dir = skill_path / 'tests'
    examples_file = tests_dir / 'examples.yaml'

    if not examples_file.exists():
        return SkillTestResult(
            skill=skill_name,
            has_tests=False,
            tests_run=0,
            tests_passed=0,
            trigger_tests_run=0,
            trigger_tests_passed=0,
            results=[]
        )

    # Load skill keywords from SKILL.md
    skill_keywords = extract_keywords_from_skill(skill_path)

    # Load test examples
    test_content = examples_file.read_text(encoding='utf-8')
    test_data = parse_simple_yaml(test_content)

    tests_run = 0
    tests_passed = 0
    trigger_tests_run = 0
    trigger_tests_passed = 0

    # Get test keywords (from test file or skill)
    test_keywords = test_data.get('keywords', [])
    if isinstance(test_keywords, str):
        test_keywords = [test_keywords]
    keywords_to_use = test_keywords if test_keywords else skill_keywords

    # Find and validate numbered tests (test_N_name, test_N_input, test_N_expected)
    test_nums = set()
    for key in test_data.keys():
        match = re.match(r'test_(\d+)_', key)
        if match:
            test_nums.add(int(match.group(1)))

    for num in sorted(test_nums):
        tests_run += 1
        test_name = test_data.get(f'test_{num}_name', f'Test {num}')
        test_input = test_data.get(f'test_{num}_input', '')
        test_expected = test_data.get(f'test_{num}_expected', '')

        if test_input and test_expected:
            tests_passed += 1
            results.append(TestResult(test_name, True, "Test case well-formed"))

            # Check if input triggers any keyword
            if keywords_to_use and not validate_trigger(test_input, keywords_to_use):
                results.append(TestResult(
                    f"{test_name} - trigger check",
                    False,
                    f"Input doesn't trigger any keyword"
                ))
        else:
            missing = []
            if not test_input:
                missing.append('input')
            if not test_expected:
                missing.append('expected')
            results.append(TestResult(
                test_name, False,
                f"Missing: {', '.join(missing)}"
            ))

    # Trigger tests
    should_trigger = test_data.get('should_trigger', [])
    should_not_trigger = test_data.get('should_not_trigger', [])

    if isinstance(should_trigger, str):
        should_trigger = [should_trigger]
    if isinstance(should_not_trigger, str):
        should_not_trigger = [should_not_trigger]

    # Should trigger tests
    for input_text in should_trigger:
        trigger_tests_run += 1
        if validate_trigger(input_text, keywords_to_use):
            trigger_tests_passed += 1
            if verbose:
                results.append(TestResult(
                    f"Trigger: '{input_text[:25]}...'",
                    True, "Correctly triggers"
                ))
        else:
            results.append(TestResult(
                f"Trigger: '{input_text[:25]}...'",
                False, "Should trigger but doesn't"
            ))

    # Should NOT trigger tests
    for input_text in should_not_trigger:
        trigger_tests_run += 1
        if not validate_trigger(input_text, keywords_to_use):
            trigger_tests_passed += 1
            if verbose:
                results.append(TestResult(
                    f"No-trigger: '{input_text[:25]}...'",
                    True, "Correctly doesn't trigger"
                ))
        else:
            results.append(TestResult(
                f"No-trigger: '{input_text[:25]}...'",
                False, "Should not trigger but does"
            ))

    return SkillTestResult(
        skill=skill_name,
        has_tests=True,
        tests_run=tests_run,
        tests_passed=tests_passed,
        trigger_tests_run=trigger_tests_run,
        trigger_tests_passed=trigger_tests_passed,
        results=results
    )


def main():
    parser = argparse.ArgumentParser(description='Test JikiME-ADK skills')
    parser.add_argument('--skills-dir', type=Path,
                        default=Path(__file__).parent.parent / 'templates/.claude/skills',
                        help='Skills directory path')
    parser.add_argument('--skill', type=str,
                        help='Test specific skill only')
    parser.add_argument('--verbose', '-v', action='store_true',
                        help='Show detailed results')
    args = parser.parse_args()

    print(f"Testing skills in: {args.skills_dir}")
    print()

    if not args.skills_dir.exists():
        print(f"Error: Skills directory not found: {args.skills_dir}")
        return 1

    total_skills = 0
    skills_with_tests = 0
    total_tests_passed = 0
    total_tests_run = 0
    total_trigger_passed = 0
    total_trigger_run = 0
    failed_results = []

    for skill_folder in sorted(args.skills_dir.iterdir()):
        if not skill_folder.is_dir() or skill_folder.name.startswith('.') or skill_folder.name.startswith('_'):
            continue

        if args.skill and skill_folder.name != args.skill:
            continue

        total_skills += 1
        result = test_skill(skill_folder, args.verbose)

        if result.has_tests:
            skills_with_tests += 1
            total_tests_run += result.tests_run
            total_tests_passed += result.tests_passed
            total_trigger_run += result.trigger_tests_run
            total_trigger_passed += result.trigger_tests_passed

            all_passed = (result.tests_passed == result.tests_run and
                          result.trigger_tests_passed == result.trigger_tests_run)
            status = "✓" if all_passed else "✗"
            print(f"{status} {result.skill}")
            print(f"    Tests: {result.tests_passed}/{result.tests_run}")
            print(f"    Triggers: {result.trigger_tests_passed}/{result.trigger_tests_run}")

            if args.verbose or not all_passed:
                for r in result.results:
                    if not r.passed or args.verbose:
                        symbol = "✓" if r.passed else "✗"
                        print(f"      {symbol} {r.name}: {r.message}")

            if not all_passed:
                failed_results.append(result)

        elif args.verbose:
            print(f"○ {result.skill} (no tests)")

    print()
    print("=" * 50)
    print(f"Skills: {total_skills} total, {skills_with_tests} with tests")
    print(f"Tests: {total_tests_passed}/{total_tests_run} passed")
    print(f"Triggers: {total_trigger_passed}/{total_trigger_run} passed")

    if skills_with_tests == 0:
        print("\nNo tests found. Add tests/examples.yaml to skills.")
        return 0

    all_passed = (total_tests_passed == total_tests_run and
                  total_trigger_passed == total_trigger_run)

    if all_passed:
        print("\nAll tests PASSED")
        return 0
    else:
        print("\nSome tests FAILED")
        return 1


if __name__ == '__main__':
    exit(main())

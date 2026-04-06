package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSkillContent_ValidFrontmatterAndBody(t *testing.T) {
	content := `---
name: test-skill
description: A test skill
tags:
  - testing
  - example
triggers:
  keywords:
    - test
  phases:
    - run
---

# Test Skill Body

This is the body content.`

	skill, err := ParseSkillContent(content)
	if err != nil {
		t.Fatalf("ParseSkillContent failed: %v", err)
	}

	if skill.Name != "test-skill" {
		t.Errorf("Name = %q, want %q", skill.Name, "test-skill")
	}
	if skill.Description != "A test skill" {
		t.Errorf("Description = %q, want %q", skill.Description, "A test skill")
	}
	if len(skill.Tags) != 2 {
		t.Errorf("Tags length = %d, want 2", len(skill.Tags))
	}
	if skill.Tags[0] != "testing" {
		t.Errorf("Tags[0] = %q, want %q", skill.Tags[0], "testing")
	}
	if len(skill.Triggers.Keywords) != 1 || skill.Triggers.Keywords[0] != "test" {
		t.Errorf("Triggers.Keywords = %v, want [test]", skill.Triggers.Keywords)
	}
	if skill.Body == "" {
		t.Error("Body should not be empty")
	}
	if skill.Body != "# Test Skill Body\n\nThis is the body content." {
		t.Errorf("Body = %q", skill.Body)
	}
}

func TestParseSkillContent_MissingFrontmatter(t *testing.T) {
	content := `# Just a markdown file

No frontmatter here.`

	skill, err := ParseSkillContent(content)
	if err != nil {
		t.Fatalf("ParseSkillContent failed: %v", err)
	}

	// frontmatter가 없으면 전체 내용이 body로 처리됨
	if skill.Name != "" {
		t.Errorf("Name = %q, want empty", skill.Name)
	}
	if skill.Body != content {
		t.Errorf("Body = %q, want original content", skill.Body)
	}
}

func TestParseSkillContent_FrontmatterOnly(t *testing.T) {
	content := `---
name: metadata-only
description: No body
---`

	skill, err := ParseSkillContent(content)
	if err != nil {
		t.Fatalf("ParseSkillContent failed: %v", err)
	}

	if skill.Name != "metadata-only" {
		t.Errorf("Name = %q, want %q", skill.Name, "metadata-only")
	}
	if skill.Body != "" {
		t.Errorf("Body = %q, want empty", skill.Body)
	}
}

func TestSplitFrontmatter_WithFrontmatter(t *testing.T) {
	content := `---
name: my-skill
---

Body here.`

	fm, body, err := SplitFrontmatter(content)
	if err != nil {
		t.Fatalf("SplitFrontmatter failed: %v", err)
	}

	if fm != "name: my-skill" {
		t.Errorf("frontmatter = %q, want %q", fm, "name: my-skill")
	}
	if body != "Body here." {
		t.Errorf("body = %q, want %q", body, "Body here.")
	}
}

func TestSplitFrontmatter_WithoutFrontmatter(t *testing.T) {
	content := `Just plain markdown content.`

	fm, body, err := SplitFrontmatter(content)
	if err != nil {
		t.Fatalf("SplitFrontmatter failed: %v", err)
	}

	if fm != "" {
		t.Errorf("frontmatter = %q, want empty", fm)
	}
	if body != content {
		t.Errorf("body = %q, want original content", body)
	}
}

func TestSplitFrontmatter_EmptyContent(t *testing.T) {
	fm, body, err := SplitFrontmatter("")
	if err != nil {
		t.Fatalf("SplitFrontmatter failed: %v", err)
	}

	if fm != "" {
		t.Errorf("frontmatter = %q, want empty", fm)
	}
	if body != "" {
		t.Errorf("body = %q, want empty", body)
	}
}

func TestSplitFrontmatter_UnclosedDelimiter(t *testing.T) {
	// 닫는 delimiter가 없는 경우 전체가 frontmatter로 처리됨
	content := `---
name: unclosed
description: no closing delimiter`

	fm, body, err := SplitFrontmatter(content)
	if err != nil {
		t.Fatalf("SplitFrontmatter failed: %v", err)
	}

	if fm == "" {
		t.Error("frontmatter should not be empty for unclosed delimiter")
	}
	if body != "" {
		t.Errorf("body = %q, want empty for unclosed delimiter", body)
	}
}

func TestLoadMetadataOnly_FromTempFile(t *testing.T) {
	tmpDir := t.TempDir()
	skillFile := filepath.Join(tmpDir, "SKILL.md")

	content := `---
name: temp-skill
description: A temporary skill for testing
tags:
  - test
triggers:
  keywords:
    - temp
  phases:
    - plan
  agents: []
  languages: []
---

# Body content that should NOT be loaded`

	if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	skill, err := LoadMetadataOnly(skillFile)
	if err != nil {
		t.Fatalf("LoadMetadataOnly failed: %v", err)
	}

	if skill.Name != "temp-skill" {
		t.Errorf("Name = %q, want %q", skill.Name, "temp-skill")
	}
	if skill.Description != "A temporary skill for testing" {
		t.Errorf("Description = %q", skill.Description)
	}
	if skill.FilePath != skillFile {
		t.Errorf("FilePath = %q, want %q", skill.FilePath, skillFile)
	}
	// Level 1 로딩: Body는 비어 있어야 함
	if skill.Body != "" {
		t.Errorf("Body should be empty for metadata-only load, got %q", skill.Body)
	}
}

func TestLoadMetadataOnly_FileNotFound(t *testing.T) {
	_, err := LoadMetadataOnly("/nonexistent/path/SKILL.md")
	if err == nil {
		t.Error("LoadMetadataOnly should fail for nonexistent file")
	}
}

func TestLoadFromFile_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	skillFile := filepath.Join(tmpDir, "SKILL.md")

	content := `---
name: full-skill
description: Full load test
tags:
  - integration
triggers:
  keywords: []
  phases: []
  agents: []
  languages: []
---

# Full Body

This body should be loaded.`

	if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	skill, err := LoadFromFile(skillFile)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if skill.Name != "full-skill" {
		t.Errorf("Name = %q, want %q", skill.Name, "full-skill")
	}
	if skill.Body == "" {
		t.Error("Body should not be empty for full load")
	}
	if skill.FilePath != skillFile {
		t.Errorf("FilePath = %q, want %q", skill.FilePath, skillFile)
	}
}

func TestLoadFromDirectory_Recursive(t *testing.T) {
	tmpDir := t.TempDir()

	// 중첩 디렉토리에 SKILL.md 파일 생성
	subDir := filepath.Join(tmpDir, "sub", "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	skill1Content := `---
name: skill-root
description: Root skill
tags: []
triggers:
  keywords: []
  phases: []
  agents: []
  languages: []
---`

	skill2Content := `---
name: skill-nested
description: Nested skill
tags: []
triggers:
  keywords: []
  phases: []
  agents: []
  languages: []
---`

	os.WriteFile(filepath.Join(tmpDir, "SKILL.md"), []byte(skill1Content), 0644)
	os.WriteFile(filepath.Join(subDir, "SKILL.md"), []byte(skill2Content), 0644)

	skills, err := LoadFromDirectory(tmpDir, true)
	if err != nil {
		t.Fatalf("LoadFromDirectory failed: %v", err)
	}

	if len(skills) != 2 {
		t.Errorf("Found %d skills, want 2", len(skills))
	}
}

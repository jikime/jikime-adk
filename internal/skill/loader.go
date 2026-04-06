// Package skill provides SKILL.md file loading for jikime-adk.
package skill

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillFileName is the standard skill definition filename.
const SkillFileName = "SKILL.md"

// FrontmatterDelimiter is the YAML frontmatter delimiter.
const FrontmatterDelimiter = "---"

// LoadFromFile loads a Skill from a SKILL.md file.
// It parses the YAML frontmatter and extracts the markdown body.
func LoadFromFile(filePath string) (*Skill, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	skill, err := ParseSkillContent(string(content))
	if err != nil {
		return nil, err
	}

	skill.FilePath = filePath
	return skill, nil
}

// ParseSkillContent parses SKILL.md content into a Skill struct.
// Extracts YAML frontmatter and markdown body.
func ParseSkillContent(content string) (*Skill, error) {
	frontmatter, body, err := SplitFrontmatter(content)
	if err != nil {
		return nil, err
	}

	skill := &Skill{}
	if err := yaml.Unmarshal([]byte(frontmatter), skill); err != nil {
		return nil, err
	}

	skill.Body = body
	return skill, nil
}

// SplitFrontmatter splits content into YAML frontmatter and markdown body.
// Returns (frontmatter, body, error).
func SplitFrontmatter(content string) (string, string, error) {
	content = strings.TrimSpace(content)

	// Check if content starts with frontmatter delimiter
	if !strings.HasPrefix(content, FrontmatterDelimiter) {
		return "", content, nil
	}

	// Remove leading delimiter
	content = content[len(FrontmatterDelimiter):]

	// Find closing delimiter
	endIdx := strings.Index(content, "\n"+FrontmatterDelimiter)
	if endIdx == -1 {
		// No closing delimiter, treat entire content as frontmatter
		return strings.TrimSpace(content), "", nil
	}

	frontmatter := strings.TrimSpace(content[:endIdx])
	body := strings.TrimSpace(content[endIdx+len("\n"+FrontmatterDelimiter):])

	return frontmatter, body, nil
}

// LoadFromDirectory loads all SKILL.md files from a directory.
// recursive: whether to search subdirectories
func LoadFromDirectory(directory string, recursive bool) ([]*Skill, error) {
	var skills []*Skill

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on error
		}

		// Skip directories
		if info.IsDir() {
			// If not recursive and not the root, skip subdirectories
			if !recursive && path != directory {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if it's a SKILL.md file
		if filepath.Base(path) == SkillFileName {
			skill, err := LoadFromFile(path)
			if err != nil {
				// Log error but continue
				return nil
			}
			skills = append(skills, skill)
		}

		return nil
	}

	if err := filepath.Walk(directory, walkFn); err != nil {
		return nil, err
	}

	return skills, nil
}

// LoadSkillsDir loads all skills from a .claude/skills directory.
// This is the standard location for skills in jikime-adk projects.
func LoadSkillsDir(projectRoot string) ([]*Skill, error) {
	skillsDir := filepath.Join(projectRoot, ".claude", "skills")
	return LoadFromDirectory(skillsDir, true)
}

// LoadMetadataOnly loads only the YAML frontmatter (Level 1 - ~100 tokens).
// Use this for initial skill discovery without loading the full body.
func LoadMetadataOnly(filePath string) (*Skill, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	frontmatter, _, err := SplitFrontmatter(string(content))
	if err != nil {
		return nil, err
	}

	skill := &Skill{}
	if err := yaml.Unmarshal([]byte(frontmatter), skill); err != nil {
		return nil, err
	}

	skill.FilePath = filePath
	// Body is intentionally not loaded (Level 1)
	return skill, nil
}

// LoadMetadataFromDirectory loads only metadata from all SKILL.md files.
// More efficient for initial discovery and filtering.
func LoadMetadataFromDirectory(directory string, recursive bool) ([]*Skill, error) {
	var skills []*Skill

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if !recursive && path != directory {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Base(path) == SkillFileName {
			skill, err := LoadMetadataOnly(path)
			if err != nil {
				return nil
			}
			skills = append(skills, skill)
		}

		return nil
	}

	if err := filepath.Walk(directory, walkFn); err != nil {
		return nil, err
	}

	return skills, nil
}

// SerializeSkill converts a Skill back to SKILL.md format.
func SerializeSkill(skill *Skill) ([]byte, error) {
	var buf bytes.Buffer

	// Write opening frontmatter delimiter
	buf.WriteString(FrontmatterDelimiter + "\n")

	// Serialize YAML frontmatter
	yamlData, err := yaml.Marshal(skill)
	if err != nil {
		return nil, err
	}
	buf.Write(yamlData)

	// Write closing frontmatter delimiter
	buf.WriteString(FrontmatterDelimiter + "\n")

	// Write body if present
	if skill.Body != "" {
		buf.WriteString("\n")
		buf.WriteString(skill.Body)
	}

	return buf.Bytes(), nil
}

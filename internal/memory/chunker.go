package memory

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// bytesPerToken is the approximate ratio of bytes to tokens for English text.
const bytesPerToken = 4

// ChunkFile splits markdown content into heading-aware chunks.
// It prefers to split at ## heading boundaries, then at paragraph boundaries
// (blank lines), and finally at line boundaries if sections are still too large.
// Each chunk includes ~overlapBytes worth of trailing lines from the previous chunk.
func ChunkFile(path string, content []byte, opts ChunkOpts) []Chunk {
	if opts.MaxTokens == 0 {
		opts = DefaultChunkOpts()
	}

	maxBytes := opts.MaxTokens * bytesPerToken
	overlapBytes := opts.Overlap * bytesPerToken

	lines := strings.Split(string(content), "\n")
	sections := splitByHeadings(lines)

	var chunks []Chunk
	var prevTailLines []string // lines carried forward as overlap

	for _, sec := range sections {
		blocks := splitLargeSection(sec.lines, maxBytes)
		for _, block := range blocks {
			// Prepend overlap lines from the previous chunk.
			var merged []string
			if len(prevTailLines) > 0 {
				merged = append(merged, prevTailLines...)
			}
			merged = append(merged, block...)

			text := strings.Join(merged, "\n")
			if len(text) < opts.MinChunkSize {
				continue
			}

			// Compute start/end line numbers (1-based, relative to original content).
			startLine := sec.startLine
			if len(prevTailLines) > 0 {
				// The overlap lines come from before this section's block,
				// but the canonical start is the block's own start.
				startLine = sec.startLine + lineOffset(sec.lines, block)
			} else {
				startLine = sec.startLine + lineOffset(sec.lines, block)
			}
			endLine := startLine + len(block) - 1

			hash := sha256Hash(text)

			chunks = append(chunks, Chunk{
				Path:      path,
				StartLine: startLine,
				EndLine:   endLine,
				Text:      text,
				Hash:      hash,
				Heading:   sec.heading,
			})

			// Prepare overlap: take trailing lines from this block up to overlapBytes.
			prevTailLines = tailLines(block, overlapBytes)
		}
	}

	return chunks
}

// section groups lines under a heading.
type section struct {
	heading   string
	lines     []string
	startLine int // 1-based line number in the original file
}

// splitByHeadings groups lines into sections delimited by ## headings.
// Lines before the first heading form a section with an empty heading.
func splitByHeadings(lines []string) []section {
	var sections []section
	cur := section{startLine: 1}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			// Flush current section if it has content.
			if len(cur.lines) > 0 {
				sections = append(sections, cur)
			}
			cur = section{
				heading:   strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")),
				lines:     []string{line},
				startLine: i + 1, // 1-based
			}
		} else {
			cur.lines = append(cur.lines, line)
		}
	}
	if len(cur.lines) > 0 {
		sections = append(sections, cur)
	}

	return sections
}

// splitLargeSection splits a section's lines into blocks that each fit within maxBytes.
// It first tries to split at paragraph boundaries (blank lines), then at line boundaries.
func splitLargeSection(lines []string, maxBytes int) [][]string {
	text := strings.Join(lines, "\n")
	if len(text) <= maxBytes {
		return [][]string{lines}
	}

	// Try splitting at paragraph boundaries.
	blocks := splitAtParagraphs(lines, maxBytes)

	// Further split any blocks that are still too large.
	var result [][]string
	for _, block := range blocks {
		blockText := strings.Join(block, "\n")
		if len(blockText) <= maxBytes {
			result = append(result, block)
		} else {
			result = append(result, splitAtLines(block, maxBytes)...)
		}
	}

	return result
}

// splitAtParagraphs splits lines into groups at blank-line boundaries,
// merging consecutive groups that together fit within maxBytes.
func splitAtParagraphs(lines []string, maxBytes int) [][]string {
	// Identify paragraph groups (separated by blank lines).
	var paragraphs [][]string
	var current []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if len(current) > 0 {
				paragraphs = append(paragraphs, current)
				current = nil
			}
			// Keep the blank line attached to the next paragraph or current group.
			current = append(current, line)
		} else {
			current = append(current, line)
		}
	}
	if len(current) > 0 {
		paragraphs = append(paragraphs, current)
	}

	// Merge paragraphs into blocks that fit within maxBytes.
	var blocks [][]string
	var block []string

	for _, para := range paragraphs {
		candidate := append(block, para...)
		if len(strings.Join(candidate, "\n")) > maxBytes && len(block) > 0 {
			blocks = append(blocks, block)
			block = para
		} else {
			block = candidate
		}
	}
	if len(block) > 0 {
		blocks = append(blocks, block)
	}

	return blocks
}

// splitAtLines splits lines into blocks of at most maxBytes each.
func splitAtLines(lines []string, maxBytes int) [][]string {
	var blocks [][]string
	var block []string
	size := 0

	for _, line := range lines {
		lineSize := len(line) + 1 // +1 for newline
		if size+lineSize > maxBytes && len(block) > 0 {
			blocks = append(blocks, block)
			block = nil
			size = 0
		}
		block = append(block, line)
		size += lineSize
	}
	if len(block) > 0 {
		blocks = append(blocks, block)
	}

	return blocks
}

// tailLines returns the last N lines from lines that together don't exceed maxBytes.
func tailLines(lines []string, maxBytes int) []string {
	if maxBytes <= 0 || len(lines) == 0 {
		return nil
	}

	size := 0
	start := len(lines)
	for i := len(lines) - 1; i >= 0; i-- {
		lineSize := len(lines[i]) + 1 // +1 for newline
		if size+lineSize > maxBytes {
			break
		}
		size += lineSize
		start = i
	}

	if start >= len(lines) {
		return nil
	}

	result := make([]string, len(lines)-start)
	copy(result, lines[start:])
	return result
}

// lineOffset finds the index of block's first line within sectionLines.
// Returns 0 if not found.
func lineOffset(sectionLines []string, block []string) int {
	if len(block) == 0 || len(sectionLines) == 0 {
		return 0
	}
	target := block[0]
	for i, line := range sectionLines {
		if line == target {
			// Verify consecutive match.
			match := true
			for j := 1; j < len(block) && i+j < len(sectionLines); j++ {
				if sectionLines[i+j] != block[j] {
					match = false
					break
				}
			}
			if match {
				return i
			}
		}
	}
	return 0
}

// sha256Hash returns the hex-encoded SHA256 hash of s.
func sha256Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

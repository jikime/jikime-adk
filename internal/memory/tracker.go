package memory

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
)

const maxTrackFileSize = 1 << 20 // 1MB

// TrackBufferPath returns the buffer file path for PostToolUse tracking.
func TrackBufferPath(projectDir string) string {
	return filepath.Join(projectDir, ".jikime", "memory", "track_buffer.jsonl")
}

// AppendTrack writes a file modification record to the buffer file.
// Uses O_APPEND for atomic append.
func AppendTrack(projectDir string, record FileTrackRecord) error {
	bufPath := TrackBufferPath(projectDir)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(bufPath), 0755); err != nil {
		return err
	}

	// Check file size limit
	if info, err := os.Stat(bufPath); err == nil && info.Size() > maxTrackFileSize {
		return nil // silently skip if too large
	}

	f, err := os.OpenFile(bufPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	_, err = f.Write(data)
	return err
}

// FlushTrack reads and clears the buffer, returns all records.
func FlushTrack(projectDir string) ([]FileTrackRecord, error) {
	bufPath := TrackBufferPath(projectDir)

	f, err := os.Open(bufPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var records []FileTrackRecord
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var rec FileTrackRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			continue
		}
		records = append(records, rec)
	}

	// Close before removing
	f.Close()
	os.Remove(bufPath)

	return records, nil
}

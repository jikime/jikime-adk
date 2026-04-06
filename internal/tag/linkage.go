// Package tag provides TAG linkage management for TAG System v2.0.
package tag

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// LinkageDatabase represents the structure of the linkage database.
type LinkageDatabase struct {
	Tags  []map[string]interface{} `json:"tags"`
	Files map[string][]string      `json:"files"`
}

// LinkageManager manages bidirectional TAG↔CODE mapping database.
// Provides atomic database operations for TAG tracking.
type LinkageManager struct {
	DBPath string
	mu     sync.RWMutex
}

// NewLinkageManager creates a new LinkageManager instance.
func NewLinkageManager(dbPath string) (*LinkageManager, error) {
	lm := &LinkageManager{DBPath: dbPath}
	if err := lm.ensureDatabase(); err != nil {
		return nil, err
	}
	return lm, nil
}

// ensureDatabase ensures the database file exists with proper structure.
func (lm *LinkageManager) ensureDatabase() error {
	if _, err := os.Stat(lm.DBPath); os.IsNotExist(err) {
		return lm.writeDatabase(&LinkageDatabase{
			Tags:  []map[string]interface{}{},
			Files: map[string][]string{},
		})
	}
	return nil
}

// loadDatabase loads the database from disk.
func (lm *LinkageManager) loadDatabase() (*LinkageDatabase, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if _, err := os.Stat(lm.DBPath); os.IsNotExist(err) {
		return &LinkageDatabase{
			Tags:  []map[string]interface{}{},
			Files: map[string][]string{},
		}, nil
	}

	data, err := os.ReadFile(lm.DBPath)
	if err != nil {
		return nil, err
	}

	var db LinkageDatabase
	if err := json.Unmarshal(data, &db); err != nil {
		return &LinkageDatabase{
			Tags:  []map[string]interface{}{},
			Files: map[string][]string{},
		}, nil
	}

	// Ensure structure
	if db.Tags == nil {
		db.Tags = []map[string]interface{}{}
	}
	if db.Files == nil {
		db.Files = map[string][]string{}
	}

	return &db, nil
}

// writeDatabase writes the database to disk atomically.
func (lm *LinkageManager) writeDatabase(db *LinkageDatabase) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(lm.DBPath), 0755); err != nil {
		return err
	}

	return AtomicWriteJSON(lm.DBPath, db)
}

// AddTag adds a TAG to the linkage database.
func (lm *LinkageManager) AddTag(tag *TAG) error {
	db, err := lm.loadDatabase()
	if err != nil {
		return err
	}

	// Create TAG entry
	tagEntry := tag.ToMap()

	// Check for duplicate
	duplicate := false
	for _, existing := range db.Tags {
		if equalMaps(existing, tagEntry) {
			duplicate = true
			break
		}
	}

	if !duplicate {
		db.Tags = append(db.Tags, tagEntry)
	}

	// Update file index
	fileKey := tag.FilePath
	if db.Files[fileKey] == nil {
		db.Files[fileKey] = []string{}
	}

	// Add spec_id to file index if not present
	found := false
	for _, id := range db.Files[fileKey] {
		if id == tag.SpecID {
			found = true
			break
		}
	}
	if !found {
		db.Files[fileKey] = append(db.Files[fileKey], tag.SpecID)
	}

	return lm.writeDatabase(db)
}

// RemoveTag removes a specific TAG from the database.
func (lm *LinkageManager) RemoveTag(tag *TAG) error {
	db, err := lm.loadDatabase()
	if err != nil {
		return err
	}

	tagEntry := tag.ToMap()

	// Remove from tags list
	newTags := make([]map[string]interface{}, 0, len(db.Tags))
	for _, existing := range db.Tags {
		if !equalMaps(existing, tagEntry) {
			newTags = append(newTags, existing)
		}
	}
	db.Tags = newTags

	// Update file index
	fileKey := tag.FilePath
	if ids, ok := db.Files[fileKey]; ok {
		newIDs := make([]string, 0, len(ids))
		for _, id := range ids {
			if id != tag.SpecID {
				newIDs = append(newIDs, id)
			}
		}
		if len(newIDs) == 0 {
			delete(db.Files, fileKey)
		} else {
			db.Files[fileKey] = newIDs
		}
	}

	return lm.writeDatabase(db)
}

// RemoveFileTags removes all TAGs for a file.
func (lm *LinkageManager) RemoveFileTags(filePath string) error {
	db, err := lm.loadDatabase()
	if err != nil {
		return err
	}

	// Remove all TAGs for this file
	newTags := make([]map[string]interface{}, 0, len(db.Tags))
	for _, tag := range db.Tags {
		if fp, ok := tag["file_path"].(string); !ok || fp != filePath {
			newTags = append(newTags, tag)
		}
	}
	db.Tags = newTags

	// Remove file index
	delete(db.Files, filePath)

	return lm.writeDatabase(db)
}

// GetAllTags returns all TAGs in the database.
func (lm *LinkageManager) GetAllTags() []*TAG {
	db, err := lm.loadDatabase()
	if err != nil {
		return []*TAG{}
	}

	tags := make([]*TAG, 0, len(db.Tags))
	for _, entry := range db.Tags {
		tag := TAGFromMap(entry)
		if tag != nil {
			tags = append(tags, tag)
		}
	}

	return tags
}

// GetCodeLocations returns all code locations for a SPEC-ID.
func (lm *LinkageManager) GetCodeLocations(specID string) []map[string]interface{} {
	db, err := lm.loadDatabase()
	if err != nil {
		return []map[string]interface{}{}
	}

	var locations []map[string]interface{}
	for _, tag := range db.Tags {
		if id, ok := tag["spec_id"].(string); ok && id == specID {
			locations = append(locations, map[string]interface{}{
				"file_path": tag["file_path"],
				"line":      tag["line"],
				"verb":      tag["verb"],
			})
		}
	}

	return locations
}

// GetTagsByFile returns all TAGs for a specific file.
func (lm *LinkageManager) GetTagsByFile(filePath string) []*TAG {
	db, err := lm.loadDatabase()
	if err != nil {
		return []*TAG{}
	}

	var tags []*TAG
	for _, entry := range db.Tags {
		if fp, ok := entry["file_path"].(string); ok && fp == filePath {
			tag := TAGFromMap(entry)
			if tag != nil {
				tags = append(tags, tag)
			}
		}
	}

	return tags
}

// GetAllSpecIDs returns all unique SPEC-IDs in the database.
func (lm *LinkageManager) GetAllSpecIDs() []string {
	db, err := lm.loadDatabase()
	if err != nil {
		return []string{}
	}

	specIDSet := make(map[string]bool)
	for _, tag := range db.Tags {
		if id, ok := tag["spec_id"].(string); ok {
			specIDSet[id] = true
		}
	}

	specIDs := make([]string, 0, len(specIDSet))
	for id := range specIDSet {
		specIDs = append(specIDs, id)
	}
	sort.Strings(specIDs)

	return specIDs
}

// FindOrphanedTags finds TAGs referencing nonexistent SPEC documents.
func (lm *LinkageManager) FindOrphanedTags(specsDir string) []*TAG {
	allTags := lm.GetAllTags()

	var orphans []*TAG
	for _, tag := range allTags {
		specDir := filepath.Join(specsDir, tag.SpecID)
		if _, err := os.Stat(specDir); os.IsNotExist(err) {
			orphans = append(orphans, tag)
		}
	}

	return orphans
}

// Clear removes all TAGs from the database.
func (lm *LinkageManager) Clear() error {
	return lm.writeDatabase(&LinkageDatabase{
		Tags:  []map[string]interface{}{},
		Files: map[string][]string{},
	})
}

// equalMaps compares two maps for equality.
func equalMaps(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		// Handle numeric comparison (JSON unmarshals numbers as float64)
		switch av := v.(type) {
		case float64:
			if bvf, ok := bv.(float64); ok {
				if av != bvf {
					return false
				}
			} else if bvi, ok := bv.(int); ok {
				if int(av) != bvi {
					return false
				}
			} else {
				return false
			}
		case int:
			if bvf, ok := bv.(float64); ok {
				if av != int(bvf) {
					return false
				}
			} else if bvi, ok := bv.(int); ok {
				if av != bvi {
					return false
				}
			} else {
				return false
			}
		default:
			if v != bv {
				return false
			}
		}
	}
	return true
}

// SpecDocumentExists checks if SPEC document exists in specs directory.
func SpecDocumentExists(specID, specsDir string) bool {
	specDir := filepath.Join(specsDir, specID)
	info, err := os.Stat(specDir)
	return err == nil && info.IsDir()
}

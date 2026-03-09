// Package workflow handles WORKFLOW.md parsing and config loading.
// Implements Symphony SPEC Sections 5 and 6.
package workflow

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"

	"jikime-adk/internal/serve"
)

// Loader loads and hot-reloads a WORKFLOW.md file.
type Loader struct {
	path     string
	mu       sync.RWMutex
	current  *serve.WorkflowDefinition
	watcher  *fsnotify.Watcher
	onChange func(*serve.WorkflowDefinition)
}

// NewLoader creates a Loader, parses the file, and starts watching for changes.
func NewLoader(path string, onChange func(*serve.WorkflowDefinition)) (*Loader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create watcher: %w", err)
	}

	l := &Loader{
		path:     path,
		watcher:  watcher,
		onChange: onChange,
	}

	def, err := l.load()
	if err != nil {
		watcher.Close()
		return nil, err
	}
	l.current = def

	if err := watcher.Add(path); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("watch %s: %w", path, err)
	}

	go l.watchLoop()
	return l, nil
}

// Current returns the current workflow definition (thread-safe).
func (l *Loader) Current() *serve.WorkflowDefinition {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.current
}

// Close stops the file watcher.
func (l *Loader) Close() {
	l.watcher.Close()
}

// load reads and parses the WORKFLOW.md file.
func (l *Loader) load() (*serve.WorkflowDefinition, error) {
	data, err := os.ReadFile(l.path)
	if err != nil {
		return nil, fmt.Errorf("missing_workflow_file: %w", err)
	}
	return Parse(data)
}

// Parse parses WORKFLOW.md content: YAML front matter + Liquid-style prompt body.
// If no front matter (---), the entire file is the prompt body.
func Parse(data []byte) (*serve.WorkflowDefinition, error) {
	content := string(data)
	config := map[string]any{}
	promptTemplate := ""

	if strings.HasPrefix(content, "---") {
		rest := content[3:]
		// Require newline after opening ---
		newlineIdx := strings.Index(rest, "\n")
		if newlineIdx == -1 {
			return nil, fmt.Errorf("workflow_parse_error: invalid front matter")
		}
		rest = rest[newlineIdx+1:]

		// Find closing ---
		closeIdx := strings.Index(rest, "\n---")
		if closeIdx == -1 {
			return nil, fmt.Errorf("workflow_parse_error: unclosed front matter (missing closing ---)")
		}
		frontMatter := rest[:closeIdx]
		promptTemplate = strings.TrimSpace(rest[closeIdx+4:])

		dec := yaml.NewDecoder(bytes.NewBufferString(frontMatter))
		if err := dec.Decode(&config); err != nil {
			return nil, fmt.Errorf("workflow_parse_error: %w", err)
		}
		if config == nil {
			return nil, fmt.Errorf("workflow_front_matter_not_a_map")
		}
	} else {
		promptTemplate = strings.TrimSpace(content)
	}

	return &serve.WorkflowDefinition{
		Config:         config,
		PromptTemplate: promptTemplate,
	}, nil
}

func (l *Loader) watchLoop() {
	for {
		select {
		case event, ok := <-l.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				def, err := l.load()
				if err != nil {
					// Keep last good config on invalid reload
					continue
				}
				l.mu.Lock()
				l.current = def
				l.mu.Unlock()
				if l.onChange != nil {
					l.onChange(def)
				}
			}
		case _, ok := <-l.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

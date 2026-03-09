package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ABC-123", "ABC-123"},
		{"owner/repo#42", "owner_repo_42"},
		{"PROJ-1 (fix)", "PROJ-1__fix_"},
		{"normal.key-123", "normal.key-123"},
		{"../../etc/passwd", ".._.._etc_passwd"},
	}
	for _, tt := range tests {
		got := sanitizeKey(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestManager_CreateForIssue_NewDirectory(t *testing.T) {
	root := t.TempDir()
	m := NewManager(root)

	ws, err := m.CreateForIssue("PROJ-42")
	if err != nil {
		t.Fatalf("CreateForIssue() error = %v", err)
	}

	if !ws.CreatedNow {
		t.Error("expected CreatedNow = true for new directory")
	}
	if ws.Key != "PROJ-42" {
		t.Errorf("Key = %q, want PROJ-42", ws.Key)
	}

	expectedPath := filepath.Join(root, "PROJ-42")
	if ws.Path != expectedPath {
		t.Errorf("Path = %q, want %q", ws.Path, expectedPath)
	}

	info, err := os.Stat(ws.Path)
	if err != nil {
		t.Fatalf("workspace dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("workspace path is not a directory")
	}
}

func TestManager_CreateForIssue_ExistingDirectory(t *testing.T) {
	root := t.TempDir()
	m := NewManager(root)

	// First call — creates
	ws1, err := m.CreateForIssue("PROJ-42")
	if err != nil {
		t.Fatalf("first CreateForIssue() error = %v", err)
	}
	if !ws1.CreatedNow {
		t.Error("first call: expected CreatedNow = true")
	}

	// Second call — reuses
	ws2, err := m.CreateForIssue("PROJ-42")
	if err != nil {
		t.Fatalf("second CreateForIssue() error = %v", err)
	}
	if ws2.CreatedNow {
		t.Error("second call: expected CreatedNow = false (reuse)")
	}
	if ws1.Path != ws2.Path {
		t.Errorf("paths differ: %q vs %q", ws1.Path, ws2.Path)
	}
}

func TestManager_CreateForIssue_SanitizesIdentifier(t *testing.T) {
	root := t.TempDir()
	m := NewManager(root)

	ws, err := m.CreateForIssue("owner/repo#42")
	if err != nil {
		t.Fatalf("CreateForIssue() error = %v", err)
	}
	if ws.Key != "owner_repo_42" {
		t.Errorf("Key = %q, want owner_repo_42", ws.Key)
	}
}

func TestManager_PathContainmentInvariant(t *testing.T) {
	root := t.TempDir()
	m := NewManager(root)

	// Path traversal attempt via identifier
	_, err := m.CreateForIssue("../../etc/passwd")
	// Should succeed but be safely contained under root
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it didn't escape root
	wsPath, err := m.PathForIssue("../../etc/passwd")
	if err != nil {
		t.Fatalf("PathForIssue error: %v", err)
	}

	absRoot, _ := filepath.Abs(root)
	if len(wsPath) <= len(absRoot) {
		t.Errorf("workspace path %q appears to escape root %q", wsPath, absRoot)
	}
}

func TestManager_AfterCreateHook_FatalOnFailure(t *testing.T) {
	root := t.TempDir()
	m := NewManager(root,
		WithHookAfterCreate("exit 1"), // always fails
		WithHookTimeoutMS(5000),
	)

	_, err := m.CreateForIssue("HOOK-FAIL")
	if err == nil {
		t.Error("expected error when after_create hook fails")
	}

	// Workspace should be cleaned up after failed after_create
	wsPath := filepath.Join(root, "HOOK-FAIL")
	if _, statErr := os.Stat(wsPath); !os.IsNotExist(statErr) {
		t.Error("workspace directory should be removed after failed after_create hook")
	}
}

func TestManager_AfterCreateHook_RunsOnlyOnce(t *testing.T) {
	root := t.TempDir()
	counter := filepath.Join(root, "counter.txt")

	// Hook appends a line each time it runs
	hook := "echo 'ran' >> " + counter
	m := NewManager(root,
		WithHookAfterCreate(hook),
		WithHookTimeoutMS(5000),
	)

	// First call — hook runs
	if _, err := m.CreateForIssue("ONCE-42"); err != nil {
		t.Fatalf("first call error: %v", err)
	}

	// Second call — hook must NOT run again
	if _, err := m.CreateForIssue("ONCE-42"); err != nil {
		t.Fatalf("second call error: %v", err)
	}

	data, _ := os.ReadFile(counter)
	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	if lines != 1 {
		t.Errorf("after_create hook ran %d times, want exactly 1", lines)
	}
}

func TestManager_BeforeRunHook_FatalOnFailure(t *testing.T) {
	root := t.TempDir()
	wsPath := filepath.Join(root, "ws")
	os.MkdirAll(wsPath, 0o755)

	m := NewManager(root, WithHookBeforeRun("exit 1"), WithHookTimeoutMS(5000))
	err := m.BeforeRun(wsPath)
	if err == nil {
		t.Error("expected error when before_run hook fails")
	}
}

func TestManager_CleanupForIssue(t *testing.T) {
	root := t.TempDir()
	m := NewManager(root)

	ws, _ := m.CreateForIssue("CLEAN-99")
	if _, err := os.Stat(ws.Path); err != nil {
		t.Fatalf("workspace not created: %v", err)
	}

	m.CleanupForIssue("CLEAN-99")

	if _, err := os.Stat(ws.Path); !os.IsNotExist(err) {
		t.Error("workspace should be removed after CleanupForIssue")
	}
}

func TestManager_HookTimeout(t *testing.T) {
	root := t.TempDir()
	m := NewManager(root,
		WithHookAfterCreate("sleep 10"), // longer than timeout
		WithHookTimeoutMS(100),          // 100ms timeout
	)

	_, err := m.CreateForIssue("TIMEOUT-1")
	if err == nil {
		t.Error("expected timeout error")
	}
}

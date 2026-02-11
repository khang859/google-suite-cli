package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestInstallSkillFiles(t *testing.T) {
	t.Parallel()

	fakeFS := fstest.MapFS{
		"skills/gsuite-manager/SKILL.md":               {Data: []byte("# Skill")},
		"skills/gsuite-manager/references/commands.md": {Data: []byte("# Commands")},
	}

	tests := []struct {
		name     string
		client   string
		wantBase string
	}{
		{
			name:     "claude client writes to .claude/skills",
			client:   "claude",
			wantBase: filepath.Join(".claude", "skills", "gsuite-manager"),
		},
		{
			name:     "openclaw-workspace writes to skills",
			client:   "openclaw-workspace",
			wantBase: filepath.Join("skills", "gsuite-manager"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			targetDir := filepath.Join(tmpDir, tt.wantBase)

			err := installSkillFiles(fakeFS, "skills/gsuite-manager", targetDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			wantFiles := []struct {
				rel     string
				content string
			}{
				{"SKILL.md", "# Skill"},
				{filepath.Join("references", "commands.md"), "# Commands"},
			}
			for _, wf := range wantFiles {
				path := filepath.Join(targetDir, wf.rel)
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("expected file %s: %v", wf.rel, err)
				}
				if string(data) != wf.content {
					t.Errorf("file %s = %q, want %q", wf.rel, string(data), wf.content)
				}
			}
		})
	}
}

func TestInstallSkillFilesOverwritesExisting(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "skills", "gsuite-manager")

	// Write an old file that should be overwritten.
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "SKILL.md"), []byte("old content"), 0o644); err != nil {
		t.Fatal(err)
	}

	fakeFS := fstest.MapFS{
		"skills/gsuite-manager/SKILL.md": {Data: []byte("new content")},
	}

	if err := installSkillFiles(fakeFS, "skills/gsuite-manager", targetDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new content" {
		t.Errorf("file content = %q, want %q", string(data), "new content")
	}
}

func TestRunInstallSkillUnknownClient(t *testing.T) {
	t.Parallel()

	cmd := installSkillCmd
	if err := cmd.Flags().Set("client", "nonexistent"); err != nil {
		t.Fatalf("setting flag: %v", err)
	}
	t.Cleanup(func() { cmd.Flags().Set("client", "claude") })

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error for unknown client, got nil")
	}
	if !strings.Contains(err.Error(), `unknown client "nonexistent"`) {
		t.Fatalf("error = %q, want it to contain unknown client message", err.Error())
	}
}

func TestClientSkillDirsContainsExpectedEntries(t *testing.T) {
	t.Parallel()

	expected := []string{"claude", "openclaw-workspace"}
	for _, name := range expected {
		if _, ok := clientSkillDirs[name]; !ok {
			t.Errorf("clientSkillDirs missing entry for %q", name)
		}
	}
}

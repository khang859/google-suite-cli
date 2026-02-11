package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// SkillFS holds the embedded skill files, set by main.go before Execute().
var SkillFS embed.FS

// clientSkillDirs maps client names to their skills directory prefix.
var clientSkillDirs = map[string]string{
	"claude":              filepath.Join(".claude", "skills"),
	"openclaw-workspace":  "skills",
}

var installSkillCmd = &cobra.Command{
	Use:   "install-skill",
	Short: "Install the Claude Code skill for Gmail management",
	Long: `Writes the bundled gsuite-manager skill files into the appropriate skills directory.

By default, installs to .claude/skills/gsuite-manager/ (Claude Code).
Use --client to target a different client:
  --client openclaw-workspace  → skills/gsuite-manager/

Existing files are overwritten.`,
	RunE: runInstallSkill,
}

func init() {
	installSkillCmd.Flags().String("client", "claude", "target client (claude, openclaw-workspace)")
	rootCmd.AddCommand(installSkillCmd)
}

func runInstallSkill(cmd *cobra.Command, args []string) error {
	client, _ := cmd.Flags().GetString("client")
	skillsDir, ok := clientSkillDirs[client]
	if !ok {
		return fmt.Errorf("unknown client %q (supported: claude, openclaw-workspace)", client)
	}
	targetDir := filepath.Join(skillsDir, "gsuite-manager")

	return installSkillFiles(SkillFS, "skills/gsuite-manager", targetDir)
}

func installSkillFiles(source fs.FS, embeddedRoot, targetDir string) error {
	sub, err := fs.Sub(source, embeddedRoot)
	if err != nil {
		return fmt.Errorf("reading embedded skill files: %w", err)
	}

	return fs.WalkDir(sub, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		dest := filepath.Join(targetDir, path)

		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}

		data, err := fs.ReadFile(sub, path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", dest, err)
		}

		fmt.Println("  wrote", dest)
		return nil
	})
}

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

var installSkillCmd = &cobra.Command{
	Use:   "install-skill",
	Short: "Install the Claude Code skill for Gmail management",
	Long: `Writes the bundled gsuite-manager skill files into .claude/skills/gsuite-manager/
in the current working directory. Existing files are overwritten.`,
	RunE: runInstallSkill,
}

func init() {
	rootCmd.AddCommand(installSkillCmd)
}

func runInstallSkill(cmd *cobra.Command, args []string) error {
	const embeddedRoot = "skills/gsuite-manager"
	targetDir := filepath.Join(".claude", "skills", "gsuite-manager")

	sub, err := fs.Sub(SkillFS, embeddedRoot)
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

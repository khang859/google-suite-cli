package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags
var Version = "dev"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gsuite",
	Long:  `Display the version number of the gsuite CLI tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gsuite version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

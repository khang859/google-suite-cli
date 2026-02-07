package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose      bool
	outputFormat string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gsuite",
	Short: "Gmail CLI tool",
	Long: `gsuite is a command-line interface for Gmail mailbox management.

Authenticate with 'gsuite login' to get started.

Provides full access to Gmail operations including reading, sending, searching,
and managing messages, threads, labels, and drafts.

Designed for automation workflows and scripting with support for both
human-readable and JSON output formats.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "text", "Output format: text or json")
}

// GetVerbose returns whether verbose mode is enabled.
func GetVerbose() bool {
	return verbose
}

// GetOutputFormat returns the output format from the --format flag.
func GetOutputFormat() string {
	return outputFormat
}

// outputJSON marshals v as indented JSON and prints it to stdout.
func outputJSON(v interface{}) error {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonBytes))
	return nil
}

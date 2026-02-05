package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// verbose flag for future use
	verbose bool
	// credentialsFile is the path to service account JSON credentials
	credentialsFile string
	// userEmail is the email of the user to impersonate
	userEmail string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gsuite",
	Short: "Gmail CLI tool",
	Long: `gsuite is a command-line interface for Gmail mailbox management.

It uses service account authentication with domain-wide delegation to provide
full access to Gmail operations including reading, sending, searching, and
managing messages, threads, labels, and drafts.

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
	// Persistent flags available to all subcommands
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&credentialsFile, "credentials-file", "c", "", "Path to service account JSON credentials")
	rootCmd.PersistentFlags().StringVarP(&userEmail, "user", "u", "", "Email of user to impersonate")
}

// GetCredentialsFile returns the credentials file path from the --credentials-file flag.
func GetCredentialsFile() string {
	return credentialsFile
}

// GetUserEmail returns the user email from the --user flag.
func GetUserEmail() string {
	return userEmail
}

// GetVerbose returns whether verbose mode is enabled.
func GetVerbose() bool {
	return verbose
}

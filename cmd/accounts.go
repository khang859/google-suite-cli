package cmd

import (
	"fmt"
	"strings"

	"github.com/khang/google-suite-cli/internal/auth"
	"github.com/spf13/cobra"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Manage authenticated Gmail accounts",
	Long:  "List, switch between, or remove authenticated Gmail accounts.",
}

var accountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all authenticated accounts",
	Long: `List all authenticated Gmail accounts.

Displays each account's email and the date it was added.
The active account is marked with an asterisk (*).`,
	Example: `  # List all accounts
  gsuite accounts list

  # List in JSON format
  gsuite accounts list -f json`,
	RunE: runAccountsList,
}

var accountsSwitchCmd = &cobra.Command{
	Use:   "switch <email>",
	Short: "Switch the active account",
	Long: `Switch the active Gmail account.

The specified email must be an already authenticated account.
Use 'gsuite accounts list' to see available accounts.`,
	Example: `  # Switch active account
  gsuite accounts switch user@gmail.com`,
	Args: cobra.ExactArgs(1),
	RunE: runAccountsSwitch,
}

var accountsRemoveCmd = &cobra.Command{
	Use:   "remove <email>",
	Short: "Remove an authenticated account",
	Long: `Remove an authenticated Gmail account and its stored token.

If the removed account was the active account, another account
will be set as active automatically (if any remain).`,
	Example: `  # Remove an account
  gsuite accounts remove user@gmail.com`,
	Args: cobra.ExactArgs(1),
	RunE: runAccountsRemove,
}

func init() {
	accountsCmd.AddCommand(accountsListCmd)
	accountsCmd.AddCommand(accountsSwitchCmd)
	accountsCmd.AddCommand(accountsRemoveCmd)
	rootCmd.AddCommand(accountsCmd)
}

func runAccountsList(cmd *cobra.Command, args []string) error {
	store, err := auth.LoadAccountStore()
	if err != nil {
		return fmt.Errorf("failed to load account store: %w", err)
	}

	accounts := store.List()
	if len(accounts) == 0 {
		if GetOutputFormat() == "json" {
			return outputJSON([]struct{}{})
		}
		fmt.Println("No authenticated accounts. Run 'gsuite login' to add one.")
		return nil
	}

	if GetOutputFormat() == "json" {
		type accountItem struct {
			Email   string `json:"email"`
			AddedAt string `json:"added_at"`
			Active  bool   `json:"active"`
		}
		var results []accountItem
		for _, entry := range accounts {
			results = append(results, accountItem{
				Email:   entry.Email,
				AddedAt: entry.AddedAt.Format("2006-01-02"),
				Active:  strings.EqualFold(entry.Email, store.Active),
			})
		}
		return outputJSON(results)
	}

	for _, entry := range accounts {
		marker := " "
		if strings.EqualFold(entry.Email, store.Active) {
			marker = "*"
		}
		fmt.Printf("%s %s  (added %s)\n", marker, entry.Email, entry.AddedAt.Format("2006-01-02"))
	}

	return nil
}

func runAccountsSwitch(cmd *cobra.Command, args []string) error {
	email := args[0]

	store, err := auth.LoadAccountStore()
	if err != nil {
		return fmt.Errorf("failed to load account store: %w", err)
	}

	if err := store.SetActive(email); err != nil {
		return err
	}

	if err := store.Save(); err != nil {
		return fmt.Errorf("failed to save account store: %w", err)
	}

	fmt.Printf("Switched to %s\n", email)
	return nil
}

func runAccountsRemove(cmd *cobra.Command, args []string) error {
	email := args[0]

	store, err := auth.LoadAccountStore()
	if err != nil {
		return fmt.Errorf("failed to load account store: %w", err)
	}

	if err := store.RemoveAccount(email); err != nil {
		return err
	}

	if err := store.Save(); err != nil {
		return fmt.Errorf("failed to save account store: %w", err)
	}

	if err := auth.DeleteTokenFor(email); err != nil {
		return fmt.Errorf("failed to remove token: %w", err)
	}

	fmt.Printf("Removed account %s\n", email)
	return nil
}

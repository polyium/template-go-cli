package commands

import (
	"template-go-cli/internal/commands/example"

	"github.com/spf13/cobra"
)

// Execute runs the root command and handles any CLI execution exception. Additionally,
// all child command(s) are added to the root command.
func Execute(root *cobra.Command) {
	examples := &cobra.Group{ID: "examples", Title: "Example Commands"}

	root.AddGroup(examples)

	root.AddCommand(example.Command)

	if e := root.Execute(); e != nil {
		cobra.CheckErr(e)
	}
}

package main

import (
	_ "embed"
	"fmt"
	"log/slog"
	"template-go-cli/internal/constants"
	"template-go-cli/internal/flags/format"

	"template-go-cli/internal/commands"
	"template-go-cli/internal/logging"
	"template-go-cli/internal/types/level"
	"template-go-cli/internal/types/output"

	"github.com/spf13/cobra"
)

var (
	version = "0.0.0"  // See go linking for compile-time variable overwrites.
	commit  = "n/a"    // See go linking for compile-time variable overwrites.
	date    = "latest" // See go linking for compile-time variable overwrites.

	sources = "include" // Include source logging. See go linking for compile-time variable overwrites.
)

// lvl represents the log-level flag set by a persisted global flag.
var lvl level.Type = "info"

// src represents the cli flag to include source logging.
var src bool = true

// the output format if applicable for downstream commands.
var out output.Type = "json"

func main() {
	// The PersistentPreRun and PreRun functions will be executed before Run. PersistentPostRun and PostRun will be executed
	// after Run. The Persistent*Run functions will be inherited by children if they do not declare their own. The *PreRun
	// and *PostRun functions will only be executed if the Run function of the current command has been declared. These
	// functions are run in the following order:
	//
	// - PersistentPreRun
	// - PreRun
	// - Run
	// - PostRun
	// - PersistentPostRun
	//
	// https://github.com/spf13/cobra/blob/main/site/content/user_guide.md
	//

	var root = &cobra.Command{
		Use:                        fmt.Sprintf("%s", constants.Name),
		Short:                      "A cli tool [...]",
		Long:                       "A cli tool [...]",
		Example:                    "",
		Annotations:                map[string]string{},
		Version:                    version,
		SuggestionsMinimumDistance: 3,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Setup slog-specific logging.
			writer := cmd.ErrOrStderr()
			addsource := src && sources == "include"
			options := &slog.HandlerOptions{AddSource: addsource, Level: lvl.Level(), ReplaceAttr: logging.Replacements}
			handler := slog.NewTextHandler(writer, options)
			logger := slog.New(handler)

			log := logger.With(slog.String("command", cmd.Name()))

			slog.SetDefault(log)

			// Propagate updated logger into context.
			ctx = logging.With(ctx, log)
			cmd.SetContext(ctx)

			slog.Log(ctx, level.Trace.Level(), "Starting Application", slog.String("version", version), slog.String("commit", commit), slog.String("date", date))

			// Propagate persistent flags into context for easy retrieval and strict typing.

			ctx = format.With(ctx, out)
			cmd.SetContext(ctx)

			return nil
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   false,
			DisableNoDescFlag:   false,
			DisableDescriptions: false,
			HiddenDefaultCmd:    false,
		},
		TraverseChildren: true,
	}

	root.PersistentFlags().VarP(&lvl, "log-level", "z", "log-level verbosity")
	root.PersistentFlags().BoolVarP(&src, "include-source-locations", "x", true, "include log locations")
	root.PersistentFlags().VarP(&out, "output", "o", "command output format; not applicable to all commands")

	commands.Execute(root)
}

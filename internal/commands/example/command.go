package example

import (
    "fmt"
    "log/slog"
    "os"
    "polyforge/internal/constants"
    "polyforge/internal/flags/format"
    "polyforge/internal/logging"
    "polyforge/internal/types/level"
    "strings"

    "polyforge/internal/types/output"

    "github.com/spf13/cobra"
)

var (
    name string
)

var Command = &cobra.Command{
    Use:        "example",
    Aliases:    []string{},
    SuggestFor: nil,
    GroupID:    "examples",
    Short:      "An example command and template",
    Long:       "The example's command long-description -- value should be in full sentences, and can span multiple lines.",
    Example: strings.Join([]string{
        fmt.Sprintf("  %s", "# General command usage"),
        fmt.Sprintf("  %s", fmt.Sprintf("%s example --name \"test-value\"", constants.Name)),
        "",
        fmt.Sprintf("  %s", "# Extended usage demonstrating configuration of default(s)"),
        fmt.Sprintf("  %s", fmt.Sprintf("%s example --name \"test-value\" --output json", constants.Name)),
        "",
        fmt.Sprintf("  %s", "# Display help information and command usage"),
        fmt.Sprintf("  %s", fmt.Sprintf("%s example --help", constants.Name)),
    }, "\n"),
    PreRunE: func(cmd *cobra.Command, args []string) error {
        ctx := cmd.Context()

        // Update logger instance to include command's flags.
        var logger = logging.Get(ctx)
        log := logger.With(slog.Any("flags", cmd.Flags()))
        slog.SetDefault(log)

        ctx = logging.With(ctx, log)

        cmd.SetContext(ctx)

        return nil
    },
    RunE: func(cmd *cobra.Command, args []string) error {
        ctx := cmd.Context()

        var log = logging.Get(ctx)

        log.Log(ctx, level.Trace.Level(), "Running example command")

        var datum = map[string]string{
            "name": name,
        }

        buffer, e := output.Write(format.Get(ctx), datum)
        if e != nil {
            return e
        }

        fmt.Fprintf(os.Stdout, "%s", buffer.String())

        return nil
    },
    TraverseChildren: true,
    SilenceErrors:    true,
}

func init() {
    flags := Command.Flags()

    flags.StringVarP(&name, "name", "n", "", "a required example named-string-flag")
    if e := Command.MarkFlagRequired("name"); e != nil {
        if exception := Command.Help(); exception != nil {
            panic(exception)
        }
    }
}

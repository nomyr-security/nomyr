package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"nomyr/internal/apperror"
	"nomyr/internal/buildinfo"
	"nomyr/internal/demo"
	platformserver "nomyr/internal/platform/server"
	"nomyr/internal/scaffold"
	"nomyr/internal/usagemetrics"
	"nomyr/internal/webui"
)

type cliOptions struct {
	output  string
	noInput bool
}

func main() {
	root := newRootCommand()
	startedAt := time.Now()
	executed, executeError := root.ExecuteC()
	commandPath := root.CommandPath()
	if executed != nil {
		commandPath = executed.CommandPath()
	}
	metricContext, cancelMetrics := context.WithTimeout(context.Background(), 800*time.Millisecond)
	usagemetrics.RecordCommand(metricContext, commandPath, time.Since(startedAt), executeError == nil)
	cancelMetrics()
	if executeError != nil {
		_ = renderExecutionError(os.Stderr, rootOptions(root).output, executeError)
		os.Exit(1)
	}
}

func rootOptions(root *cobra.Command) *cliOptions {
	output, _ := root.PersistentFlags().GetString("output")
	noInput, _ := root.PersistentFlags().GetBool("no-input")
	return &cliOptions{output: output, noInput: noInput}
}

func newRootCommand() *cobra.Command {
	options := &cliOptions{}
	root := &cobra.Command{
		Use:           "nomyr",
		Short:         "Nomyr non-human identity security platform",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if options.output != "text" && options.output != "json" && options.output != "yaml" {
				return apperror.InvalidArgument("output must be one of text, json, or yaml")
			}
			return nil
		},
	}
	root.PersistentFlags().StringVar(&options.output, "output", "text", "output format: text, json, or yaml")
	root.PersistentFlags().BoolVar(&options.noInput, "no-input", false, "disable interactive input")
	root.AddCommand(newVersionCommand(options), newDoctorCommand(options), newLoginCommand(), newContextCommand(options), newDemoCommand(options), newTelemetryCommand(options), newCompletionCommand())
	root.AddCommand(newScaffoldCommand(options))
	return root
}

func newCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "completion SHELL",
		Short:     "Generate shell completion scripts",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(command *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return command.Root().GenBashCompletionV2(command.OutOrStdout(), true)
			case "zsh":
				return command.Root().GenZshCompletion(command.OutOrStdout())
			case "fish":
				return command.Root().GenFishCompletion(command.OutOrStdout(), true)
			case "powershell":
				return command.Root().GenPowerShellCompletionWithDesc(command.OutOrStdout())
			default:
				return apperror.InvalidArgument("completion shell must be bash, zsh, fish, or powershell")
			}
		},
	}
}

func newScaffoldCommand(options *cliOptions) *cobra.Command {
	command := &cobra.Command{Use: "scaffold", Short: "Create non-functional starter files"}
	for _, kind := range []string{"connector", "rule"} {
		kind := kind
		kindCommand := &cobra.Command{Use: kind, Short: "Manage " + kind + " scaffolds"}
		child := &cobra.Command{
			Use:   "new NAME",
			Short: "Create a structural " + kind + " starter",
			Args:  cobra.ExactArgs(1),
			RunE: func(command *cobra.Command, args []string) error {
				directory, _ := command.Flags().GetString("directory")
				created, err := scaffold.Create(directory, kind, args[0])
				if err != nil {
					return err
				}
				if options.output == "text" {
					_, err = fmt.Fprintf(command.OutOrStdout(), "Created structural %s scaffold at %s (no provider support or execution is implemented).\n", kind, created)
					return err
				}
				return writeOutput(command.OutOrStdout(), options.output, map[string]string{"path": created, "kind": kind, "status": "scaffold-only", "provider_support": "not-implemented"})
			},
		}
		child.Flags().String("directory", ".", "parent directory for the new scaffold")
		kindCommand.AddCommand(child)
		command.AddCommand(kindCommand)
	}
	return command
}

func newTelemetryCommand(options *cliOptions) *cobra.Command {
	command := &cobra.Command{Use: "telemetry", Short: "Manage anonymous usage metrics"}
	command.AddCommand(
		&cobra.Command{Use: "enable", Short: "Enable anonymous usage metrics", RunE: func(command *cobra.Command, _ []string) error {
			if _, err := usagemetrics.Enable(); err != nil {
				return err
			}
			return emit(command, options, map[string]string{"status": "enabled", "notice": "Anonymous usage metrics are enabled."}, "Anonymous usage metrics enabled. See https://github.com/nomyr-security/nomyr/blob/main/TELEMETRY.md\n")
		}},
		&cobra.Command{Use: "disable", Short: "Disable anonymous usage metrics", RunE: func(command *cobra.Command, _ []string) error {
			if err := usagemetrics.Disable(); err != nil {
				return err
			}
			return emit(command, options, map[string]string{"status": "disabled", "notice": "The local installation ID was removed."}, "Anonymous usage metrics disabled and the local installation ID removed.\n")
		}},
		&cobra.Command{Use: "status", Short: "Show anonymous usage metrics status", RunE: func(command *cobra.Command, _ []string) error {
			config, err := usagemetrics.Status()
			if err != nil {
				return err
			}
			status := "disabled"
			if config.Enabled {
				status = "enabled"
			}
			return emit(command, options, map[string]string{"status": status}, "Anonymous usage metrics: "+status+"\n")
		}},
	)
	return command
}

func newVersionCommand(options *cliOptions) *cobra.Command {
	var legacyJSON bool
	command := &cobra.Command{Use: "version", Short: "Print build version", RunE: func(command *cobra.Command, _ []string) error {
		info := buildinfo.Current()
		if legacyJSON {
			return writeOutput(command.OutOrStdout(), "json", info)
		}
		return emit(command, options, info, fmt.Sprintf("nomyr %s (%s, %s)\n", info.Version, info.Commit, info.Date))
	}}
	command.Flags().BoolVar(&legacyJSON, "json", false, "emit JSON")
	return command
}

func newDoctorCommand(options *cliOptions) *cobra.Command {
	var legacyJSON bool
	command := &cobra.Command{Use: "doctor", Short: "Check the local installation", RunE: func(command *cobra.Command, _ []string) error {
		report := map[string]any{
			"status": "ok", "mode": "local-demo", "platform": runtime.GOOS + "/" + runtime.GOARCH,
			"checks": []map[string]string{{"name": "embedded-web-assets", "status": "available"}, {"name": "product-services", "status": "not-implemented"}},
		}
		if legacyJSON {
			return writeOutput(command.OutOrStdout(), "json", report)
		}
		return emit(command, options, report, "Nomyr doctor: local demo is healthy\n  embedded web assets  available\n  product services      not implemented\n")
	}}
	command.Flags().BoolVar(&legacyJSON, "json", false, "emit JSON")
	return command
}

func newLoginCommand() *cobra.Command {
	return &cobra.Command{Use: "login", Short: "Authenticate with a Nomyr server", RunE: func(_ *cobra.Command, _ []string) error {
		return apperror.NotImplemented("device-code login")
	}}
}

func newContextCommand(options *cliOptions) *cobra.Command {
	return &cobra.Command{Use: "context", Short: "Show the active context", RunE: func(command *cobra.Command, _ []string) error {
		return emit(command, options, map[string]string{"context": "local (demo)"}, "local (demo)\n")
	}}
}

func newDemoCommand(options *cliOptions) *cobra.Command {
	listenAddress := "127.0.0.1:8080"
	command := &cobra.Command{Use: "demo", Short: "Serve a local demo with synthetic data", RunE: func(command *cobra.Command, _ []string) error {
		if !isLoopbackAddress(listenAddress) {
			return &apperror.Error{Code: "NHI-CONFIG-0001", What: "demo mode cannot bind to a non-loopback address", Cause: "the local demo does not provide TLS or authentication", Fix: "use 127.0.0.1 or localhost until secure exposure is implemented", Docs: "https://github.com/nomyr-security/nomyr#local-development"}
		}
		server := &http.Server{Addr: listenAddress, Handler: platformserver.Handler(webui.Public(), demo.Seed()), ReadHeaderTimeout: 5 * time.Second}
		if err := emit(command, options, map[string]string{"status": "listening", "address": "http://" + listenAddress, "watermark": demo.Seed().Watermark}, fmt.Sprintf("Nomyr demo shell listening on http://%s (demo data)\n", listenAddress)); err != nil {
			return err
		}
		if err := server.ListenAndServe(); err != nil {
			return apperror.OperationFailed("start local demo server")
		}
		return nil
	}}
	command.Flags().StringVar(&listenAddress, "listen", listenAddress, "loopback listen address")
	return command
}

func emit(command *cobra.Command, options *cliOptions, value any, text string) error {
	if options.output == "text" {
		_, err := fmt.Fprint(command.OutOrStdout(), text)
		return err
	}
	return writeOutput(command.OutOrStdout(), options.output, value)
}

func writeOutput(writer interface{ Write([]byte) (int, error) }, format string, value any) error {
	switch format {
	case "json":
		return json.NewEncoder(writer).Encode(value)
	case "yaml":
		return yaml.NewEncoder(writer).Encode(value)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func renderExecutionError(writer interface{ Write([]byte) (int, error) }, format string, err error) error {
	var contractError *apperror.Error
	if errors.As(err, &contractError) {
		if format == "json" || format == "yaml" {
			return writeOutput(writer, format, contractError)
		}
		_, writeError := fmt.Fprintln(writer, contractError)
		return writeError
	}
	_, writeError := fmt.Fprintln(writer, err)
	return writeError
}

func isLoopbackAddress(address string) bool {
	host, _, splitError := net.SplitHostPort(address)
	if splitError != nil {
		return false
	}
	if host == "localhost" {
		return true
	}
	parsed := net.ParseIP(host)
	return parsed != nil && parsed.IsLoopback()
}

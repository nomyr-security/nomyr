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

	"nomyr/internal/apperror"
	"nomyr/internal/buildinfo"
	"nomyr/internal/demo"
	platformserver "nomyr/internal/platform/server"
	"nomyr/internal/usagemetrics"
	"nomyr/internal/webui"
)

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
		var contractError *apperror.Error
		if errors.As(executeError, &contractError) {
			fmt.Fprintln(os.Stderr, contractError)
		} else {
			fmt.Fprintln(os.Stderr, executeError)
		}
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "nomyr",
		Short:         "Nomyr non-human identity security platform",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(newVersionCommand(), newDoctorCommand(), newLoginCommand(), newContextCommand(), newDemoCommand(), newTelemetryCommand())
	return root
}

func newTelemetryCommand() *cobra.Command {
	command := &cobra.Command{Use: "telemetry", Short: "Manage anonymous usage metrics"}
	command.AddCommand(
		&cobra.Command{Use: "enable", Short: "Enable anonymous usage metrics", RunE: func(command *cobra.Command, _ []string) error {
			if _, err := usagemetrics.Enable(); err != nil {
				return err
			}
			fmt.Fprintln(command.OutOrStdout(), "Anonymous usage metrics enabled. See https://github.com/nomyr-security/nomyr/blob/main/TELEMETRY.md")
			return nil
		}},
		&cobra.Command{Use: "disable", Short: "Disable anonymous usage metrics", RunE: func(command *cobra.Command, _ []string) error {
			if err := usagemetrics.Disable(); err != nil {
				return err
			}
			fmt.Fprintln(command.OutOrStdout(), "Anonymous usage metrics disabled and the local installation ID removed.")
			return nil
		}},
		&cobra.Command{Use: "status", Short: "Show anonymous usage metrics status", RunE: func(command *cobra.Command, _ []string) error {
			config, err := usagemetrics.Status()
			if err != nil {
				return err
			}
			if config.Enabled {
				fmt.Fprintln(command.OutOrStdout(), "Anonymous usage metrics: enabled")
			} else {
				fmt.Fprintln(command.OutOrStdout(), "Anonymous usage metrics: disabled")
			}
			return nil
		}},
	)
	return command
}

func newVersionCommand() *cobra.Command {
	var outputJSON bool
	command := &cobra.Command{Use: "version", Short: "Print build version", RunE: func(command *cobra.Command, _ []string) error {
		info := buildinfo.Current()
		if outputJSON {
			return json.NewEncoder(command.OutOrStdout()).Encode(info)
		}
		fmt.Fprintf(command.OutOrStdout(), "nomyr %s (%s, %s)\n", info.Version, info.Commit, info.Date)
		return nil
	}}
	command.Flags().BoolVar(&outputJSON, "json", false, "emit JSON")
	return command
}

func newDoctorCommand() *cobra.Command {
	var outputJSON bool
	command := &cobra.Command{Use: "doctor", Short: "Check the local installation", RunE: func(command *cobra.Command, _ []string) error {
		report := map[string]any{
			"status":   "ok",
			"mode":     "local-demo",
			"platform": runtime.GOOS + "/" + runtime.GOARCH,
			"checks":   []map[string]string{{"name": "embedded-web-assets", "status": "available"}},
		}
		if outputJSON {
			return json.NewEncoder(command.OutOrStdout()).Encode(report)
		}
		fmt.Fprintln(command.OutOrStdout(), "Nomyr doctor: local demo is healthy")
		fmt.Fprintln(command.OutOrStdout(), "  embedded web assets  available")
		fmt.Fprintln(command.OutOrStdout(), "  product services      not implemented")
		return nil
	}}
	command.Flags().BoolVar(&outputJSON, "json", false, "emit JSON")
	return command
}

func newLoginCommand() *cobra.Command {
	return &cobra.Command{Use: "login", Short: "Authenticate with a Nomyr server", RunE: func(_ *cobra.Command, _ []string) error {
		return apperror.NotImplemented("device-code login")
	}}
}

func newContextCommand() *cobra.Command {
	return &cobra.Command{Use: "context", Short: "Show the active context", RunE: func(command *cobra.Command, _ []string) error {
		fmt.Fprintln(command.OutOrStdout(), "local (demo)")
		return nil
	}}
}

func newDemoCommand() *cobra.Command {
	listenAddress := "127.0.0.1:8080"
	command := &cobra.Command{Use: "demo", Short: "Serve a local demo with synthetic data", RunE: func(command *cobra.Command, _ []string) error {
		if !isLoopbackAddress(listenAddress) {
			return &apperror.Error{
				Code:  "NHI-CONFIG-0001",
				What:  "demo mode cannot bind to a non-loopback address",
				Cause: "the local demo does not provide TLS or authentication",
				Fix:   "use 127.0.0.1 or localhost until secure exposure is implemented",
				Docs:  "https://github.com/nomyr-security/nomyr#local-development",
			}
		}
		server := &http.Server{
			Addr:              listenAddress,
			Handler:           platformserver.Handler(webui.Public(), demo.Seed()),
			ReadHeaderTimeout: 5 * time.Second,
		}
		fmt.Fprintf(command.OutOrStdout(), "Nomyr demo shell listening on http://%s (demo data)\n", listenAddress)
		return server.ListenAndServe()
	}}
	command.Flags().StringVar(&listenAddress, "listen", listenAddress, "loopback listen address")
	return command
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

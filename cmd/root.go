// cmd/root.go
package cmd

import (
	"fmt"

	"github.com/lehigh-university-libraries/mountain-hawk/internal/config"
	"github.com/lehigh-university-libraries/mountain-hawk/internal/server"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	daemon  bool
	port    string
	verbose bool
)

// NewRootCommand creates the root command
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mountain-hawk",
		Short: "GitHub PR reviewer using MCP and LLM",
		Long: `A comprehensive pull request reviewer that uses the Model Context Protocol (MCP) 
to interact with GitHub and leverages AI models to provide intelligent code reviews.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if daemon {
				return runDaemon()
			}
			return cmd.Help()
		},
	}

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&daemon, "daemon", "d", false, "Run as webhook server daemon")
	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "", "Port for daemon mode")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	rootCmd.AddCommand(NewReviewCommand())

	return rootCmd
}

func runDaemon() error {
	cfg := config.MustLoad()

	// Override port if specified
	if port != "" {
		cfg.Port = port
	}

	fmt.Printf("Starting MCP webhook server on port %s...\n", cfg.Port)
	srv := server.New(cfg)
	return srv.ListenAndServe()
}

// GetVerbose returns the verbose flag value for use in subcommands
func GetVerbose() bool {
	return verbose
}

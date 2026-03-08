package main

import (
	"fmt"
	"os"

	"github.com/da3az/lazado/internal/app"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:     "lazado",
		Short:   "Azure DevOps TUI — manage work items, PRs, pipelines, and repos",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.Run()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Set up lazado (configure org, project, and authentication)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.RunInit()
		},
	}

	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

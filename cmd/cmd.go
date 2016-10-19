package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kraken",
	Short: "A utility for load testing applications",
	Long: `
Kraken is a utility for load testing applications on the host system.
It's a work in-progress, expect issues and you won't be disappointed.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main().  It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.AddCommand(attackCmd())
	rootCmd.AddCommand(pummelCmd())
	rootCmd.AddCommand(siegeCmd())
	rootCmd.AddCommand(reportCmd())
	// rootCmd.AddCommand(NewVersionCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

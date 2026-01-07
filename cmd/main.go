package main

import (
	"fmt"
	"os"

	start "github.com/maxcelant/sinkplot/internal/commands"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{Use: "sinkctl"}

func init() {
	startCmd := start.NewCommand()
	startCmd.Flags().String("path", "Sinkfile", "path to the initial Sinkfile config")
	rootCmd.AddCommand(startCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

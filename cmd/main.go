package main

import (
	"fmt"
	"os"

	start "github.com/maxcelant/sinkplot/internal/commands"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{Use: "sinkctl"}

func init() {
	rootCmd.AddCommand(start.NewCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

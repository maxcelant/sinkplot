package start

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start sinkplot proxy in the foreground",
		Long:  "I'll fill this in later :)",
		Run:   runStart,
	}
}

func runStart(cmd *cobra.Command, args []string) {
	fmt.Println("hello world")
}

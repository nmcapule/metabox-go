package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "metabox [restore|backup]",
	Short: "VCS-friendly backup/restore tool",
	Args:  cobra.MinimumNArgs(1),
}

func Execute() {
	if err := root.Execute(); err != nil {
		log.Fatalln(err)
	}
}

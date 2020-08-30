package cmd

import (
	"fmt"
	"log"

	"github.com/nmcapule/metabox-go/config"
	"github.com/nmcapule/metabox-go/metabox"
	"github.com/spf13/cobra"
)

type Backup struct {
	filename string
	flagTags *[]string
}

func (cmd *Backup) Execute() error {
	cfg, err := config.FromFile(cmd.filename)
	if err != nil {
		return fmt.Errorf("get config: %v", err)
	}

	// Attach flagTags if exists.
	cfg.Workspace.TagsGenerator = append(cfg.Workspace.TagsGenerator, (*cmd.flagTags)...)

	box, err := metabox.New(cfg)
	if err != nil {
		return fmt.Errorf("metabox from config: %v", err)
	}

	_, err = box.StartBackup()
	if err != nil {
		return fmt.Errorf("backup: %v", err)
	}

	return nil
}

var cmdBackup = &cobra.Command{
	Use:   "backup",
	Short: "Creates a backup record",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := Backup{
			filename: args[0],
			flagTags: cmd.Flags().StringArrayP("tags", "t", nil, "Tag matchers"),
		}
		if err := r.Execute(); err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	root.AddCommand(cmdBackup)
}

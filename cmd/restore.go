package cmd

import (
	"fmt"
	"log"

	"github.com/nmcapule/metabox-go/metabox"
	"github.com/nmcapule/metabox-go/tracker"
	"github.com/spf13/cobra"
)

type Restore struct {
	configPath string
	flagTags   []string
}

func (cmd *Restore) Execute() error {
	box, err := metabox.FromConfigFile(cmd.configPath)
	if err != nil {
		return fmt.Errorf("metabox from config: %v", err)
	}

	var matchers []tracker.Predicate
	for _, tag := range cmd.flagTags {
		matchers = append(matchers, tracker.PredicateTag(tag))
	}

	item, err := box.DB.QueryLatest(matchers...)
	if err != nil {
		return fmt.Errorf("retrieving item tagged %+v: %v", cmd.flagTags, err)
	}

	return box.StartRestore(item)
}

func init() {
	cmdRestore := &cobra.Command{
		Use:   "restore",
		Short: "Restore file from backup",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			tags, err := cmd.Flags().GetStringArray("tags")
			if err != nil {
				log.Fatalln(err)
			}

			r := Restore{
				configPath: args[0],
				flagTags:   tags,
			}
			if err := r.Execute(); err != nil {
				log.Fatalln(err)
			}
		},
	}
	cmdRestore.Flags().StringArrayP("tags", "t", nil, "Tag matchers")

	root.AddCommand(cmdRestore)
}

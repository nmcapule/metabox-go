package cmd

import (
	"fmt"
	"log"

	"github.com/nmcapule/metabox-go/metabox"
	"github.com/nmcapule/metabox-go/tracker"
	"github.com/spf13/cobra"
)

type Restore struct {
	filename string
	flagTags *[]string
}

func (cmd *Restore) Execute() error {
	box, err := metabox.FromConfigFile(cmd.filename)
	if err != nil {
		return fmt.Errorf("metabox from config: %v", err)
	}

	var matchers []tracker.Predicate
	for _, tag := range *cmd.flagTags {
		matchers = append(matchers, tracker.PredicateTag(tag))
	}

	item, err := box.DB.QueryLatest(matchers...)
	if err != nil {
		return fmt.Errorf("retrieving item tagged %+v: %v", *cmd.flagTags, err)
	}

	return box.StartRestore(item)
}

var cmdRestore = &cobra.Command{
	Use:   "restore",
	Short: "Restore file from backup",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := Restore{
			filename: args[0],
			flagTags: cmd.Flags().StringArrayP("tags", "t", nil, "Tag matchers"),
		}
		if err := r.Execute(); err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	root.AddCommand(cmdRestore)
}

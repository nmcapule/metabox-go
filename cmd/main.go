// Package main implements a simple backup / restore tool.
//
// Example (backup and tag with "mytag"):
//   ./metabox -f examples/default.metabox.yml -t mytag restore
//
// Example (restore tagged with "mytag"):
//   ./metabox -f examples/default.metabox.yml -t mytag backup
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/nmcapule/metabox-go/config"
	"github.com/nmcapule/metabox-go/metabox"
	"github.com/nmcapule/metabox-go/tracker"
)

var (
	flagConfigPath = flag.String("config", "default.metabox.yml", "Config file path")
	flagUseTag     = flag.String("tag", "", "Use tag in backup or restore")
)

func backup(box *metabox.Metabox) error {
	item, err := box.StartBackup()
	if err != nil {
		return err
	}

	log.Printf("Created item: %+v", item)
	return nil
}

func restore(box *metabox.Metabox, tag string) error {
	cond := tracker.PredicateAll
	if tag != "" {
		cond = tracker.PredicateTag(tag)
	}

	item, err := box.DB.QueryLatest(cond)
	if err != nil {
		return fmt.Errorf("retrieving item tagged %q: %v", tag, err)
	}

	return box.StartRestore(item)
}

func main() {
	flag.Parse()

	cmd := flag.Arg(0)
	if cmd == "" {
		log.Fatalln("Required argument <command> is empty!")
	}

	cfg, err := config.FromFile(*flagConfigPath)
	if err != nil {
		log.Fatalf("Reading config from %q: %v", *flagConfigPath, err)
	}
	if *flagUseTag != "" {
		cfg.Workspace.TagsGenerator = append(cfg.Workspace.TagsGenerator, *flagUseTag)
	}

	box, err := metabox.New(cfg)
	if err != nil {
		log.Fatalln("Create metabox:", err)
	}

	switch cmd {
	case "backup":
		if err := backup(box); err != nil {
			log.Fatalln("Metabox backup:", err)
		}
	case "restore":
		if err := restore(box, *flagUseTag); err != nil {
			log.Fatalln("Metabox restore:", err)
		}
	default:
		log.Fatalf("Unknown command %q", cmd)
	}
}

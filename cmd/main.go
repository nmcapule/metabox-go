package main

import (
	"flag"
	"log"

	"github.com/nmcapule/metabox-go/metabox"
)

func main() {
	flag.Parse()

	cfgpath := flag.Arg(0)
	if cfgpath == "" {
		log.Fatalln("Required argument <config-path> is empty!")
	}

	box, err := metabox.FromConfigFile(cfgpath)
	if err != nil {
		log.Fatalln(err)
	}

	if err := box.Start(); err != nil {
		log.Fatalln(err)
	}
}

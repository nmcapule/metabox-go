package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/nmcapule/metabox-go/config"
)

func main() {
	flag.Parse()

	cfgpath := flag.Arg(0)
	if cfgpath == "" {
		log.Fatalln("Required argument <config-path> is empty!")
	}

	cfg, err := config.FromFile(cfgpath)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%+v", cfg)
}

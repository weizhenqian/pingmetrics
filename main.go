package main

import (
		"flag"
		"os"
		"fmt"
		"./g"
		"./funcs"
)

var (
	vers *bool
	help *bool
	cfg  *string
)

func init() {
	cfg = flag.String("c", "cfg.json", "configuration file")
	vers = flag.Bool("v", false, "display the version.")
	help = flag.Bool("h", false, "print this help.")
	flag.Parse()
	if *vers {
		fmt.Println("Version:", version)
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}
}

func main() {
	g.ParseConfig(*cfg)
	UpdateSSNetState()
	PingMetrics()
}
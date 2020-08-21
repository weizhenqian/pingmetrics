package main

import (
	"flag"
	"os"
	"./g"
	"./funcs"
)

var (
	cfg *string
	help *bool

)

func init() {
	cfg  = flag.String("c", "cfg.json", "configuration file")
	help = flag.Bool("h", false, "print this help.")
	flag.Parse()


	if *help {
		flag.Usage()
		os.Exit(0)
	}
}

func main() {
	g.ParseConfig(*cfg)
	funcs.UpdateSSNetState()
	funcs.PingMetrics()
}

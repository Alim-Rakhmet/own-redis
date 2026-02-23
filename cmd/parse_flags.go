package cmd

import (
	"flag"
	"fmt"
	"os"
)

func parseFlags() string {
	fs := flag.NewFlagSet("own-redis", flag.ContinueOnError)

	port := fs.String("port", "8080", "server port")

	fs.Usage = printUsage

	for _, arg := range os.Args {
		if arg == "--help" || arg == "-help" {
			printUsage()
			os.Exit(0)
		}
	}

	err := fs.Parse(os.Args[1:])
	if err != nil {
		os.Exit(0)
	}

	return *port
}

func printUsage() {
	fmt.Print("Own Redis\n\nUsage:  own-redis [--port <N>]  own-redis --help\n\nOptions:\n  --help       Show this screen.\n  --port N     Port number.\n")
}

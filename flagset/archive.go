package main

import (
	"flag"
	"log"
)

type ArchiveCmdConfig struct {
	Format string
	List   bool

	Verbose bool
}

func runArchive(conf Config, args []string) error {
	fs := flag.NewFlagSet("archive", flag.ExitOnError)

	var cmdConf ArchiveCmdConfig
	fs.StringVar(&cmdConf.Format, "format", "", "Format of the resulting archive.")
	fs.BoolVar(&cmdConf.List, "list", false, "Show all available formats.")
	fs.BoolVar(&cmdConf.Verbose, "verbose", false, "Be verbose.")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if cmdConf.Verbose {
		log.Printf("(verbose mode) run archive with config %+v, cmd config %+v\n", conf, cmdConf)
	}

	if cmdConf.List {
		log.Printf("available formats: all")
		return nil
	}

	log.Printf("archive, format %q, command args %v\n", cmdConf.Format, fs.Args())

	return nil
}

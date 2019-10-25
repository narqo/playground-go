package main

import (
	"flag"
	"log"
)

type AddCmdConfig struct {
	DryRun      bool
	Interactive bool

	Verbose bool
}

func runAdd(conf Config, args []string) error {
	fs := flag.NewFlagSet("add", flag.ExitOnError)

	var cmdConf AddCmdConfig
	fs.BoolVar(&cmdConf.DryRun, "dry-run", false, "Don't actually do a thing.")
	fs.BoolVar(&cmdConf.Interactive, "interactive", false, "Run in \"Interactive mode\".")
	fs.BoolVar(&cmdConf.Verbose, "verbose", false, "Be verbose.")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if cmdConf.Verbose {
		log.Printf("(verbose mode) run add with app config %+v, cmd config %+v\n", conf, cmdConf)
	}

	if cmdConf.DryRun {
		log.Println("(dry run)")
	}

	if cmdConf.Interactive {
		log.Println("Hi! I'm Clippy, your office assistant!. Looks like you are trying to make something useful with \"flag\" package.")
	}

	log.Printf("add, command args %v\n", fs.Args())

	return nil
}

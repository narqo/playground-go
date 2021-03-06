package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	WorkTree  string
	Namespace string

	Branch []string
}

func main() {
	version := flag.Bool("version", false, "Print version and exit.")

	var conf Config
	flag.StringVar(&conf.WorkTree, "work-tree", "", "Set the path to the working tree.")
	flag.StringVar(&conf.Namespace, "namespace", "", "Set namespace.")
	flag.Var((*stringsValue)(&conf.Branch), "branch", "Set branch or branches.")

	flag.Usage = usage

	flag.Parse()

	if *version {
		runVersion()
		os.Exit(2)
	}

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	var err error
	args := flag.Args()
	cmd, args := args[0], args[1:]
	switch cmd {
	case "add":
		err = runAdd(conf, args)
	case "archive":
		err = runArchive(conf, args)
	default:
		fmt.Fprintf(flag.CommandLine.Output(), "unknown command %s\n", cmd)
		flag.Usage()
		os.Exit(2)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func runVersion() {
	log.Println("version 0.0.1-1+build.1")
}

func usage() {
	w := flag.CommandLine.Output()

	fmt.Fprintf(w, "Usage: %s [options] <commands> <args>\n\n", os.Args[0])

	fmt.Fprintln(w, "Options")
	flag.PrintDefaults()

	fmt.Fprintln(w, "\nCommands")
	fmt.Fprintln(w, "  add\n\tAdd command.")
	fmt.Fprintln(w, "  archive\n\tArchive command.")
}

type stringsValue []string

func (ss stringsValue) Get() interface{} {
	return ss
}

func (ss *stringsValue) Set(s string) error {
	for _, v := range strings.Split(s, ",") {
		*ss = append(*ss, strings.TrimSpace(v))
	}
	return nil
}

func (ss stringsValue) String() string {
	if ss == nil {
		return ""
	}
	return strings.Join(ss, ",")
}

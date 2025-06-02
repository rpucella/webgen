package main

import (
	"fmt"
	"log"
	"os"
	"rpucella.net/webgen/internal/gen"
	"strings"
)

var rep *log.Logger = log.New(os.Stdout, "" /* log.Ldate| */, log.Ltime)

type flags struct {
	draft bool
	help  bool
}

func main() {

	args, flags := ClassifyArgs(os.Args[1:])

	if flags.help {
		Usage()
	} else if len(args) == 0 {
		gen.WalkAndProcessPosts(".")
		gen.WalkAndProcessMarkdowns(".")
		gen.WalkAndProcessContents(".")
	} else if len(args) == 1 {
		fi, err := os.Stat(args[0])
		if err != nil {
			rep.Fatalf("ERROR: %s\n", err)
		}
		if fi.IsDir() {
			gen.WalkAndProcessPosts(args[0])
			gen.WalkAndProcessMarkdowns(args[0])
			gen.WalkAndProcessContents(args[0])
		} else if gen.IsContent(args[0]) {
			if err := gen.ProcessFileContent(os.Stdout, args[0]); err != nil {
				rep.Fatal(fmt.Sprintf("ERROR: %s\n", err))
			}
		} else if gen.IsMarkdown(args[0]) {
			if flags.draft {
				if err := gen.ProcessFileMarkdownDraft(args[0]); err != nil {
					rep.Fatal(fmt.Sprintf("ERROR: %s\n", err))
				}
			} else {
				if err := gen.ProcessFileMarkdown(os.Stdout, args[0]); err != nil {
					rep.Fatal(fmt.Sprintf("ERROR: %s\n", err))
				}
			}
		} else {
			rep.Fatal(fmt.Sprintf("ERROR: unknown file extension %s\n", args[0]))
		}
	} else {
		Usage()
	}
}

func Usage() {
	rep.Println("USAGE: webgen [--help] [<folder> | <file.content>]")
}

func ClassifyArgs(args []string) ([]string, flags) {
	rArgs := make([]string, 0, len(args))
	flags := flags{}
	for _, arg := range args {
		if arg == "--help" {
			flags.help = true
		} else if strings.HasPrefix(arg, "--") {
			rep.Println(fmt.Sprintf("Unknown flag: %s", strings.TrimPrefix(arg, "--")))
		} else {
			rArgs = append(rArgs, arg)
		}
	}
	return rArgs, flags
}

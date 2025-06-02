package main

import (
	"fmt"
	"log"
	"os"
	"rpucella.net/webgen/internal/gen"
)

var rep *log.Logger = log.New(os.Stdout, "" /* log.Ldate| */, log.Ltime)

func main() {

	args := os.Args[1:]
	if len(args) < 1 {
		Usage()
		return
	}
	command := args[0]
	switch command {
	case "draft":
		if len(args) != 2 {
			Usage()
			return
		}
		err := gen.ProcessFileMarkdownDraft(args[1])
		if err != nil {
			rep.Fatal(fmt.Sprintf("ERROR: %s\n", err))
		}

	default:
		rep.Fatalf("ERROR: Unknown command %s\n", command)
	}
}

func Usage() {
	rep.Println("USAGE: weblog <command> <arg>...")
	rep.Println("  commands := ")
	rep.Println("    draft <file.md>")
}

package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const TEMPLATE = "__content.template"
const SUBTEMPLATE = "__sub.template"
const MDTEMPLATE = "__markdown.template"
const SUMMARYTEMPLATE = "__summary.template"
const GENDIR = "__src"
const GENPOSTS = "__posts"
const POSTMD = "index.md"

var rep *log.Logger = log.New(os.Stdout, "" /* log.Ldate| */, log.Ltime)

type Content struct {
	Title string
	Date  string
	Key   string
	Body  template.HTML
}

func main() {

	args := os.Args[1:]

	if len(args) == 0 {
		WalkAndProcessPosts(".")
		WalkAndProcessMarkdowns(".")
		WalkAndProcessContents(".")
	} else if len(args) == 1 {
		fi, err := os.Stat(args[0])
		if err != nil {
			rep.Fatalf("ERROR: %s\n", err)
		}
		if fi.IsDir() {
			WalkAndProcessPosts(args[0])
			WalkAndProcessMarkdowns(args[0])
			WalkAndProcessContents(args[0])
		} else if isContent(args[0]) {
			if err := ProcessFileContent(os.Stdout, args[0]); err != nil {
				rep.Fatal(fmt.Sprintf("ERROR: %s\n", err))
			}
		} else if isMarkdown(args[0]) {
			if err := ProcessFileMarkdown(os.Stdout, args[0]); err != nil {
				rep.Fatal(fmt.Sprintf("ERROR: %s\n", err))
			}
		} else {
			rep.Fatal(fmt.Sprintf("ERROR: unknown file extension %s\n", args[0]))
		}
	} else {
		Usage()
	}
}

func Usage() {
	rep.Println("USAGE: webgen [<folder> | <file.content> | <file.md>]")
}

func isContent(fname string) bool {
	return strings.HasSuffix(fname, ".content")
}

func isMarkdown(fname string) bool {
	return strings.HasSuffix(fname, ".md")
}

func WalkAndProcessContents(root string) {
	cwd, err := os.Getwd()
	if err != nil {
		rep.Fatal("ERROR: %s\n", err)
	}
	walk := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Error in processing the path - skip.
			return nil
		}
		if !d.IsDir() {
			// Skip over files.
			return nil
		}
		if filepath.Base(path) == ".git" {
			return fs.SkipDir
		}
		if filepath.Base(path) == GENDIR {
			// Skip GENDIR.
			return fs.SkipDir
		}
		ProcessFilesContent(cwd, path)
		return nil
	}
	if err := filepath.WalkDir(root, walk); err != nil {
		rep.Fatal("ERROR: %s\n", err)
	}
}

func WalkAndProcessMarkdowns(root string) {
	cwd, err := os.Getwd()
	if err != nil {
		rep.Fatal("ERROR: %s\n", err)
	}
	walk := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Error in processing the path - skip.
			return nil
		}
		if !d.IsDir() {
			// Skip over files.
			return nil
		}
		if filepath.Base(path) == ".git" {
			return fs.SkipDir
		}
		if filepath.Base(path) == GENDIR {
			// Skip GENDIR.
			return fs.SkipDir
		}
		ProcessFilesMarkdown(cwd, path)
		return nil
	}
	if err := filepath.WalkDir(root, walk); err != nil {
		rep.Fatal("ERROR: %s\n", err)
	}
}

func WalkAndProcessPosts(root string) {
	cwd, err := os.Getwd()
	if err != nil {
		rep.Fatal("ERROR: %s\n", err)
	}
	walk := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Error in processing the path - skip.
			return nil
		}
		if !d.IsDir() {
			// Skip over files.
			return nil
		}
		if filepath.Base(path) == ".git" {
			return fs.SkipDir
		}
		if filepath.Base(path) != GENDIR {
			return nil
		}
		// We are in GENDIR - do we have a GENPOSTS subfolder?
		target := filepath.Join(path, GENPOSTS)
		stat, err := os.Stat(target)
		if os.IsNotExist(err) || !stat.IsDir() {
			// GENPOSTS doesn't exist (or is not a directory) so abort.
			return fs.SkipDir
		}
		//fmt.Println("ABOUT TO PROCESS POSTS IN ", target)
		ProcessFilesPosts(cwd, target)
		return fs.SkipDir
	}
	if err := filepath.WalkDir(root, walk); err != nil {
		rep.Fatal("ERROR: %s\n", err)
	}
}

func targetFilename(src string, srcSuffix string, tgtSuffix string) string {
	target := src
	if strings.HasSuffix(src, "."+srcSuffix) {
		target = strings.TrimSuffix(target, "."+srcSuffix)
	}
	return target + "." + tgtSuffix
}

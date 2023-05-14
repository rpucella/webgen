package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const TEMPLATE = "CONTENT.template"
const SUBTEMPLATE = "SUB.template"
const MDTEMPLATE = "MARKDOWN.template"
const SUMMARYTEMPLATE = "SUMMARY.template"
const GENDIR = "__src" // Can also have a leading .
const GENPOSTS = "POSTS"
const POSTMD = "index.md"

var rep *log.Logger = log.New(os.Stdout, "" /* log.Ldate| */, log.Ltime)

type Content struct {
	Title         string
	Date          time.Time
	FormattedDate string
	Reading       string
	Key           string
	Body          template.HTML
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

func isGenDir(path string) bool {
	base := filepath.Base(path)
	if base == GENDIR {
		return true
	}
	if base == "."+GENDIR {
		return true
	}
	return false
}

func isGenPosts(path string) bool {
	base := filepath.Base(path)
	if base == GENPOSTS {
		return true
	}
	if base == "."+GENPOSTS {
		return true
	}
	return false
}

func isSkippedDirectory(path string) bool {
	if filepath.Base(path) == ".git" {
		return true
	}
	if isGenDir(path) {
		return true
	}
	if isGenPosts(path) {
		return true
	}
	return false
}

func identifyGenDir(path string) (string, error) {
	fileinfo, err := os.Stat(filepath.Join(path, GENDIR))
	if err != nil {
		fileinfo, err := os.Stat(filepath.Join(path, "."+GENDIR))
		if err != nil {
			return "", err
		}
		if fileinfo.IsDir() {
			return "." + GENDIR, nil
		}
		return "", fmt.Errorf("GENDIR not a directory")
	}
	if fileinfo.IsDir() {
		return GENDIR, nil
	}
	return "", fmt.Errorf("GENDIR not a directory")
}

func identifyGenPosts(path string) (string, error) {
	fileinfo, err := os.Stat(filepath.Join(path, GENDIR, GENPOSTS))
	if err != nil {
		fileinfo, err := os.Stat(filepath.Join(path, "."+GENDIR, GENPOSTS))
		if err != nil {
			return "", err
		}
		if fileinfo.IsDir() {
			return filepath.Join("."+GENDIR, GENPOSTS), nil
		}
		return "", fmt.Errorf("GENPOSTS not a directory")
	}
	if fileinfo.IsDir() {
		return filepath.Join(GENDIR, GENPOSTS), nil
	}
	return "", fmt.Errorf("GENPOSTS not a directory")
}

func identifyGenDirPath(path string) (string, error) {
	genDir, err := identifyGenDir(path)
	if err != nil {
		return "", err
	}
	return filepath.Join(path, genDir), nil
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
		if isSkippedDirectory(path) {
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
		if isSkippedDirectory(path) {
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
		if isSkippedDirectory(path) {
			return fs.SkipDir
		}
		ProcessFilesPosts(cwd, path)
		return nil
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

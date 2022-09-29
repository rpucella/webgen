package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func ProcessFileContent(w io.Writer, fname string) error {
	rep.Println("------------------------------------------------------------")
	rep.Printf("Processing %s\n", fname)
	templates, err := findTemplate(fname)
	if err != nil {
		return err
	}
	if len(templates) == 0 {
		return fmt.Errorf("No template found")
	}
	main, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	current := template.HTML(main)
	for _, tinfo := range templates {
		tpl := tinfo.template
		tname := tinfo.name
		rep.Printf("Using template %s\n", tname)
		c := Content{"", "", "", current}
		current, err = ProcessTemplate(tpl, c)
		if err != nil {
			return err
		}
	}
	bytes := []byte(current)
	if _, err := w.Write(bytes); err != nil {
		return err
	}
	return nil
}

func ProcessTemplate(tpl *template.Template, content Content) (template.HTML, error) {
	var b strings.Builder
	if err := tpl.Execute(&b, content); err != nil {
		return template.HTML(""), err
	}
	result := template.HTML(b.String())
	return result, nil
}

type template_info struct {
	template *template.Template
	name     string
}

func findTemplate(path string) ([]template_info, error) {
	// Given a path, find the nearest enclosing _gentemplate file.
	// If encountering _gentemplate_sub file, add to list but continue looking.
	result := make([]template_info, 0)
	previous, _ := filepath.Abs(path)
	current := filepath.Dir(previous)
	for current != previous {
		subtname := filepath.Join(current, GENDIR, SUBTEMPLATE)
		subtpl, err := template.ParseFiles(subtname)
		if err == nil {
			result = append(result, template_info{subtpl, subtname})
		}
		tname := filepath.Join(current, GENDIR, TEMPLATE)
		tpl, err := template.ParseFiles(tname)
		if err == nil {
			result = append(result, template_info{tpl, tname})
			return result, nil
		}
		previous = current
		current = filepath.Dir(current)
	}
	return nil, fmt.Errorf("no template found")
}

func ProcessFilesContent(cwd string, path string) {
	entries, err := os.ReadDir(filepath.Join(path, GENDIR))
	if err != nil {
		// if we can't read GENDIR, skip.
		return
	}
	for _, d := range entries {
		if !d.IsDir() && isContent(d.Name()) {
			relPath, err := filepath.Rel(cwd, path)
			if err != nil {
				relPath = path
			}
			target := filepath.Join(relPath, targetFilename(d.Name(), "content", "html"))
			w, err := os.Create(target)
			if err != nil {
				w.Close()
				rep.Printf("ERROR: %s\n", err)
				continue
			}
			if err := ProcessFileContent(w, filepath.Join(relPath, GENDIR, d.Name())); err != nil {
				w.Close()
				rep.Printf("ERROR: %s\n", err)
				continue
			}
			rep.Printf("Wrote %s", target)
			w.Close()
		}
	}
}
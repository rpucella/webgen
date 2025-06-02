package gen

import (
	"github.com/russross/blackfriday/v2"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Metadata struct {
	Title   string
	Date    time.Time
	Reading string
}

func ProcessFileMarkdown(w io.Writer, fname string) error {
	rep.Printf("%s\n", fname)
	md, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	metadata, restmd, err := ExtractMetadata(md)
	if err != nil {
		return err
	}
	output := blackfriday.Run(restmd, blackfriday.WithNoExtensions())
	tpl, tname, err := FindMarkdownTemplate(fname)
	if tpl != nil {
		rep.Printf("  using markdown template %s\n", tname)
		result, err := ProcessMarkdownTemplate(tpl, metadata, template.HTML(output))
		if err != nil {
			return err
		}
		output = []byte(result)
	}
	if _, err := w.Write(output); err != nil {
		return err
	}
	return nil
}

const draftTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <style>
      body {
          font-family: serif;
          font-size: 24px;
          margin: 0px 128px;
      }

      p {
          line-height: 2.5;
          margin: 32px 0;
      }

      p code {
          font-size: 20px;
      }

      pre {
          font-size: 16px;
          line-height: initial;
          margin: 32px 0;
          border-left: 4px solid #333333;
          padding-left: 16px;
      }

      h1, h2, h3, h4, h5, h6 {
          font-weight: normal;
          margin: 64px 0 32px 0;
      }

    </style>
  </head>
  <body>
{{.}}
  </body>
</html>
`

func ProcessFileMarkdownDraft(fname string) error {
	rep.Printf("%s\n", fname)
	f, err := os.CreateTemp("", "draft*.html")
	if err != nil {
		return err
	}
	md, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	_, restmd, err := ExtractMetadata(md)
	if err != nil {
		return err
	}
	body := blackfriday.Run(restmd, blackfriday.WithNoExtensions())
	mdtpl, err := template.New("draft").Parse(draftTemplate)
	var sb strings.Builder
	if err := mdtpl.Execute(&sb, template.HTML(body)); err != nil {
		return err
	}
	output := []byte(template.HTML(sb.String()))
	//rep.Println(sb.String())
	if _, err := f.Write(output); err != nil {
		return err
	}
	rep.Printf("Draft file: %s\n", f.Name())
	cmd := exec.Command("open", "-a", "Firefox", f.Name())
	cmd.Run()
	return nil
}

func ExtractMetadata(md []byte) (Metadata, []byte, error) {
	title := ""
	date := time.Time{}
	reading := ""
	lines := strings.Split(string(md), "\n")
	foundMetadata := false
	for idx, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			if line == "---" {
				if foundMetadata {
					// We're done.
					rest := []byte(strings.Join(lines[idx+1:], "\n"))
					return Metadata{title, date, reading}, rest, nil
				}
				foundMetadata = true
			} else if foundMetadata {
				fields := strings.Split(line, ":")
				if len(fields) == 2 {
					fieldname := strings.TrimSpace(fields[0])
					fieldvalue := strings.TrimSpace(fields[1])
					switch fieldname {
					case "title":
						title = fieldvalue
					case "reading":
						reading = fieldvalue
					case "date":
						tDate, err := time.Parse("2006-01-02", fieldvalue)
						if err != nil {
							date = time.Time{}
						} else {
							date = tDate
						}
					}
				}
			}
		}
	}
	if foundMetadata {
		return Metadata{}, md, nil
	}
	return Metadata{}, md, nil
}

func FormatDate(date time.Time) string {
	if date.IsZero() {
		return "-"
	} else {
		return date.Format("Jan 2, 2006")
	}
}

func ProcessMarkdownTemplate(tpl *template.Template, metadata Metadata, content template.HTML) (template.HTML, error) {
	c := Content{metadata.Title, metadata.Date, FormatDate(metadata.Date), metadata.Reading, "", content}
	var b strings.Builder
	if err := tpl.Execute(&b, c); err != nil {
		return template.HTML(""), err
	}
	result := template.HTML(b.String())
	return result, nil
}

func FindMarkdownTemplate(path string) (*template.Template, string, error) {
	// Given a path, find the nearest enclosing _gentemplate_md file.
	previous, _ := filepath.Abs(path)
	current := filepath.Dir(previous)
	for current != previous {
		gdPath, err := identifyGenDirPath(current)
		if err == nil {
			mdtname := filepath.Join(gdPath, MDTEMPLATE)
			mdtpl, err := template.ParseFiles(mdtname)
			if err == nil {
				return mdtpl, mdtname, nil
			}
		}
		previous = current
		current = filepath.Dir(current)
	}
	return nil, "", nil
}

func ProcessFilesMarkdown(cwd string, path string) {
	gdPath, err := identifyGenDirPath(path)
	if err != nil {
		return
	}
	entries, err := os.ReadDir(gdPath)
	if err != nil {
		// if we can't read GENDIR, skip.
		return
	}
	for _, d := range entries {
		if !d.IsDir() && IsMarkdown(d.Name()) {
			relPath, err := filepath.Rel(cwd, gdPath)
			if err != nil {
				relPath = gdPath
			}
			target := filepath.Join(relPath, targetFilename(d.Name(), "md", "content"))
			w, err := os.Create(target)
			if err != nil {
				w.Close()
				rep.Printf("ERROR: %s\n", err)
				continue
			}
			if err := ProcessFileMarkdown(w, filepath.Join(relPath, d.Name())); err != nil {
				w.Close()
				rep.Printf("ERROR: %s\n", err)
				continue
			}
			rep.Printf("  wrote %s", target)
			w.Close()
		}
	}
}

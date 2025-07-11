package gen

import (
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// This type might be related to Metadata

type PostInfo struct {
	Title   string
	Date    time.Time
	Reading string
	// Key is of the form YYYY/entry-name and is the folder under `posts/` that contains the generated post.
	Key string
	// Field `year` is not used but kept around in case it's needed.
	Year int
}

func ExtractPosts(path string) ([]PostInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	posts := make([]PostInfo, 0)
	for _, y := range entries {
		// Only look at years.
		if y.IsDir() {
			year, err := strconv.Atoi(y.Name())
			if err != nil {
				continue
			}
			year = year
			subEntries, err := os.ReadDir(filepath.Join(path, y.Name()))
			if err != nil {
				return nil, err
			}
			for _, d := range subEntries {
				if d.IsDir() && d.Name() != GENDIR && d.Name() != ("."+GENDIR) {
					md, err := ioutil.ReadFile(filepath.Join(path, y.Name(), d.Name(), POSTMD))
					if err != nil {
						return nil, err
					}
					metadata, _, err := ExtractMetadata(md)
					if err != nil {
						return nil, err
					}
					posts = append(posts, PostInfo{metadata.Title, metadata.Date, metadata.Reading, filepath.Join(y.Name(), d.Name()), year})
				}
			}
		}
	}
	return posts, nil
}

type byDate []PostInfo

func (s byDate) Len() int {
	return len(s)
}

func (s byDate) Swap(i int, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byDate) Less(i int, j int) bool {
	// Alphabetical order is fine for now for dates.
	return s[i].Date.After(s[j].Date)
}

func ProcessFilesPosts(cwd string, path string) {
	genPosts, err := identifyGenPosts(path)
	if err != nil {
		return
	}
	// Get full list of posts.
	postPath := filepath.Join(path, genPosts)
	rep.Printf("%s\n", postPath)
	posts, err := ExtractPosts(postPath)
	if err != nil {
		rep.Printf("ERROR: %s\n", err)
		return
	}
	sort.Sort(byDate(posts))
	///rep.Printf("Posts = %v\n", posts)
	relPath, err := filepath.Rel(cwd, path)
	if err != nil {
		relPath = path
	}
	// Clear out /post folder completely.
	postDir := filepath.Join(relPath, POSTDIR)
	rep.Printf("  removing %s\n", postDir)
	os.RemoveAll(postDir)
	if err := os.Mkdir(postDir, 0755); err != nil {
		rep.Printf("ERROR: %s\n", err)
		return
	}
	// Copy post folders.
	for _, p := range posts {
		if err := os.MkdirAll(filepath.Join(postDir, p.Key), 0755); err != nil {
			rep.Printf("ERROR: %s\n", err)
			continue
		}
		// Copy content of folder p.Key.
		// This does not go into subfolders!
		rep.Printf("  copying %s\n", p.Key)
		postEntries, err := os.ReadDir(filepath.Join(relPath, genPosts, p.Key))
		if err != nil {
			rep.Printf("ERROR: %s\n", err)
			continue
		}
		for _, f := range postEntries {
			if !f.IsDir() {
				srcPath := filepath.Join(relPath, genPosts, p.Key)
				srcName := f.Name()
				dstPath := filepath.Join(postDir, p.Key)
				dstName := f.Name()
				if f.Name() == POSTMD {
					// Eventually, want to do something smart, like insert
					//  previous/next keys in the metadata to be able to handle
					//  navigation at the level of posts.
					// We can only do that if we have the full list of posts
					//  though, so we'll need to restructure to read all
					//  posts, sort them by date, THEN process them.
					// genDir, err := identifyGenDir(dstPath)
					// if err != nil {
					// 	rep.Printf("ERROR: %s\n", err)
					// 	continue
					// }
					dstPath = filepath.Join(dstPath, "."+GENDIR)
					dstName = "index.md"
					if err := os.Mkdir(dstPath, 0755); err != nil {
						rep.Printf("ERROR: %s\n", err)
						continue
					}
				}
				fsrc, err := os.Open(filepath.Join(srcPath, srcName))
				if err != nil {
					rep.Printf("ERROR: %s\n", err)
					continue
				}
				fdst, err := os.Create(filepath.Join(dstPath, dstName))
				if err != nil {
					rep.Printf("ERROR: %s\n", err)
					fsrc.Close()
					continue
				}
				if _, err := io.Copy(fdst, fsrc); err != nil {
					rep.Printf("ERROR: %s\n", err)
					fsrc.Close()
					fdst.Close()
					continue
				}
			}
		}
	}
	// Extract list of summaries.
	genDir, err := identifyGenDir(relPath)
	if err != nil {
		rep.Printf("ERROR: %s\n", err)
		return
	}
	target := filepath.Join(relPath, genDir, "index.content")
	w, err := os.Create(target)
	defer w.Close()
	if err != nil {
		rep.Printf("ERROR: %s\n", err)
		return
	}
	postsContent := make([]Content, 0, len(posts))
	for _, p := range posts {
		src := filepath.Join(relPath, genPosts, p.Key, POSTMD)
		metadata, err := ProcessFilePost(p.Key, src)
		if err != nil {
			rep.Printf("ERROR: %s\n", err)
			continue
		}
		content := Content{metadata.Title, metadata.Date, FormatDate(metadata.Date), metadata.Reading, p.Key, template.HTML("")}
		postsContent = append(postsContent, content)
	}
	tpl, tname, err := FindSummaryTemplate(postPath)
	output := []byte("")
	if tpl != nil {
		rep.Printf("  using summary template %s\n", tname)
		content := SummaryContent{postsContent}
		result, err := ProcessSummaryTemplate(tpl, content)
		if err != nil {
			rep.Printf("ERROR: %s\n", err)
			return
		}
		output = []byte(result)
	}
	if _, err := w.Write(output); err != nil {
		rep.Printf("ERROR: %s\n", err)
		return
	}
	rep.Printf("  wrote %s", target)
}

type SummaryContent struct {
	Posts []Content
}

func ProcessSummaryTemplate(tpl *template.Template, content SummaryContent) (template.HTML, error) {
	var b strings.Builder
	if err := tpl.Execute(&b, content); err != nil {
		return template.HTML(""), err
	}
	result := template.HTML(b.String())
	return result, nil
}

func ProcessFilePost(key string, fname string) (Metadata, error) {
	rep.Printf("%s\n", fname)
	md, err := ioutil.ReadFile(fname)
	if err != nil {
		return Metadata{}, err
	}
	metadata, _, err := ExtractMetadata(md)
	return metadata, err
}

func FindSummaryTemplate(path string) (*template.Template, string, error) {
	// Given a path, find the nearest enclosing SUMMARY.template file.
	previous, _ := filepath.Abs(path)
	current := filepath.Dir(previous)
	for current != previous {
		///rep.Printf("[trying %s]\n", current)
		gdPath, err := identifyGenDirPath(current)
		if err == nil {
			mdtname := filepath.Join(gdPath, SUMMARYTEMPLATE)
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

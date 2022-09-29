package main

import (
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

// This type might be related to Metadata

type PostInfo struct {
	Title string
	Date  string
	Key   string
}

func ProcessFilePost(w io.Writer, key string, fname string) error {
	rep.Println("------------------------------------------------------------")
	rep.Printf("Processing %s\n", fname)
	md, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	metadata, _, err := ExtractMetadata(md)
	if err != nil {
		return err
	}
	tpl, tname, err := FindMarkdownTemplate(fname)
	output := []byte("")
	if tpl != nil {
		rep.Printf("Using markdown template %s\n", tname)
		content := Content{metadata.Title, metadata.Date, key, template.HTML("")}
		result, err := ProcessTemplate(tpl, content)
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

func ExtractPosts(path string) ([]PostInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	posts := make([]PostInfo, 0)
	for _, d := range entries {
		if d.IsDir() {
			md, err := ioutil.ReadFile(filepath.Join(path, d.Name(), POSTMD))
			if err != nil {
				return nil, err
			}
			metadata, _, err := ExtractMetadata(md)
			if err != nil {
				return nil, err
			}
			posts = append(posts, PostInfo{metadata.Title, metadata.Date, d.Name()})
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
	return s[i].Date > s[j].Date
}

func ProcessFilesPosts(cwd string, path string) {
	// Called with path = path of the GENPOSTS folder.
	// Get full list of posts.
	posts, err := ExtractPosts(path)
	if err != nil {
		return
	}
	sort.Sort(byDate(posts))
	rep.Printf("Posts = %v\n", posts)
	relPath, err := filepath.Rel(cwd, path)
	if err != nil {
		relPath = path
	}
	// Clear out /post folder completely.
	postDir := filepath.Join(relPath, "..", "post")
	rep.Printf("Removing %s\n", postDir)
	os.RemoveAll(postDir)
	if err := os.Mkdir(postDir, 0755); err != nil {
		rep.Printf("ERROR: %s\n", err)
		return
	}
	// Copy post folders.
	for _, p := range posts {
		if err := os.Mkdir(filepath.Join(postDir, p.Key), 0755); err != nil {
			rep.Printf("ERROR: %s\n", err)
			continue
		}
		// Copy content of folder p.Key.
		// This does not go into subfolders!
		rep.Printf("Copying %s\n", p.Key)
		postEntries, err := os.ReadDir(filepath.Join(relPath, p.Key))
		if err != nil {
			rep.Printf("ERROR: %s\n", err)
			continue
		}
		for _, f := range postEntries {
			if !f.IsDir() {
				srcPath := filepath.Join(relPath, p.Key)
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
					dstPath = filepath.Join(dstPath, GENDIR)
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
	target := filepath.Join(relPath, "..", GENDIR, "index.content")
	w, err := os.Create(target)
	if err != nil {
		w.Close()
		rep.Printf("ERROR: %s\n", err)
		return
	}
	for _, p := range posts {
		src := filepath.Join(relPath, p.Key, POSTMD)
		if err := ProcessFilePost(w, p.Key, src); err != nil {
			rep.Printf("ERROR: %s\n", err)
			continue
		}
		rep.Printf("Wrote to %s", target)
	}
	w.Close()
}
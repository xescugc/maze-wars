package templates

import (
	"embed"
	"html/template"
	"io/fs"
	"path/filepath"
	"regexp"
)

const (
	viewsDir  = "views"
	extension = "/*.tmpl"
)

var (
	layoutsDir = filepath.Join(viewsDir, "layouts")

	//go:embed views/layouts/* views/home/* views/game/* views/docs/*
	files embed.FS

	// Templates is the cache of all the templates we have
	Templates map[string]*template.Template

	idR = regexp.MustCompile(`^[^a-z]+|[^\w]+`)
)

func init() {
	if Templates == nil {
		Templates = make(map[string]*template.Template)
	}

	loadTemplates(viewsDir)

	return
}

func loadTemplates(path string) error {
	tmplFiles, err := fs.ReadDir(files, path)
	if err != nil {
		panic(err)
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			loadTemplates(filepath.Join(path, tmpl.Name()))
			continue
		}

		newpath := filepath.Join(path, tmpl.Name())

		if _, ok := Templates[newpath]; ok {
			continue
		}

		pt := template.New(tmpl.Name())

		pt, err := pt.ParseFS(files, newpath, filepath.Join(layoutsDir, extension))
		if err != nil {
			panic(err)
		}

		Templates[newpath] = pt
	}

	return nil
}

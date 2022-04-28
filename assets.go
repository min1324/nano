package main

import (
	"embed"
	_ "embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"path"
)

//go:embed static
var Assets embed.FS

var Tmpl *template.Template

func init() {
	Tmpl = template.Must(template.New("").ParseFS(Assets, "static/*.html"))
}

type fsFunc func(name string) (fs.File, error)

// Open implement fs.FS.
func (fs fsFunc) Open(name string) (fs.File, error) {
	return fs(name)
}

// AssetsHandler static assets handler.
//
// add to route:
// ServeMux.Handle("/prefix/", AssetsHandler("/prefix/", Assets, "./static"))
func AssetsHandler(prefix string, assets embed.FS, root string) http.Handler {
	handler := fsFunc(func(name string) (fs.File, error) {
		assetsPath := path.Join(root, name)
		file, err := assets.Open(assetsPath)
		if err != nil {
			return nil, err
		}
		return file, err
	})
	return http.StripPrefix(prefix, http.FileServer(http.FS(handler)))
}

// redirect
func redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusFound)
}

// render *.html template
func render(w http.ResponseWriter, html string, data interface{}) {
	t := Tmpl.Lookup(html)
	if t != nil {
		t.Execute(w, data)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func Json(w http.ResponseWriter, msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	_, err = w.Write(data)
	return err
}

func Response(w http.ResponseWriter, code int, msg string, data any) {
	Json(w, Data{"code": code, "msg": msg, "data": data})
}

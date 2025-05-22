package web

import (
	"embed"
	"html/template"
	"net/http"
	"path"
	"strings"
)

//go:embed static/**
var content embed.FS

type WebUsecase struct{}

func NewWebUsecase() *WebUsecase {
	return &WebUsecase{}
}

func (u *WebUsecase) RenderIndex(w http.ResponseWriter, r *http.Request) error {
	var file string
	if r.URL.Path == "/" {
		file = "static/home/index.html"
	} else {
		cleanPath := strings.Trim(r.URL.Path, "/")
		lastSection := path.Base(cleanPath)
		if strings.Contains(lastSection, ".") {
			file = path.Join("static", cleanPath)
		} else {
			file = path.Join("static", cleanPath, "index.html")
		}
	}

	ext := path.Ext(file)
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		data, err := content.ReadFile(file)
		if err != nil {
			http.NotFound(w, r)
			return err
		}
		w.Write(data)
		return nil
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		data, err := content.ReadFile(file)
		if err != nil {
			http.NotFound(w, r)
			return err
		}
		w.Write(data)
		return nil
	}

	tmpl, err := template.ParseFS(content, file)
	if err != nil {
		http.NotFound(w, r)
		return err
	}
	return tmpl.Execute(w, nil)
}

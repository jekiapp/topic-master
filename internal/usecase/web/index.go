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
	cleanPath := strings.Trim(r.URL.Path, "/")
	file := path.Join("static", cleanPath)
	if cleanPath == "" || !strings.Contains(path.Base(cleanPath), ".") {
		file = path.Join("static", cleanPath, "index.html")
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
	case ".png":
		w.Header().Set("Content-Type", "image/png")
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

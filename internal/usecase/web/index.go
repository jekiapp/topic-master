package web

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed home/index.html
var content embed.FS

type WebUsecase struct {
	tmpl *template.Template
}

func NewWebUsecase() (*WebUsecase, error) {
	tmpl, err := template.ParseFS(content, "home/index.html")
	if err != nil {
		return nil, err
	}

	return &WebUsecase{
		tmpl: tmpl,
	}, nil
}

func (u *WebUsecase) RenderIndex(w http.ResponseWriter, r *http.Request) error {
	return u.tmpl.Execute(w, nil)
}

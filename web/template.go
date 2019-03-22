package web

import (
	"bytes"
	"filekeep/assets/templates"
	"fmt"
	"html/template"
	"net/http"
)

var (
	headerTpl = template.Must(template.New("header").Parse(templates.HTMLHeader))
	footerTpl = template.Must(template.New("footer").Parse(templates.HTMLFooter))

	notFoundTpl = template.Must(template.New("404").Parse(templates.HTML404))

	aboutTpl   = template.Must(template.New("about").Parse(templates.HTMLAbout))
	dirListTpl = template.Must(template.New("list").Funcs(funcMap).Parse(templates.HTMLDirList))
)

func templateHandler(w http.ResponseWriter, r *http.Request, t *template.Template, data interface{}) {
	pageIncomplete := false
	headerDataCopy := &headerData
	headerDataCopy.DarkTheme = r.Context().Value("dark-theme").(bool)
	buffer := bytes.NewBufferString("")

	if err := headerTpl.Execute(buffer, headerDataCopy); err != nil {
		pageIncomplete = true
	}

	if err := t.Execute(buffer, data); err != nil {
		pageIncomplete = true
	}

	if err := footerTpl.Execute(buffer, nil); err != nil {
		pageIncomplete = true
	}

	if _, err := fmt.Fprint(w, buffer); err != nil {
		pageIncomplete = true
	}

	if pageIncomplete {
		res := httpResponse{Error: true, Message: "couldn't execute template"}
		res.JSON(http.StatusInternalServerError, w)
		return
	}
}

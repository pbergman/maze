package http

import "html/template"

var templates *template.Template

func init() {
	templates = template.Must(template.ParseFiles("template/base.html", "template/list-group.html"))
}


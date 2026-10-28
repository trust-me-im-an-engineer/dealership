package templ

import (
	"embed"
	"html/template"
)

// Шаблоны встраиваются в исполняемый файл, чтобы сервер не зависел от рабочей директории.
//
//go:embed templates/*.html
var files embed.FS

// Parse объединяет базовую разметку с блоками указанной страницы.
func Parse(name string) *template.Template {
	return template.Must(template.ParseFS(files, "templates/base.html", "templates/"+name))
}

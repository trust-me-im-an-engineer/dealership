package templ

import (
	"embed"
	"html/template"
	"io/fs"
)

// Шаблоны встраиваются в исполняемый файл, чтобы сервер не зависел от рабочей директории.
//
//go:embed templates/*.html static/js/*.js
var files embed.FS

// Static возвращает встроенные файлы JavaScript для HTTP-сервера.
func Static() fs.FS {
	return files
}

// Parse объединяет базовую разметку с блоками указанной страницы.
func Parse(name string, partials ...string) *template.Template {
	paths := []string{"templates/base.html", "templates/" + name}
	for _, partial := range partials {
		paths = append(paths, "templates/"+partial)
	}
	return template.Must(template.ParseFS(files, paths...))
}

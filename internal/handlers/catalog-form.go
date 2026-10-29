package handlers

// catalogFormPage содержит общие данные строки добавления или редактирования каталога.
type catalogFormPage struct {
	Action     string
	FormID     string
	Name       string
	OrderCount int64
	IsEdit     bool
}

package errorvals

import (
	"errors"
)

// Сюда следует добавлять все ошибки, которые нужно потом проверять через errors.Is
// Ошибка не должна выдавать внутреннюю логику программы, поэтому не стоит злоупотреблять
// errors.Is для определения статуса http ответа

var ErrObjectNotFoundInRepoError = errors.New("object not found in repo")

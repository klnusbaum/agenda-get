package sites

import "fmt"

type FinderError struct {
	error
	html     []byte
	filename string
}

func NewFinderError(err error, html []byte, entity string) error {
	return &FinderError{err, html, fmt.Sprintf("%s-error-content.html", entity)}
}

func (e *FinderError) HTML() []byte {
	var cpy []byte
	copy(cpy, e.html)
	return cpy
}

func (e *FinderError) Filename() string {
	return e.filename
}

package sites

import "io"

type Agenda struct {
	Entity  string
	Name    string
	Content io.ReadCloser
}

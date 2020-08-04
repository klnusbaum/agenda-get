package sites

import "io"

type Agenda struct {
	Name    string
	Content io.ReadCloser
}

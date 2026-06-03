package diag

import "fmt"

// Error represents a compiler diagnostic with source location.
type Error struct {
	File    string
	Line    int
	Column  int
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("[ERROR] %s:%d:%d — %s", e.File, e.Line, e.Column, e.Message)
}

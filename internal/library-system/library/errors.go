package library

import "errors"

var (
	errLibraryNotFound = errors.New("library not found")
	errBookNotFound    = errors.New("book not found")
	errRecordNotFound  = errors.New("record not found")
)

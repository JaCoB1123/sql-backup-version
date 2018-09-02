package main

import (
	"fmt"
	"net/http"

	"path"
	"strings"
)

// ShiftPath splits off the first component of p, which will be cleaned of
// relative components before processing. head will never contain a slash and
// tail will always be a rooted path without trailing slash.
func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

// HasError checks if an error has occured. If yes, returns the error message
// as HTTP Response and sets the Status to 500
func HasError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	code := http.StatusInternalServerError
	text := fmt.Sprintf("%s\n\n%s", http.StatusText(code), err.Error())
	http.Error(w, text, code)
	return true
}

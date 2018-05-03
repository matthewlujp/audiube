package main

import (
	"strings"
)

// name2ContentType returns appropriate content-type for response header according to return file extention
func name2ContentType(filename string) string {
	logger.Print("name2contentype ", filename)
	if strings.Contains(filename, ".") {
		values := strings.Split(filename, ".")
		switch extention := values[len(values)-1]; extention {
		case "jpeg", "jpg", "JPEG":
			return "image/jpeg"
		case "png":
			return "image/png"
		case "html":
			return "text/html"
		case "css":
			return "text/css"
		default:
			return "text/plain; charset=utf-8"
		}
	}

	return "text/plain; charset=utf-8" // default
}

package helpers

import "strings"

func Split(fileName string) (name string, ext string) {
	parts := strings.Split(fileName, ".")
	if len(parts) < 2 {
		return fileName, ""
	}

	return strings.Join(parts[:len(parts)-1], "."), parts[len(parts)-1]
}

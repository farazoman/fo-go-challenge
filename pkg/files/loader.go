package files

import (
	"log"
	"os"
)

type Loader interface {
	Load(path string) string
}

type SystemLoader struct {
}

func (loader SystemLoader) Load(path string) string {
	// TODO path validation
	content, error := os.ReadFile(path)

	if error != nil {
		log.Fatal(error)
	}

	return string(content)
}

var _ Loader = SystemLoader{}

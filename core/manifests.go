package core

import (
	"os"
	"path/filepath"
)

var manifestsMap = make(map[string]Manifest, 5000)

func LoadManifests(dir string) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {

	})
}

package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

var manifestsMap = make(map[string]Manifest, 5000)

func LoadManifests() error {
	_, err := os.Stat(ManifestDir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(ManifestDir, 0755)
		}
		return err
	}
	err = filepath.Walk(ManifestDir, func(path string, info os.FileInfo, ferr error) error {
		if ferr != nil {
			return ferr
		}
		if info.IsDir() {
			return nil
		}
		logrus.Debugf("load manifest: %s", path)
		ss := strings.Split(strings.Trim(path, ManifestDir), string(filepath.Separator))
		prefix := strings.Join(ss[:len(ss)-1], "/")
		tag := strings.Trim(ss[len(ss)-1], ".json")
		cacheKey := strings.TrimPrefix(fmt.Sprintf("%s:%s", prefix, tag), "/")
		logrus.Debugf("manifest cache key: %s", cacheKey)
		bs, rerr := ioutil.ReadFile(path)
		if rerr != nil {
			return err
		}
		manifestsMap[cacheKey] = Manifest(bs)
		return nil
	})
	logrus.Infof("load manifests count: %d", len(manifestsMap))
	return err
}

package core

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"

	"github.com/containers/image/v5/manifest"

	"github.com/sirupsen/logrus"
)

var manifestsMap = make(map[string]manifest.Manifest, 5000)

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
		mbs, rerr := ioutil.ReadFile(path)
		if rerr != nil {
			return rerr
		}

		mType := manifest.GuessMIMEType(mbs)
		// ignore blank json file
		if mType == "" {
			return nil
		}

		m, jerr := manifest.FromBlob(mbs, mType)
		if jerr != nil {
			// ignore err
			// refs github.com/containers/image/v5@v5.4.3/manifest/manifest.go:253
			if jerr.Error() != ErrManifestNotImplemented {
				return jerr
			}
			return nil
		}

		manifestsMap[cacheKey] = m
		return nil
	})
	logrus.Infof("load manifests count: %d", len(manifestsMap))
	return err
}

func getImageManifest(imageName string) (manifest.Manifest, error) {
	srcRef, err := docker.Transport.ParseReference("//" + imageName)
	if err != nil {
		return nil, err
	}

	sourceCtx := &types.SystemContext{DockerAuthConfig: &types.DockerAuthConfig{}}
	imageSrcCtx, imageSrcCancel := context.WithTimeout(context.Background(), DefaultCtxTimeout)
	defer imageSrcCancel()
	src, err := srcRef.NewImageSource(imageSrcCtx, sourceCtx)
	if err != nil {
		return nil, err
	}

	getManifestCtx, getManifestCancel := context.WithTimeout(context.Background(), DefaultCtxTimeout)
	defer getManifestCancel()
	mbs, _, err := src.GetManifest(getManifestCtx, nil)
	if err != nil {
		return nil, err
	}

	mType := manifest.GuessMIMEType(mbs)
	if mType == "" {
		return nil, fmt.Errorf("faile to parse image [%s] manifest type", imageName)
	}

	return manifest.FromBlob(mbs, mType)
}

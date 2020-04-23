package core

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	jsoniter "github.com/json-iterator/go"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"

	"github.com/containers/image/v5/manifest"

	"github.com/sirupsen/logrus"
)

var manifestsMap = make(map[string]interface{}, 5000)

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
		switch mType {
		case manifest.DockerV2ListMediaType:
			var m2List manifest.Schema2List
			if err := jsoniter.Unmarshal(mbs, &m2List); err == nil {
				manifestsMap[cacheKey] = m2List
			}
			logrus.Debugf("failed to parse json [%s]: %s", path, err)
		case imgspecv1.MediaTypeImageIndex:
			var o1List manifest.OCI1Index
			if err := jsoniter.Unmarshal(mbs, &o1List); err == nil {
				manifestsMap[cacheKey] = o1List
			}
			logrus.Debugf("failed to parse json [%s]: %s", path, err)
		default:
			if m, err := manifest.FromBlob(mbs, mType); err == nil {
				manifestsMap[cacheKey] = m
			}
			logrus.Debugf("failed to parse json [%s]: %s", path, err)
		}
		return nil
	})
	logrus.Infof("load manifests count: %d", len(manifestsMap))
	return err
}

func getImageManifest(imageName string) (manifest.Manifest, manifest.List, error) {
	srcRef, err := docker.Transport.ParseReference("//" + imageName)
	if err != nil {
		return nil, nil, err
	}

	sourceCtx := &types.SystemContext{DockerAuthConfig: &types.DockerAuthConfig{}}
	imageSrcCtx, imageSrcCancel := context.WithTimeout(context.Background(), DefaultCtxTimeout)
	defer imageSrcCancel()
	src, err := srcRef.NewImageSource(imageSrcCtx, sourceCtx)
	if err != nil {
		return nil, nil, err
	}

	getManifestCtx, getManifestCancel := context.WithTimeout(context.Background(), DefaultCtxTimeout)
	defer getManifestCancel()
	mbs, _, err := src.GetManifest(getManifestCtx, nil)
	if err != nil {
		return nil, nil, err
	}

	mType := manifest.GuessMIMEType(mbs)
	if mType == "" {
		return nil, nil, fmt.Errorf("faile to parse image [%s] manifest type", imageName)
	}
	switch mType {
	case manifest.DockerV2ListMediaType:
		var m2List manifest.Schema2List
		err = jsoniter.Unmarshal(mbs, &m2List)
		if err != nil {
			return nil, nil, err
		}
		return nil, &m2List, nil
	case imgspecv1.MediaTypeImageIndex:
		var o1List manifest.OCI1Index
		err = jsoniter.Unmarshal(mbs, &o1List)
		if err != nil {
			return nil, nil, err
		}
		return nil, &o1List, nil
	default:
		m, err := manifest.FromBlob(mbs, mType)
		if err != nil {
			return nil, nil, err
		}
		return m, nil, nil
	}
}

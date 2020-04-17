package gcrsync

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/containers/image/copy"
	"github.com/containers/image/docker"
	"github.com/containers/image/manifest"
	"github.com/containers/image/signature"
	"github.com/containers/image/types"
)

const (
	ChangeLog       = "CHANGELOG-%s.md"
	ManifestDir     = "manifests"
	GcrImagesTpl    = "https://gcr.io/v2/%s/tags/list"
	GcrImageTagsTpl = "https://gcr.io/v2/%s/%s/tags/list"
	DockerHubImage  = "https://hub.docker.com/v2/repositories/%s/?page_size=100"
	DockerHubTags   = "https://hub.docker.com/v2/repositories/%s/%s/tags/?page_size=100"
)

type DockerHubOption struct {
	Username string
	Password string
	Timeout  time.Duration
}

func syncDockerHub(image Image, opt DockerHubOption) error {

	destImage := Image{
		Repo: "docker.io",
		User: opt.Username,
		Name: image.MergeName(),
		Tag:  image.Tag,
	}

	logrus.Infof("sync %s => %s", image, destImage)

	ctx, cancel := context.WithTimeout(context.Background(), opt.Timeout)
	defer cancel()

	policyContext, err := signature.NewPolicyContext(&signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}})
	if err != nil {
		return err
	}
	defer func() { _ = policyContext.Destroy() }()

	srcRef, err := docker.Transport.ParseReference("//" + image.String())
	if err != nil {
		return err
	}
	destRef, err := docker.Transport.ParseReference("//" + destImage.String())
	if err != nil {
		return err
	}

	sourceCtx := &types.SystemContext{DockerAuthConfig: &types.DockerAuthConfig{}}
	destinationCtx := &types.SystemContext{DockerAuthConfig: &types.DockerAuthConfig{
		Username: opt.Username,
		Password: opt.Password,
	}}

	m, err := copy.Image(ctx, policyContext, destRef, srcRef, &copy.Options{
		ReportWriter:          ioutil.Discard,
		SourceCtx:             sourceCtx,
		DestinationCtx:        destinationCtx,
		ProgressInterval:      1 * time.Second,
		ForceManifestMIMEType: manifest.DockerV2Schema2MediaType,
	})
	if err != nil {
		return err
	}

	storageDir := filepath.Join(ManifestDir, image.Repo, image.Name)
	// ignore other error
	if _, err := os.Stat(storageDir); err != nil {
		if err = os.MkdirAll(storageDir, 0755); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(filepath.Join(storageDir, image.Tag+".json"), m, 0644)
}

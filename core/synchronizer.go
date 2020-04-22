package core

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/types"

	"github.com/sirupsen/logrus"
)

type Synchronizer interface {
	Images() Images
	Sync(ctx context.Context, opt *SyncOption)
}

type SyncOption struct {
	Timeout    time.Duration
	Limit      int
	User       string
	Password   string
	NameSpace  string
	Proxy      string
	QueryLimit int
	Kubeadm    bool
}

type TagsOption struct {
	Timeout time.Duration
}

func New(name string) (Synchronizer, error) {
	switch name {
	case "gcr":
		return &gcr, nil
	case "flannel":
		return &fl, nil
	}
	return nil, fmt.Errorf("failed to create synchronizer %s: unknown synchronizer", name)
}

func syncImages(ctx context.Context, images Images, opt *SyncOption) {
	logrus.Infof("starting sync images, image total: %d", len(images))

	processWg := new(sync.WaitGroup)
	processWg.Add(len(images))

	if opt.Limit == 0 {
		opt.Limit = DefaultLimit
	}
	limitCh := make(chan int, opt.Limit)
	defer close(limitCh)

	sort.Sort(images)
	for _, image := range images {
		go func(image Image) {
			defer func() {
				<-limitCh
				processWg.Done()
			}()
			select {
			case limitCh <- 1:
				logrus.Debugf("process image: %s", image)
				m, err := getImageManifest(image.String())
				if err != nil {
					logrus.Errorf("failed to get image [%s] manifest, error: %s", image.String(), err)
					return
				}
				sm, ok := manifestsMap[image.String()]
				if ok && m == sm {
					logrus.Warnf("image [%s] not changed, skip sync...", image.String())
					return
				}

				err = retry(defaultSyncRetry, defaultSyncRetryTime, func() error {
					return sync2DockerHub(&image, opt)
				})
				if err != nil {
					logrus.Errorf("failed to process image %s, error: %s", image.String(), err)
				}
			case <-ctx.Done():
			}
		}(image)
	}

	processWg.Wait()
}

func sync2DockerHub(image *Image, opt *SyncOption) error {
	destImage := Image{
		Repo: DefaultDockerRepo,
		User: opt.User,
		Name: image.MergeName(),
		Tag:  image.Tag,
	}

	logrus.Infof("sync %s => %s", image, destImage.String())

	ctx, cancel := context.WithTimeout(context.Background(), opt.Timeout)
	defer cancel()

	policyContext, err := signature.NewPolicyContext(
		&signature.Policy{
			Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()},
		},
	)
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
		Username: opt.User,
		Password: opt.Password,
	}}

	m, err := copy.Image(ctx, policyContext, destRef, srcRef, &copy.Options{
		SourceCtx:             sourceCtx,
		DestinationCtx:        destinationCtx,
		ForceManifestMIMEType: manifest.DockerV2Schema2MediaType,
	})
	if err != nil {
		return err
	}
	storageDir := filepath.Join(ManifestDir, image.Repo, image.User, image.Name)
	// ignore other error
	if _, err := os.Stat(storageDir); err != nil {
		if err := os.MkdirAll(storageDir, 0755); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(filepath.Join(storageDir, image.Tag+".json"), m, 0644)
}

func getImageManifest(imageName string) (Manifest, error) {
	srcRef, err := docker.Transport.ParseReference("//" + imageName)
	if err != nil {
		return "", err
	}

	sourceCtx := &types.SystemContext{DockerAuthConfig: &types.DockerAuthConfig{}}
	src, err := srcRef.NewImageSource(context.Background(), sourceCtx)
	if err != nil {
		return "", err
	}

	bs, _, err := src.GetManifest(context.Background(), nil)
	if err != nil {
		return "", err
	}
	return Manifest(bs), nil
}

func getImageTags(imageName string, opt TagsOption) ([]string, error) {
	srcRef, err := docker.Transport.ParseReference("//" + imageName)
	if err != nil {
		return nil, err
	}
	sourceCtx := &types.SystemContext{DockerAuthConfig: &types.DockerAuthConfig{}}
	tagsCtx, tagsCancel := context.WithTimeout(context.Background(), opt.Timeout)
	defer tagsCancel()

	return docker.GetRepositoryTags(tagsCtx, sourceCtx, srcRef)
}

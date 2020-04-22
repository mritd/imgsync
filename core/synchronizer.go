package core

import (
	"context"
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
	User     string        // Docker Hub User
	Password string        // Docker Hub User Password
	Timeout  time.Duration // Sync single image timeout
	Limit    int           // Images sync process limit

	QueryLimit int    // Query Gcr images limit
	NameSpace  string // Gcr image namespace
	Kubeadm    bool   // Sync kubeadm images (change gcr.io to k8s.gcr.io, and remove namespace)
}

type TagsOption struct {
	Timeout time.Duration
}

func NewSynchronizer(name string) Synchronizer {
	switch name {
	case "gcr":
		return &gcr
	case "flannel":
		return &fl
	default:
		logrus.Fatalf("failed to create synchronizer %s: unknown synchronizer", name)
		// just for compiling
		return nil
	}
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
	imageSrcCtx, imageSrcCancel := context.WithTimeout(context.Background(), DefaultCtxTimeout)
	defer imageSrcCancel()
	src, err := srcRef.NewImageSource(imageSrcCtx, sourceCtx)
	if err != nil {
		return "", err
	}

	getManifestCtx, getManifestCancel := context.WithTimeout(context.Background(), DefaultCtxTimeout)
	defer getManifestCancel()
	bs, _, err := src.GetManifest(getManifestCtx, nil)
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

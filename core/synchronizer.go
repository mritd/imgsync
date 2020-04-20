package core

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/containers/image/copy"
	"github.com/containers/image/docker"
	"github.com/containers/image/manifest"
	"github.com/containers/image/signature"
	"github.com/containers/image/types"

	"github.com/sirupsen/logrus"
)

type Synchronizer interface {
	Default() error
	Images() Images
}

type SyncOption struct {
	Timeout  time.Duration
	Limit    int
	User     string
	Password string
}

type ManifestOption struct {
	Timeout time.Duration
	Limit   int
}

func New(name string) (Synchronizer, error) {
	switch name {
	case "gcr":
		gcr.Kubeadm = false
		if err := gcr.Default(); err != nil {
			return nil, err
		}
		return &gcr, nil
	case "flannel":
		if err := fl.Default(); err != nil {
			return nil, err
		}
		return &fl, nil
	case "kubernetes":
		gcr.Kubeadm = true
		if err := gcr.Default(); err != nil {
			return nil, err
		}
		return &gcr, nil
	}

	return nil, fmt.Errorf("failed to create synchronizer %s: unknown synchronizer", name)
}

func syncImages(ctx context.Context, images Images, opt SyncOption) {
	logrus.Info("starting sync images, image total: %d", len(images))

	processWg := new(sync.WaitGroup)
	processWg.Add(len(images))

	if opt.Limit == 0 {
		opt.Limit = DefaultLimit
	}
	limitCh := make(chan int, opt.Limit)
	defer close(limitCh)

	for _, image := range images {
		tmpImage := image
		go func() {
			defer func() {
				<-limitCh
				processWg.Done()
			}()
			select {
			case limitCh <- 1:
				logrus.Debugf("process image: %s", tmpImage)
				err := retry(defaultSyncRetry, defaultSyncRetryTime, func() error {
					return sync2DockerHub(tmpImage, opt)
				})
				if err != nil {
					logrus.Errorf("failed to process image %s, error: %s", tmpImage, err)
				}
			case <-ctx.Done():
			}
		}()
	}

	processWg.Wait()
}

func sync2DockerHub(image Image, opt SyncOption) error {
	destImage := Image{
		Repo: DefaultDockerRepo,
		User: opt.User,
		Name: image.MergeName(),
		Tag:  image.Tag,
	}

	logrus.Infof("sync %s => %s", image, destImage)

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
		ReportWriter:          ioutil.Discard,
		SourceCtx:             sourceCtx,
		DestinationCtx:        destinationCtx,
		ProgressInterval:      1 * time.Second,
		ForceManifestMIMEType: manifest.DockerV2Schema2MediaType,
	})
	if err != nil {
		return err
	}

	storageDir := filepath.Join(DefaultManifestDir, image.Repo, image.Name)
	// ignore other error
	if _, err := os.Stat(storageDir); err != nil {
		if err := os.MkdirAll(storageDir, 0755); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(filepath.Join(storageDir, image.Tag+".json"), m, 0644)
}

func getManifests(ctx context.Context, images Images, opt ManifestOption) Manifests {
	logrus.Infof("starting get image manifests, image total: %d", len(images))

	processWg := new(sync.WaitGroup)
	processWg.Add(len(images))

	if opt.Limit == 0 {
		opt.Limit = DefaultLimit
	}
	limitCh := make(chan Manifest, opt.Limit)

	var ms Manifests

	for _, image := range images {
		tmpImage := image
		go func() {
			defer func() {
				<-limitCh
				processWg.Done()
			}()

			logrus.Debugf("process image: %s", tmpImage)
			err := retry(defaultSyncRetry, defaultSyncRetryTime, func() error {
				if m, err := getImageManifest(tmpImage, opt); err != nil {
					return err
				} else {
					select {
					case limitCh <- m:
					case <-ctx.Done():
					}
					return nil
				}

			})
			if err != nil {
				logrus.Errorf("failed to process image %s, error: %s", tmpImage, err)
			}
		}()
	}

	go func() {
		for m := range limitCh {
			ms = append(ms, m)
		}
	}()

	processWg.Wait()
	close(limitCh)
	return ms
}

func getImageManifest(image Image, opt ManifestOption) (Manifest, error) {
	imgSrcCtx, cancel := context.WithTimeout(context.Background(), opt.Timeout)
	defer cancel()

	srcRef, err := docker.Transport.ParseReference("//" + image.String())
	if err != nil {
		logrus.Fatal(err)
	}

	sourceCtx := &types.SystemContext{DockerAuthConfig: &types.DockerAuthConfig{}}
	src, err := srcRef.NewImageSource(imgSrcCtx, sourceCtx)
	if err != nil {
		logrus.Fatal(err)
	}

	manifestCtx, cancel := context.WithTimeout(context.Background(), opt.Timeout)
	defer cancel()

	bs, _, err := src.GetManifest(manifestCtx, nil)
	if err != nil {
		return Manifest{}, err
	}
	return Manifest{
		Name:      image.String(),
		JSONValue: string(bs),
	}, nil
}

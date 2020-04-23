package core

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"

	"github.com/containers/image/v5/manifest"

	jsoniter "github.com/json-iterator/go"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/types"

	"github.com/sirupsen/logrus"
)

type Synchronizer interface {
	Images(ctx context.Context) Images
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

	pool, err := ants.NewPool(opt.Limit, ants.WithPreAlloc(true), ants.WithPanicHandler(func(i interface{}) {
		logrus.Error(i)
	}))
	if err != nil {
		logrus.Fatalf("failed to create goroutines pool: %s", err)
	}
	sort.Sort(images)
	for _, tmpImg := range images {
		image := tmpImg
		err = pool.Submit(func() {
			defer func() {
				processWg.Done()
			}()

			select {
			case <-ctx.Done():
			default:
				logrus.Debugf("process image: %s", image)

				m, l, needSync := checkSync(image)
				if !needSync {
					return
				}

				rerr := retry(defaultSyncRetry, defaultSyncRetryTime, func() error {
					return sync2DockerHub(&image, opt)
				})
				if rerr != nil {
					logrus.Errorf("failed to process image %s, error: %s", image.String(), rerr)
					return
				}

				storageDir := filepath.Join(ManifestDir, image.Repo, image.User, image.Name)
				// ignore other error
				if _, err = os.Stat(storageDir); err != nil {
					if err = os.MkdirAll(storageDir, 0755); err != nil {
						logrus.Errorf("failed to storage image [%s] manifests: %s", image.String(), err)
					}
				}
				var bs []byte
				if m != nil {
					bs, err = jsoniter.MarshalIndent(m, "", "    ")
				} else {
					bs, err = jsoniter.MarshalIndent(l, "", "    ")
				}
				if err != nil {
					logrus.Errorf("failed to storage image [%s] manifests: %s", image.String(), err)
				}
				if ferr := ioutil.WriteFile(filepath.Join(storageDir, image.Tag+".json"), bs, 0644); ferr != nil {
					logrus.Errorf("failed to storage image [%s] manifests: %s", image.String(), ferr)
				}
			}
		})
		if err != nil {
			logrus.Fatalf("failed to submit task: %s", err)
		}
	}
	processWg.Wait()
	pool.Release()
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

	srcRef, err := docker.ParseReference("//" + image.String())
	if err != nil {
		return err
	}
	destRef, err := docker.ParseReference("//" + destImage.String())
	if err != nil {
		return err
	}

	sourceCtx := &types.SystemContext{DockerAuthConfig: &types.DockerAuthConfig{}}
	destinationCtx := &types.SystemContext{DockerAuthConfig: &types.DockerAuthConfig{
		Username: opt.User,
		Password: opt.Password,
	}}

	_, err = copy.Image(ctx, policyContext, destRef, srcRef, &copy.Options{
		SourceCtx:          sourceCtx,
		DestinationCtx:     destinationCtx,
		ImageListSelection: copy.CopyAllImages,
	})
	return err
}

func getImageTags(imageName string, opt TagsOption) ([]string, error) {
	srcRef, err := docker.ParseReference("//" + imageName)
	if err != nil {
		return nil, err
	}
	sourceCtx := &types.SystemContext{DockerAuthConfig: &types.DockerAuthConfig{}}
	tagsCtx, tagsCancel := context.WithTimeout(context.Background(), opt.Timeout)
	defer tagsCancel()
	return docker.GetRepositoryTags(tagsCtx, sourceCtx, srcRef)
}

func checkSync(image Image) (manifest.Manifest, manifest.List, bool) {
	m, l, err := getImageManifest(image.String())
	if err != nil {
		logrus.Errorf("failed to get image [%s] manifest, error: %s", image.String(), err)
		return nil, nil, false
	}
	sm, ok := manifestsMap[image.String()]
	if (ok && m != nil && reflect.DeepEqual(m, sm)) || (ok && l != nil && reflect.DeepEqual(l, sm)) {
		logrus.Warnf("image [%s] not changed, skip sync...", image.String())
		return nil, nil, false
	}
	return m, l, true
}

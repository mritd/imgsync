package core

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"text/template"
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
	User                  string        // Docker Hub User
	Password              string        // Docker Hub User Password
	Timeout               time.Duration // Sync single image timeout
	Limit                 int           // Images sync process limit
	BatchSize             int           // Batch size for batch synchronization
	BatchNumber           int           // Sync specified batch
	OnlyDownloadManifests bool          // Only download Manifests file
	Report                bool          // Report sync result
	ReportLevel           int           // Report level
	ReportFile            string        // Report file

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
	case "kNative":
		return &kNative
	default:
		logrus.Fatalf("failed to create synchronizer %s: unknown synchronizer", name)
		// just for compiling
		return nil
	}
}

func SyncImages(ctx context.Context, images Images, opt *SyncOption) Images {
	imgs := batchProcess(images, opt)
	logrus.Infof("starting sync images, image total: %d", len(imgs))

	processWg := new(sync.WaitGroup)
	processWg.Add(len(imgs))

	if opt.Limit == 0 {
		opt.Limit = DefaultLimit
	}

	pool, err := ants.NewPool(opt.Limit, ants.WithPreAlloc(true), ants.WithPanicHandler(func(i interface{}) {
		logrus.Error(i)
	}))
	if err != nil {
		logrus.Fatalf("failed to create goroutines pool: %s", err)
	}
	sort.Sort(imgs)
	for i := 0; i < len(imgs); i++ {
		k := i
		err = pool.Submit(func() {
			defer processWg.Done()

			select {
			case <-ctx.Done():
			default:
				logrus.Debugf("process image: %s", imgs[k].String())
				m, l, needSync := checkSync(imgs[k])
				if !needSync {
					return
				}
				var bs []byte
				if m != nil {
					bs, err = jsoniter.MarshalIndent(m, "", "    ")
				} else {
					bs, err = jsoniter.MarshalIndent(l, "", "    ")
				}
				if err != nil {
					logrus.Errorf("failed to storage image [%s] manifests: %s", imgs[k].String(), err)
				}
				logrus.Debug(string(bs))

				rerr := retry(defaultSyncRetry, defaultSyncRetryTime, func() error {
					return sync2DockerHub(imgs[k], opt)
				})
				if rerr != nil {
					imgs[k].Err = rerr
					logrus.Errorf("failed to process image %s, error: %s", imgs[k].String(), rerr)
					return
				}
				imgs[k].Success = true

				storageDir := filepath.Join(ManifestDir, imgs[k].Repo, imgs[k].User, imgs[k].Name)
				// ignore other error
				if _, err = os.Stat(storageDir); err != nil {
					if err = os.MkdirAll(storageDir, 0755); err != nil {
						logrus.Errorf("failed to storage image [%s] manifests: %s", imgs[k].String(), err)
					}
				}

				if ferr := ioutil.WriteFile(filepath.Join(storageDir, imgs[k].Tag+".json"), bs, 0644); ferr != nil {
					logrus.Errorf("failed to storage image [%s] manifests: %s", imgs[k].String(), ferr)
				}
			}
		})
		if err != nil {
			logrus.Fatalf("failed to submit task: %s", err)
		}
	}
	processWg.Wait()
	pool.Release()
	return imgs
}

func sync2DockerHub(image *Image, opt *SyncOption) error {
	if opt.OnlyDownloadManifests {
		return nil
	}
	destImage := Image{
		Repo: defaultDockerRepo,
		User: opt.User,
		Name: image.MergeName(),
		Tag:  image.Tag,
	}

	logrus.Infof("syncing %s => %s", image.String(), destImage.String())

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

	logrus.Debugf("copy %s to docker hub...", image.String())
	_, err = copy.Image(ctx, policyContext, destRef, srcRef, &copy.Options{
		SourceCtx:          sourceCtx,
		DestinationCtx:     destinationCtx,
		ImageListSelection: copy.CopyAllImages,
	})
	logrus.Debugf("%s copy done.", image.String())
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

func checkSync(image *Image) (manifest.Manifest, manifest.List, bool) {
	var m manifest.Manifest
	var l manifest.List
	var merr error

	err := retry(DefaultGoRequestRetry, DefaultGoRequestRetryTime, func() error {
		m, l, merr = getImageManifest(image.String())
		if merr != nil {
			return merr
		}
		return nil
	})

	if err != nil {
		image.Err = err
		logrus.Errorf("failed to get image [%s] manifest, error: %s", image.String(), err)
		return nil, nil, false
	}
	val, ok := manifestsMap[image.String()]
	if (ok && m != nil && reflect.DeepEqual(m, val)) || (ok && l != nil && reflect.DeepEqual(l, val)) {
		image.Success = true
		image.CacheHit = true
		logrus.Debugf("image [%s] not changed, skip sync...", image.String())
		return nil, nil, false
	}
	return m, l, true
}

func batchProcess(images Images, opt *SyncOption) Images {
	if opt.BatchSize > 0 && opt.BatchNumber > 0 && len(images) > opt.BatchSize {
		n := len(images) / opt.BatchSize
		var start, end int
		if opt.BatchNumber >= n {
			start = opt.BatchSize * (n - 1)
			return images[start:]
		}
		start = opt.BatchSize * (opt.BatchNumber - 1)
		end = opt.BatchSize * opt.BatchNumber
		return images[start:end]
	}
	return images
}

func report(images Images, opt *SyncOption) {
	if !opt.Report {
		return
	}
	var successCount, failedCount, cacheHitCount int
	var report string

	for _, img := range images {
		if img.Success {
			successCount++
			if img.CacheHit {
				cacheHitCount++
			}
		} else {
			failedCount++
		}
	}
	report = fmt.Sprintf(reportHeaderTpl, Banner, len(images), failedCount, successCount, cacheHitCount)

	if opt.ReportLevel > 1 {
		var buf bytes.Buffer
		reportError, _ := template.New("").Parse(reportErrorTpl)
		err := reportError.Execute(&buf, images)
		if err != nil {
			logrus.Errorf("failed to create report error: %s", err)
		}
		report += buf.String()
	}

	if opt.ReportLevel > 2 {
		var buf bytes.Buffer
		reportSuccess, _ := template.New("").Parse(reportSuccessTpl)
		err := reportSuccess.Execute(&buf, images)
		if err != nil {
			logrus.Errorf("failed to create report error: %s", err)
		}
		report += buf.String()
	}
	fmt.Println(report)
	if opt.ReportFile != "" {
		err := ioutil.WriteFile(opt.ReportFile, []byte(report), 0644)
		if err != nil {
			logrus.Errorf("failed to create report file: %s", err)
		}
	}
}

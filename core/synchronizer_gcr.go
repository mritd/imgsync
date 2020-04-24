package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/panjf2000/ants/v2"

	"github.com/parnurzeal/gorequest"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

var gcr Gcr

type Gcr struct {
	kubeadm    bool
	queryLimit int
	namespace  string
}

func (gcr *Gcr) Images(ctx context.Context) Images {
	publicImageNames := gcr.imageNames()

	logrus.Info("get gcr public image tags...")
	pool, err := ants.NewPool(gcr.queryLimit+1, ants.WithPreAlloc(true), ants.WithPanicHandler(func(i interface{}) {
		logrus.Error(i)
	}))
	if err != nil {
		logrus.Fatalf("failed to create goroutines pool: %s", err)
	}

	var images Images
	imgCh := make(chan Image, gcr.queryLimit)
	err = pool.Submit(func() {
		for image := range imgCh {
			images = append(images, image)
		}
	})
	if err != nil {
		logrus.Fatalf("failed to submit task: %s", err)
	}

	imgGetWg := new(sync.WaitGroup)
	imgGetWg.Add(len(publicImageNames))
	for _, tmpImageName := range publicImageNames {
		imageName := tmpImageName
		err = pool.Submit(func() {
			defer imgGetWg.Done()
			select {
			case <-ctx.Done():
			default:
				var iName string
				if gcr.kubeadm {
					iName = fmt.Sprintf("%s/%s/%s", DefaultGcrRepo, DefaultGcrNamespace, imageName)
				} else {
					iName = fmt.Sprintf("%s/%s/%s", DefaultGcrRepo, gcr.namespace, imageName)
				}

				logrus.Debugf("query image [%s] tags...", iName)
				tags, terr := getImageTags(iName, TagsOption{Timeout: DefaultCtxTimeout})
				if err != nil {
					logrus.Errorf("failed to get image [%s] tags, error: %s", iName, terr)
					return
				}
				logrus.Debugf("image [%s] tags count: %d", iName, len(tags))

				for _, tag := range tags {
					if gcr.kubeadm {
						imgCh <- Image{
							Repo: DefaultK8sRepo,
							Name: imageName,
							Tag:  tag,
						}
					} else {
						imgCh <- Image{
							Repo: DefaultGcrRepo,
							User: gcr.namespace,
							Name: imageName,
							Tag:  tag,
						}
					}
				}
			}
		})
		if err != nil {
			logrus.Fatalf("failed to submit task: %s", err)
		}
	}

	imgGetWg.Wait()
	pool.Release()
	close(imgCh)
	return images
}

func (gcr *Gcr) imageNames() []string {
	logrus.Info("get gcr public images...")

	var addr string
	if gcr.kubeadm {
		addr = GcrKubeadmImagesTpl
	} else {
		addr = fmt.Sprintf(GcrStandardImagesTpl, gcr.namespace)
	}

	resp, body, errs := gorequest.New().
		Timeout(DefaultHTTPTimeout).
		Retry(DefaultGoRequestRetry, DefaultGoRequestRetryTime).
		Get(addr).
		EndBytes()
	if errs != nil {
		logrus.Fatalf("failed to get gcr images, address: %s, error: %s", addr, errs)
	}
	defer func() { _ = resp.Body.Close() }()

	var imageNames []string
	err := jsoniter.UnmarshalFromString(jsoniter.Get(body, "child").ToString(), &imageNames)
	if err != nil {
		logrus.Fatalf("failed to get gcr images, address: %s, error: %s", addr, err)
	}
	return imageNames
}

func (gcr *Gcr) Sync(ctx context.Context, opt *SyncOption) {
	gcrImages := gcr.setDefault(opt).Images(ctx)
	logrus.Infof("sync images count: %d", len(gcrImages))
	SyncImages(ctx, gcrImages, opt)
}

func (gcr *Gcr) setDefault(opt *SyncOption) *Gcr {
	gcr.kubeadm = opt.Kubeadm
	if opt.QueryLimit == 0 {
		gcr.queryLimit = 20
	} else {
		gcr.queryLimit = opt.QueryLimit
	}
	gcr.namespace = opt.NameSpace
	return gcr
}

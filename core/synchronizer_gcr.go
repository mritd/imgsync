package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/parnurzeal/gorequest"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

var gcr Gcr

type Gcr struct {
	kubeadm      bool
	namespace    string
	queryLimitCh chan int
}

func (gcr *Gcr) Images() Images {
	publicImageNames := gcr.imageNames()

	logrus.Info("get gcr public image tags...")

	imgCh := make(chan Image, DefaultLimit)
	imgGetWg := new(sync.WaitGroup)
	imgGetWg.Add(len(publicImageNames))

	for _, imageName := range publicImageNames {
		go func(imageName string) {
			defer func() {
				<-gcr.queryLimitCh
				imgGetWg.Done()
			}()

			var iName string
			if gcr.kubeadm {
				iName = fmt.Sprintf("%s/%s/%s", DefaultGcrRepo, DefaultGcrNamespace, imageName)
			} else {
				iName = fmt.Sprintf("%s/%s/%s", DefaultGcrRepo, gcr.namespace, imageName)
			}

			gcr.queryLimitCh <- 1

			logrus.Debugf("query image [%s] tags...", iName)
			tags, err := getImageTags(iName, TagsOption{Timeout: DefaultCtxTimeout})
			if err != nil {
				logrus.Errorf("failed to get image [%s] tags, error: %s", iName, err)
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
		}(imageName)
	}

	var images Images
	go func() {
		for image := range imgCh {
			images = append(images, image)
		}
	}()

	imgGetWg.Wait()
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
	gcrImages := gcr.setDefault(opt).Images()
	logrus.Infof("sync images count: %d", len(gcrImages))
	syncImages(ctx, gcrImages, opt)
}

func (gcr *Gcr) setDefault(opt *SyncOption) *Gcr {
	gcr.kubeadm = opt.Kubeadm
	gcr.queryLimitCh = make(chan int, opt.QueryLimit)
	gcr.namespace = opt.NameSpace
	return gcr
}

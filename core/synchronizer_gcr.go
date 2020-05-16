package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/panjf2000/ants/v2"
	"github.com/parnurzeal/gorequest"

	"github.com/sirupsen/logrus"
)

var gcr Gcr

type Gcr struct {
	kubeadm    bool
	queryLimit int
	namespace  string
}

func (gcr *Gcr) Images(ctx context.Context) Images {
	logrus.Info("get gcr public images...")
	if gcr.kubeadm {
		gcr.namespace = "google-containers"
	}

	var images Images
	imageNames := gcrImagesQuery(ctx, gcr.namespace, gcr.queryLimit, 5)
	for _, n := range imageNames {
		ss := strings.Split(n, ":")
		if len(ss) != 2 {
			logrus.Errorf("image name format error: %s", n)
			continue
		}
		if gcr.kubeadm {
			images = append(images, &Image{
				Repo: defaultK8sRepo,
				Name: strings.TrimPrefix(ss[0], gcr.namespace+"/"),
				Tag:  ss[1],
			})
		} else {
			images = append(images, &Image{
				Repo: defaultGcrRepo,
				User: gcr.namespace,
				Name: strings.TrimPrefix(ss[0], gcr.namespace+"/"),
				Tag:  ss[1],
			})
		}
	}
	return images
}

func (gcr *Gcr) Sync(ctx context.Context, opt *SyncOption) {
	gcrImages := gcr.setDefault(opt).Images(ctx)
	logrus.Infof("sync images count: %d", len(gcrImages))
	imgs := SyncImages(ctx, gcrImages, opt)
	report(imgs, opt)
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

func gcrImagesQuery(ctx context.Context, ns string, queryLimit, maxWait int) []string {
	nsCh := make(chan string, 1000)
	imagesCh := make(chan string, 1000)
	pool, _ := ants.NewPool(queryLimit, ants.WithPreAlloc(true), ants.WithPanicHandler(func(i interface{}) {
		logrus.Error(i)
	}))

	var images []string
	go func() {
		for n := range nsCh {
			gcrImagesBackgroundQuery(n, nsCh, imagesCh, pool)
		}
	}()
	go func() {
		for i := range imagesCh {
			images = append(images, i)
		}
	}()

	nsCh <- ns
	waitCount := 0
	tick := time.NewTicker(time.Second)
	for {
		select {
		case <-tick.C:
			logrus.Infof("gcr query pool running: %d, images count: %d", pool.Running(), len(images))
			if waitCount >= maxWait {
				goto done
			}
			if pool.Running() == 0 {
				waitCount++
			}
		case <-ctx.Done():
			goto done
		}
	}
done:
	close(nsCh)
	close(imagesCh)
	pool.Release()
	return images
}

func gcrImagesBackgroundQuery(ns string, nsCh, imagesCh chan<- string, pool *ants.Pool) {
	err := pool.Submit(func() {
		logrus.Debugf("query gcr ns: %s", ns)
		_, body, errs := gorequest.New().
			Timeout(DefaultHTTPTimeout).
			Retry(DefaultGoRequestRetry, DefaultGoRequestRetryTime).
			Get(fmt.Sprintf(gcrStandardImagesTpl, ns)).
			EndBytes()
		if len(errs) > 0 {
			logrus.Error(errs)
			return
		}

		var resp GcrResp
		err := jsoniter.Unmarshal(body, &resp)
		if err != nil {
			logrus.Error(err)
			return
		}
		if len(resp.Child) > 0 {
			for _, c := range resp.Child {
				nsCh <- ns + "/" + c
			}
		} else {
			for _, t := range resp.Tags {
				imagesCh <- resp.Name + ":" + t
			}
		}
	})
	if err != nil {
		logrus.Error(err)
	}
}

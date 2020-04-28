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

var kNativeImageAddrs = []string{
	"knative-releases/knative.dev/client/cmd",
	"knative-releases/knative.dev/eventing/cmd/broker",
	"knative-releases/knative.dev/eventing/cmd/in_memory",
	"knative-releases/knative.dev/eventing/cmd/mtbroker",
	"knative-releases/knative.dev/eventing/cmd/ping",
	"knative-releases/knative.dev/eventing/cmd/upgrade",
	"knative-releases/knative.dev/eventing/cmd",
	"knative-releases/knative.dev/eventing-contrib/awssqs/cmd",
	"knative-releases/knative.dev/eventing-contrib/camel/source/cmd",
	"knative-releases/knative.dev/eventing-contrib/cmd",
	"knative-releases/knative.dev/eventing-contrib/couchdb/source/cmd",
	"knative-releases/knative.dev/eventing-contrib/github/cmd",
	"knative-releases/knative.dev/eventing-contrib/gitlab/cmd",
	"knative-releases/knative.dev/eventing-contrib/kafka/channel/cmd",
	"knative-releases/knative.dev/eventing-contrib/kafka/source/cmd",
	"knative-releases/knative.dev/eventing-contrib/natss/cmd",
	"knative-releases/knative.dev/eventing-contrib/prometheus/cmd",
	"knative-releases/knative.dev/eventing-operator/cmd",
	"knative-releases/knative.dev/net-contour/cmd",
	"knative-releases/knative.dev/net-contour/vendor/github.com/projectcontour/contour/cmd",
	"knative-releases/knative.dev/net-http01/cmd",
	"knative-releases/knative.dev/net-istio/cmd",
	"knative-releases/knative.dev/net-kourier/cmd",
	"knative-releases/knative.dev/operator/cmd",
	"knative-releases/knative.dev/serving/cmd/networking",
	"knative-releases/knative.dev/serving/cmd",
	"knative-releases/knative.dev/serving/vendor/knative.dev/pkg/apiextensions/storageversion/cmd",
	"knative-releases/knative.dev/serving-operator/cmd",
}

var kNative KNative

type KNative struct {
	queryLimit int
	repo       string
}

func (kn *KNative) Images(ctx context.Context) Images {
	publicImageNames := kn.imageNames()

	logrus.Info("get knative public image tags...")
	pool, err := ants.NewPool(kn.queryLimit+1, ants.WithPreAlloc(true), ants.WithPanicHandler(func(i interface{}) {
		logrus.Error(i)
	}))
	if err != nil {
		logrus.Fatalf("failed to create goroutines pool: %s", err)
	}

	var images Images
	imgCh := make(chan Image, kn.queryLimit)
	err = pool.Submit(func() {
		for image := range imgCh {
			img := image
			images = append(images, &img)
		}
	})
	if err != nil {
		logrus.Fatalf("failed to submit task: %s", err)
	}

	imgGetWg := new(sync.WaitGroup)
	imgGetWg.Add(len(publicImageNames))
	for tmpImageName, ns := range publicImageNames {
		imageName := tmpImageName
		namespace := ns
		err = pool.Submit(func() {
			defer imgGetWg.Done()
			select {
			case <-ctx.Done():
			default:
				iName := fmt.Sprintf("%s/%s/%s", kn.repo, namespace, imageName)
				logrus.Debugf("query image [%s] tags...", iName)
				tags, terr := getImageTags(iName, TagsOption{Timeout: DefaultCtxTimeout})
				if err != nil {
					logrus.Errorf("failed to get image [%s] tags, error: %s", iName, terr)
					return
				}
				logrus.Debugf("image [%s] tags count: %d", iName, len(tags))

				for _, tag := range tags {
					imgCh <- Image{
						Repo: kn.repo,
						User: namespace,
						Name: imageName,
						Tag:  tag,
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

func (kn *KNative) imageNames() map[string]string {
	logrus.Info("get knative public images...")

	imageNames := make(map[string]string, 100)
	for _, addr := range kNativeImageAddrs {
		resp, body, errs := gorequest.New().
			Timeout(DefaultHTTPTimeout).
			Retry(DefaultGoRequestRetry, DefaultGoRequestRetryTime).
			Get(fmt.Sprintf(gcrStandardImagesTpl, addr)).
			EndBytes()
		if errs != nil {
			logrus.Errorf("failed to get knative images, address: %s, error: %s", addr, errs)
			continue
		}

		var names []string
		err := jsoniter.UnmarshalFromString(jsoniter.Get(body, "child").ToString(), &names)
		_ = resp.Body.Close()
		if err != nil {
			logrus.Errorf("failed to get knative images, address: %s, error: %s", addr, err)
			continue
		}
		for _, name := range names {
			imageNames[name] = addr
		}
	}
	return imageNames
}

func (kn *KNative) Sync(ctx context.Context, opt *SyncOption) {
	kNativeImages := kn.setDefault(opt).Images(ctx)
	logrus.Infof("sync images count: %d", len(kNativeImages))
	imgs := SyncImages(ctx, kNativeImages, opt)
	report(imgs, opt)
}

func (kn *KNative) setDefault(opt *SyncOption) *KNative {
	if opt.QueryLimit == 0 {
		kn.queryLimit = 20
	} else {
		kn.queryLimit = opt.QueryLimit
	}
	kn.repo = defaultGcrRepo
	return kn
}

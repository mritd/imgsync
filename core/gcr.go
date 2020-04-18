package core

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/parnurzeal/gorequest"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

const (
	GcrKubeadmImagesTpl     = "https://k8s.gcr.io/v2/tags/list"
	GcrStandardImagesTpl    = "https://gcr.io/v2/%s/tags/list"
	GcrKubeadmImageTagsTpl  = "https://k8s.gcr.io/v2/%s/tags/list"
	GcrStandardImageTagsTpl = "https://gcr.io/v2/%s/%s/tags/list"
)

type Gcr struct {
	Proxy             string
	Kubeadm           bool
	NameSpace         string
	DockerHubUser     string
	DockerHubPassword string
	SyncTimeOut       time.Duration
	HTTPTimeOut       time.Duration
	QueryLimit        int
	ProcessLimit      int
	queryLimitCh      chan int
	processLimitCh    chan int
}

// init gcr client
func (gcr *Gcr) Init() *Gcr {
	if gcr.DockerHubUser == "" || gcr.DockerHubPassword == "" {
		logrus.Fatal("docker hub user or password is empty")
	}

	if gcr.NameSpace == "" {
		gcr.NameSpace = "google-containers"
	}

	if gcr.SyncTimeOut == 0 {
		gcr.SyncTimeOut = DefaultSyncTimeout
	}

	if gcr.HTTPTimeOut == 0 {
		gcr.HTTPTimeOut = DefaultHTTPTimeOut
	}

	if gcr.QueryLimit == 0 {
		// query limit default 20
		gcr.queryLimitCh = make(chan int, DefaultLimit)
	} else {
		gcr.queryLimitCh = make(chan int, gcr.QueryLimit)
	}

	if gcr.ProcessLimit == 0 {
		// process limit default 20
		gcr.processLimitCh = make(chan int, DefaultLimit)
	} else {
		gcr.processLimitCh = make(chan int, gcr.ProcessLimit)
	}

	logrus.Info("init success...")

	return gcr
}

func (gcr *Gcr) Sync() {
	logrus.Info("starting sync gcr images...")

	gcrImages := gcr.images()
	sort.Sort(gcrImages)
	logrus.Infof("Google container registry images total: %d", len(gcrImages))

	ctx, cancel := context.WithTimeout(context.Background(), gcr.SyncTimeOut)
	defer cancel()

	processWg := new(sync.WaitGroup)
	processWg.Add(len(gcrImages))

	for _, image := range gcrImages {
		tmpImage := image
		go func() {
			defer func() {
				<-gcr.processLimitCh
				processWg.Done()
			}()
			select {
			case gcr.processLimitCh <- 1:
				process(tmpImage, gcr.DockerHubUser, gcr.DockerHubPassword)
			case <-ctx.Done():
			}
		}()
	}

	processWg.Wait()
}

func (gcr *Gcr) images() Images {
	publicImageNames := gcr.imageNames()

	logrus.Info("get gcr public image tags...")

	imgCh := make(chan Image, DefaultLimit)
	imgGetWg := new(sync.WaitGroup)
	imgGetWg.Add(len(publicImageNames))

	for _, imageName := range publicImageNames {
		tmpImageName := imageName
		go func() {
			defer func() {
				<-gcr.queryLimitCh
				imgGetWg.Done()
			}()

			gcr.queryLimitCh <- 1

			var addr string
			if gcr.Kubeadm {
				addr = fmt.Sprintf(GcrKubeadmImageTagsTpl, tmpImageName)
			} else {
				addr = fmt.Sprintf(GcrStandardImageTagsTpl, gcr.NameSpace, tmpImageName)
			}

			logrus.Debugf("get gcr image tags, address: %s", addr)
			resp, body, errs := gorequest.New().
				Proxy(gcr.Proxy).
				Timeout(gcr.HTTPTimeOut).
				Retry(DefaultGoRequestRetry, DefaultGoRequestRetryTime).
				Get(addr).
				EndBytes()
			if errs != nil {
				logrus.Errorf("failed to get gcr image tags, address: %s, error: %s", addr, errs)
				return
			}
			defer func() { _ = resp.Body.Close() }()

			var tags []string
			err := jsoniter.UnmarshalFromString(jsoniter.Get(body, "tags").ToString(), &tags)
			if err != nil {
				logrus.Errorf("failed to get gcr image tags, address: %s, error: %s", addr, err)
				return
			}

			for _, tag := range tags {
				if gcr.Kubeadm {
					imgCh <- Image{
						Repo: "k8s.gcr.io",
						Name: tmpImageName,
						Tag:  tag,
					}
				} else {
					imgCh <- Image{
						Repo: "gcr.io",
						User: gcr.NameSpace,
						Name: tmpImageName,
						Tag:  tag,
					}
				}
			}
		}()
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
	if gcr.Kubeadm {
		addr = GcrKubeadmImagesTpl
	} else {
		addr = fmt.Sprintf(GcrStandardImagesTpl, gcr.NameSpace)
	}

	resp, body, errs := gorequest.New().
		Proxy(gcr.Proxy).
		Timeout(gcr.HTTPTimeOut).
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

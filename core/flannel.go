package core

import (
	"context"
	"sort"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/parnurzeal/gorequest"

	"github.com/sirupsen/logrus"
)

const (
	// only sync the last 100 images
	FlannelImagesTpl = "https://quay.io/api/v1/repository/coreos/flannel/tag/?limit=100&page=1&onlyActiveTags=true"
)

type Flannel struct {
	Proxy             string
	DockerHubUser     string
	DockerHubPassword string
	SyncTimeOut       time.Duration
	HTTPTimeOut       time.Duration
	ProcessLimit      int
	processLimitCh    chan int
}

func (fl *Flannel) Init() *Flannel {
	if fl.DockerHubUser == "" || fl.DockerHubPassword == "" {
		logrus.Fatal("docker hub user or password is empty")
	}

	if fl.SyncTimeOut == 0 {
		fl.SyncTimeOut = DefaultSyncTimeout
	}

	if fl.HTTPTimeOut == 0 {
		fl.HTTPTimeOut = DefaultHTTPTimeOut
	}

	if fl.ProcessLimit == 0 {
		// process limit default 20
		fl.processLimitCh = make(chan int, DefaultLimit)
	} else {
		fl.processLimitCh = make(chan int, fl.ProcessLimit)
	}

	logrus.Info("init success...")

	return fl
}

func (fl *Flannel) Sync() {
	logrus.Info("starting sync flannel images...")

	flImages := fl.images()
	sort.Sort(flImages)
	logrus.Infof("Flannel images total: %d", len(flImages))

	ctx, cancel := context.WithTimeout(context.Background(), fl.SyncTimeOut)
	defer cancel()

	processWg := new(sync.WaitGroup)
	processWg.Add(len(flImages))

	for _, image := range flImages {
		tmpImage := image
		go func() {
			defer func() {
				<-fl.processLimitCh
				processWg.Done()
			}()
			select {
			case fl.processLimitCh <- 1:
				process(tmpImage, fl.DockerHubUser, fl.DockerHubPassword)
			case <-ctx.Done():
			}
		}()
	}

	processWg.Wait()
}

func (fl *Flannel) images() Images {
	logrus.Debugf("get flannel images, address: %s", FlannelImagesTpl)
	resp, body, errs := gorequest.New().
		Proxy(fl.Proxy).
		Timeout(fl.HTTPTimeOut).
		Retry(DefaultGoRequestRetry, DefaultGoRequestRetryTime).
		Get(FlannelImagesTpl).
		EndBytes()
	if errs != nil {
		logrus.Fatalf("failed to get flannel images, error: %s", errs)
	}
	defer func() { _ = resp.Body.Close() }()

	var tags []struct {
		Name string `json:"name"`
	}
	err := jsoniter.UnmarshalFromString(jsoniter.Get(body, "tags").ToString(), &tags)
	if err != nil {
		logrus.Fatalf("failed to get flannel image tags, address: %s, error: %s", FlannelImagesTpl, err)
	}

	var images Images
	for _, tag := range tags {
		images = append(images, Image{
			Repo: "quay.io",
			User: "coreos",
			Name: "flannel",
			Tag:  tag.Name,
		})
	}

	return images
}

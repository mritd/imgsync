package imgsync

import (
	"context"
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
	HttpTimeOut       time.Duration
	ProcessLimit      int
	processLimitCh    chan int
}

func (fl *Flannel) Init() *Flannel {

	if fl.DockerHubUser == "" || fl.DockerHubPassword == "" {
		logrus.Fatal("docker hub user or password is empty")
	}

	if fl.SyncTimeOut == 0 {
		fl.SyncTimeOut = 1 * time.Hour
	}

	if fl.HttpTimeOut == 0 {
		fl.HttpTimeOut = 5 * time.Second
	}

	if fl.ProcessLimit == 0 {
		// process limit default 20
		fl.processLimitCh = make(chan int, 20)
	} else {
		fl.processLimitCh = make(chan int, fl.ProcessLimit)
	}

	logrus.Infoln("init success...")

	return fl
}

func (fl *Flannel) Sync() {
	logrus.Info("starting sync flannel images...")
	images := fl.images()
	logrus.Infof("Flannel images total: %d", len(images))

	ctx, cancel := context.WithTimeout(context.Background(), fl.SyncTimeOut)
	defer cancel()

	processWg := new(sync.WaitGroup)
	processWg.Add(len(images))

	for _, image := range images {
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

func (fl *Flannel) images() []Image {
	logrus.Debugf("get flannel images, address: %s", FlannelImagesTpl)
	resp, body, errs := gorequest.New().
		Proxy(fl.Proxy).
		Timeout(fl.HttpTimeOut).
		Retry(3, 1*time.Second).
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

	var images []Image
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

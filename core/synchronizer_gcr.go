package core

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/parnurzeal/gorequest"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

var gcr Gcr

type Gcr struct {
	Proxy          string
	Kubeadm        bool
	NameSpace      string
	DockerUser     string
	DockerPassword string
	HTTPTimeOut    time.Duration
	QueryLimit     int

	queryLimitCh chan int
}

// init gcr client
func (gcr *Gcr) Default() error {
	if gcr.DockerUser == "" || gcr.DockerPassword == "" {
		return errors.New("docker hub user or password is empty")
	}

	if gcr.NameSpace == "" {
		gcr.NameSpace = "google-containers"
	}

	if gcr.HTTPTimeOut == 0 {
		gcr.HTTPTimeOut = DefaultHTTPTimeOut
	}

	logrus.Info("gcr init success...")
}

func (gcr *Gcr) Images() Images {
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
						Repo: DefaultK8sRepo,
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

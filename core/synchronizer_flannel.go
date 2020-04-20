package core

import (
	"errors"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/parnurzeal/gorequest"

	"github.com/sirupsen/logrus"
)

const (
	// only sync the last 100 images
	FlannelImagesTpl = "https://quay.io/api/v1/repository/coreos/flannel/tag/?limit=100&page=1&onlyActiveTags=true"
)

var fl Flannel

type Flannel struct {
	Proxy          string
	DockerUser     string
	DockerPassword string
	HTTPTimeOut    time.Duration
}

func (fl *Flannel) Default() error {
	if fl.DockerUser == "" || fl.DockerPassword == "" {
		return errors.New("docker hub user or password is empty")
	}

	if fl.HTTPTimeOut == 0 {
		fl.HTTPTimeOut = DefaultHTTPTimeOut
	}

	logrus.Info("flannel init success...")
}

func (fl *Flannel) Images() Images {
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

func (fl *Flannel) Sync() {

}

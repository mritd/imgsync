package core

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
)

var fl Flannel

type Flannel struct {
}

func (fl *Flannel) Images() Images {
	logrus.Infof("get flannel image tags")
	tags, err := getImageTags(FlannelImageName, TagsOption{Timeout: DefaultCtxTimeout})
	if err != nil {
		logrus.Errorf("failed to get [%s] image tags, error: %s", FlannelImageName, err)
		return nil
	}

	var images Images
	ss := strings.Split(FlannelImageName, "/")
	for _, tag := range tags {
		images = append(images, Image{
			Repo: ss[0],
			User: ss[1],
			Name: ss[2],
			Tag:  tag,
		})
	}

	return images
}

func (fl *Flannel) Sync(ctx context.Context, opt *SyncOption) {
	flImages := fl.setDefault(opt).Images()
	logrus.Infof("sync images count: %d", len(flImages))
	syncImages(ctx, flImages, opt)
}

func (fl *Flannel) setDefault(_ *SyncOption) *Flannel {
	return fl
}

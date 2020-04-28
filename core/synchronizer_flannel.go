package core

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
)

var fl Flannel

type Flannel struct {
}

func (fl *Flannel) Images(ctx context.Context) Images {
	logrus.Infof("get flannel image tags")
	var images Images
	select {
	case <-ctx.Done():
	default:
		tags, err := getImageTags(flannelImageName, TagsOption{Timeout: DefaultCtxTimeout})
		if err != nil {
			logrus.Errorf("failed to get [%s] image tags, error: %s", flannelImageName, err)
			return nil
		}

		ss := strings.Split(flannelImageName, "/")
		for _, tag := range tags {
			images = append(images, &Image{
				Repo: ss[0],
				User: ss[1],
				Name: ss[2],
				Tag:  tag,
			})
		}
	}

	return images
}

func (fl *Flannel) Sync(ctx context.Context, opt *SyncOption) {
	flImages := fl.setDefault(opt).Images(ctx)
	logrus.Infof("sync images count: %d", len(flImages))
	imgs := SyncImages(ctx, flImages, opt)
	report(imgs, opt)
}

func (fl *Flannel) setDefault(_ *SyncOption) *Flannel {
	return fl
}

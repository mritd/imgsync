package imgsync

import (
	"context"
	"fmt"
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
	Kubeadm           bool
	NameSpace         string
	DockerHubUser     string
	DockerHubPassword string
	SyncTimeOut       time.Duration
	HttpTimeOut       time.Duration
	QueryLimit        int
	ProcessLimit      int
	queryLimitCh      chan int
	processLimitCh    chan int
}

// init gcr client
func (g *Gcr) Init() *Gcr {

	if g.NameSpace == "" {
		g.NameSpace = "google-containers"
	}

	if g.SyncTimeOut == 0 {
		g.SyncTimeOut = 1 * time.Hour
	}

	if g.HttpTimeOut == 0 {
		g.HttpTimeOut = 5 * time.Second
	}

	if g.QueryLimit == 0 {
		// query limit default 20
		g.queryLimitCh = make(chan int, 20)
	} else {
		g.queryLimitCh = make(chan int, g.QueryLimit)
	}

	if g.ProcessLimit == 0 {
		// process limit default 20
		g.processLimitCh = make(chan int, 20)
	} else {
		g.processLimitCh = make(chan int, g.ProcessLimit)
	}

	if g.DockerHubUser == "" || g.DockerHubPassword == "" {
		logrus.Fatal("docker hub user or password is empty")
	}

	logrus.Infoln("init success...")

	return g
}

func (g *Gcr) Sync() {

	logrus.Info("starting sync gcr images...")

	gcrImages := g.gcrImageList()
	logrus.Infof("Google container registry images total: %d", len(gcrImages))

	ctx, cancel := context.WithTimeout(context.Background(), g.SyncTimeOut)
	defer cancel()

	processWg := new(sync.WaitGroup)
	processWg.Add(len(gcrImages))

	for _, image := range gcrImages {
		tmpImage := image
		go func() {
			defer func() {
				<-g.processLimitCh
				processWg.Done()
			}()
			select {
			case g.processLimitCh <- 1:
				g.process(tmpImage)
			case <-ctx.Done():
			}
		}()
	}

	processWg.Wait()

}

func (g *Gcr) gcrImageList() []Image {

	publicImageNames := g.gcrPublicImageNames()

	logrus.Info("get gcr public image tags...")

	imgCh := make(chan Image, 20)
	imgGetWg := new(sync.WaitGroup)
	imgGetWg.Add(len(publicImageNames))

	for _, imageName := range publicImageNames {
		tmpImageName := imageName
		go func() {
			defer func() {
				<-g.queryLimitCh
				imgGetWg.Done()
			}()

			g.queryLimitCh <- 1

			var addr string
			if g.Kubeadm {
				addr = fmt.Sprintf(GcrKubeadmImageTagsTpl, tmpImageName)
			} else {
				addr = fmt.Sprintf(GcrStandardImageTagsTpl, g.NameSpace, tmpImageName)
			}

			logrus.Debugf("get gcr image tags, address: %s", addr)
			resp, body, errs := gorequest.New().
				Timeout(g.HttpTimeOut).
				Retry(3, 1*time.Second).
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
				if g.Kubeadm {
					imgCh <- Image{
						Repo: "k8s.gcr.io",
						Name: tmpImageName,
						Tag:  tag,
					}
				} else {
					imgCh <- Image{
						Repo: "gcr.io",
						User: g.NameSpace,
						Name: tmpImageName,
						Tag:  tag,
					}
				}

			}

		}()
	}

	var images []Image
	go func() {
		for {
			select {
			case image, ok := <-imgCh:
				if ok {
					images = append(images, image)
				} else {
					break
				}
			}
		}
	}()

	imgGetWg.Wait()
	close(imgCh)
	return images
}

func (g *Gcr) gcrPublicImageNames() []string {

	logrus.Info("get gcr public images...")

	var addr string
	if g.Kubeadm {
		addr = GcrKubeadmImagesTpl
	} else {
		addr = fmt.Sprintf(GcrStandardImagesTpl, g.NameSpace)
	}

	resp, body, errs := gorequest.New().
		Timeout(g.HttpTimeOut).
		Retry(3, 1*time.Second).
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

func (g *Gcr) process(image Image) {
	logrus.Debugf("process image: %s", image)
	err := syncDockerHub(image, DockerHubOption{
		Username: g.DockerHubUser,
		Password: g.DockerHubPassword,
		Timeout:  10 * time.Minute,
	})
	if err != nil {
		logrus.Errorf("failed to process image %s, error: %s", image, err)
	}
}

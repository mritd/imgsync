package gcrsync

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/mritd/gcrsync/utils"
)

const (
	ChangeLog      = "CHANGELOG-%s.md"
	GcrRegistryTpl = "gcr.io/%s/%s"
	GcrImages      = "https://gcr.io/v2/%s/tags/list"
	GcrImageTags   = "https://gcr.io/v2/%s/%s/tags/list"
	DockerHubImage = "https://hub.docker.com/v2/repositories/%s/?page_size=100"
	DockerHubTags  = "https://hub.docker.com/v2/repositories/%s/%s/tags/?page_size=100"
)

func (g *Gcr) Sync() {

	gcrImages := g.gcrImageList()

	logrus.Infof("Google container registry images total: %d", len(gcrImages))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if g.SyncTimeOut != 0 {
			select {
			case <-time.After(g.SyncTimeOut):
				cancel()
			}
		}
	}()

	processWg := new(sync.WaitGroup)
	processWg.Add(len(gcrImages))

	for _, imageName := range gcrImages {
		tmpImageName := imageName
		go func() {
			defer func() {
				g.ProcessLimit <- 1
				processWg.Done()
			}()
			select {
			case <-g.ProcessLimit:
				g.Process(tmpImageName)
			case <-ctx.Done():
			}
		}()
	}

	// doc gen
	chgWg := new(sync.WaitGroup)
	chgWg.Add(1)
	go func() {
		defer chgWg.Done()

		var images []string
		for {
			select {
			case imageName, ok := <-g.update:
				if ok {
					images = append(images, imageName)
				} else {
					goto ChangeLogDone
				}
			case <-ctx.Done():
				goto ChangeLogDone
			}
		}
	ChangeLogDone:
		if len(images) > 0 {
			g.Commit(images)
		}
	}()

	processWg.Wait()
	close(g.update)
	chgWg.Wait()

}

func (g *Gcr) Monitor() {

	if g.MonitorCount == -1 {
		for {
			select {
			case <-time.Tick(5 * time.Second):
				gcrImages := g.gcrImageList()
				dockerHubImages := g.dockerHubImageList()
				needSyncImages := utils.SliceDiff(gcrImages, dockerHubImages)
				logrus.Infof("Gcr images: %d | Docker hub images: %d | Waiting process: %d", len(gcrImages), len(dockerHubImages), len(needSyncImages))
			}
		}
	} else {
		for i := 0; i < g.MonitorCount; i++ {
			select {
			case <-time.Tick(5 * time.Second):
				gcrImages := g.gcrImageList()
				dockerHubImages := g.dockerHubImageList()
				needSyncImages := utils.SliceDiff(gcrImages, dockerHubImages)
				logrus.Infof("Gcr images: %d | Docker hub images: %d | Waiting process: %d", len(gcrImages), len(dockerHubImages), len(needSyncImages))
			}
		}
	}

}

func (g *Gcr) Init() {

	logrus.Infoln("init http client.")
	g.httpClient = &http.Client{
		Timeout: g.HttpTimeOut,
	}
	if g.Proxy != "" {
		g.httpClient.Transport = &http.Transport{
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return url.Parse(g.Proxy)
			},
		}
	}

	logrus.Infoln("init limit channel.")
	for i := 0; i < cap(g.QueryLimit); i++ {
		g.QueryLimit <- 1
	}
	for i := 0; i < cap(g.ProcessLimit); i++ {
		g.ProcessLimit <- 1
	}

	logrus.Infoln("init update channel.")
	g.update = make(chan string, 20)

	logrus.Infoln("Init success...")
}

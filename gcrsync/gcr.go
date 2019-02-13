package gcrsync

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/docker/docker/client"

	"github.com/json-iterator/go"
	"github.com/mritd/gcrsync/utils"
	"github.com/sirupsen/logrus"
)

type Image struct {
	Name string
	Tags []string
}

type Gcr struct {
	Proxy          string
	DockerUser     string
	DockerPassword string
	NameSpace      string
	GithubToken    string
	GithubRepo     string
	CommitMsg      string
	MonitorCount   int
	TestMode       bool
	Debug          bool
	SyncTimeOut    time.Duration
	QueryLimit     chan int
	ProcessLimit   chan int
	HttpTimeOut    time.Duration
	httpClient     *http.Client
	dockerClient   *client.Client
	dockerHubToken string
	update         chan string
	commitURL      string
}

func (g *Gcr) gcrImageList() []string {

	var images []string
	publicImageNames := g.gcrPublicImageNames()

	logrus.Debugf("Number of gcr images: %d", len(publicImageNames))

	imgNameCh := make(chan string, 20)
	imgGetWg := new(sync.WaitGroup)
	imgGetWg.Add(len(publicImageNames))

	for _, imageName := range publicImageNames {

		tmpImageName := imageName

		go func() {
			defer func() {
				g.QueryLimit <- 1
				imgGetWg.Done()
			}()

			select {
			case <-g.QueryLimit:
				req, err := http.NewRequest("GET", fmt.Sprintf(GcrImageTags, g.NameSpace, tmpImageName), nil)
				utils.CheckAndExit(err)

				resp, err := g.httpClient.Do(req)
				utils.CheckAndExit(err)

				b, err := ioutil.ReadAll(resp.Body)
				utils.CheckAndExit(err)
				_ = resp.Body.Close()

				var tags []string
				_ = jsoniter.UnmarshalFromString(jsoniter.Get(b, "tags").ToString(), &tags)

				for _, tag := range tags {
					imgNameCh <- tmpImageName + ":" + tag
				}
			}

		}()
	}

	var imgReceiveWg sync.WaitGroup
	imgReceiveWg.Add(1)
	go func() {
		defer imgReceiveWg.Done()
		for {
			select {
			case imageName, ok := <-imgNameCh:
				if ok {
					images = append(images, imageName)
				} else {
					goto imgSetExit
				}
			}
		}
	imgSetExit:
	}()

	imgGetWg.Wait()
	close(imgNameCh)
	imgReceiveWg.Wait()
	return images
}

func (g *Gcr) gcrPublicImageNames() []string {

	req, err := http.NewRequest("GET", fmt.Sprintf(GcrImages, g.NameSpace), nil)
	utils.CheckAndExit(err)

	resp, err := g.httpClient.Do(req)
	utils.CheckAndExit(err)
	defer func() {
		_ = resp.Body.Close()
	}()

	b, err := ioutil.ReadAll(resp.Body)
	utils.CheckAndExit(err)

	var imageNames []string
	_ = jsoniter.UnmarshalFromString(jsoniter.Get(b, "child").ToString(), &imageNames)
	return imageNames
}

package gcrsync

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/docker/docker/client"

	"github.com/Sirupsen/logrus"
	"github.com/json-iterator/go"
	"github.com/mritd/gcrsync/pkg/utils"
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
	TestMode       bool
	QueryLimit     chan int
	ProcessLimit   chan int
	httpClient     *http.Client
	dockerClient   *client.Client
	dockerHubToken string
	update         chan string
}

func (g *Gcr) gcrImageList() map[string]bool {

	images := make(map[string]bool)
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
				resp.Body.Close()

				var tags []string
				jsoniter.UnmarshalFromString(jsoniter.Get(b, "tags").ToString(), &tags)

				for _, tag := range tags {
					imgNameCh <- tmpImageName + ":" + tag
				}
			}

		}()
	}

	go func() {
		for {
			select {
			case imageName, ok := <-imgNameCh:
				if ok {
					images[imageName] = true
				} else {
					goto imgSetExit
				}
			}
		}
	imgSetExit:
	}()

	imgGetWg.Wait()
	close(imgNameCh)

	return images
}

func (g *Gcr) gcrPublicImageNames() []string {

	req, err := http.NewRequest("GET", fmt.Sprintf(GcrImages, g.NameSpace), nil)
	utils.CheckAndExit(err)

	resp, err := g.httpClient.Do(req)
	utils.CheckAndExit(err)
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	utils.CheckAndExit(err)

	var imageNames []string
	jsoniter.UnmarshalFromString(jsoniter.Get(b, "child").ToString(), &imageNames)
	return imageNames
}

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
	Proxy           string
	DockerUser      string
	DockerPassword  string
	NameSpace       string
	TestMode        bool
	httpClient      *http.Client
	dockerClient    *client.Client
	dockerHubToken  string
	dockerHubImages map[string]bool
	update          chan string
}

func (g *Gcr) gcrImageList() map[string]bool {

	images := make(map[string]bool)
	publicImageNames := g.gcrPublicImageNames()

	imgGetWg := new(sync.WaitGroup)
	imgGetWg.Add(len(publicImageNames))
	imgNameCh := make(chan string, 20)

	for _, imageName := range publicImageNames {
		tmpImgName := imageName
		go func() {
			defer imgGetWg.Done()

			req, err := http.NewRequest("GET", fmt.Sprintf(GcrImageTags, g.NameSpace, tmpImgName), nil)
			utils.CheckAndExit(err)

			resp, err := g.httpClient.Do(req)
			utils.CheckAndExit(err)

			b, err := ioutil.ReadAll(resp.Body)
			utils.CheckAndExit(err)
			resp.Body.Close()

			var tags []string
			jsoniter.UnmarshalFromString(jsoniter.Get(b, "tags").ToString(), &tags)
			logrus.Debugf("Found image [%s] tags: %s", tmpImgName, tags)

			for _, tag := range tags {
				imgNameCh <- tmpImgName + ":" + tag
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

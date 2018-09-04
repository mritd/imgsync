package gcrsync

import (
	"fmt"
	"io/ioutil"
	"net/http"

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

func (g *Gcr) gcrImageList() []Image {

	var images []Image

	publicImages := g.gcrPublicImages()
	for _, imageName := range publicImages {
		req, err := http.NewRequest("GET", fmt.Sprintf(GcrImageTags, g.NameSpace, imageName), nil)
		utils.CheckAndExit(err)

		resp, err := g.httpClient.Do(req)
		utils.CheckAndExit(err)

		b, err := ioutil.ReadAll(resp.Body)
		utils.CheckAndExit(err)
		resp.Body.Close()

		var tags []string
		jsoniter.UnmarshalFromString(jsoniter.Get(b, "tags").ToString(), &tags)

		logrus.Debugf("Found image [%s] tags: %s", imageName, tags)

		images = append(images, Image{
			Name: fmt.Sprintf(GcrRegistryPrefix, g.NameSpace) + imageName,
			Tags: tags,
		})

	}
	return images
}

func (g *Gcr) gcrPublicImages() []string {

	req, err := http.NewRequest("GET", fmt.Sprintf(GcrImages, g.NameSpace), nil)
	utils.CheckAndExit(err)

	resp, err := g.httpClient.Do(req)
	utils.CheckAndExit(err)
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	utils.CheckAndExit(err)

	var images []string
	jsoniter.UnmarshalFromString(jsoniter.Get(b, "child").ToString(), &images)
	return images
}

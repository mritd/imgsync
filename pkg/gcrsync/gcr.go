package gcrsync

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

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
	Prefix          string
	DockerUser      string
	DockerPassword  string
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
		req, err := http.NewRequest("GET", fmt.Sprintf(GcrImageTags, imageName), nil)
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
			Name: GcrRegistryPrefix + imageName,
			Tags: tags,
		})

	}
	return images
}

func (g *Gcr) gcrPublicImages() []string {

	req, err := http.NewRequest("GET", GcrImages, nil)
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

func (g *Gcr) Init() {
	logrus.Debugln("Init google container client.")
	var httpClient *http.Client
	if g.Proxy != "" {
		p := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(g.Proxy)
		}
		transport := &http.Transport{Proxy: p}
		httpClient = &http.Client{
			Timeout:   5 * time.Second,
			Transport: transport,
		}
	} else {
		httpClient = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	g.httpClient = httpClient

	logrus.Debugln("Init docker client.")
	dockerClient, err := client.NewEnvClient()
	utils.CheckAndExit(err)
	g.dockerClient = dockerClient

	logrus.Debugln("Init docker hub.")
	g.hubImages()

	logrus.Debugln("Init update channel.")
	g.update = make(chan string, 10)

	logrus.Debugln("Init success...")
}

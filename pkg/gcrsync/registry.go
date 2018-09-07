package gcrsync

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/json-iterator/go"

	"github.com/Sirupsen/logrus"

	"github.com/mritd/gcrsync/pkg/utils"
)

func (g *Gcr) hubToken() {
	req, err := http.NewRequest("POST", HubLogin, bytes.NewBufferString(`{"username": "`+g.DockerUser+`", "password": "`+g.DockerPassword+`"}`))
	utils.CheckAndExit(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	utils.CheckAndExit(err)
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	utils.CheckAndExit(err)

	token := jsoniter.Get(b, "token").ToString()
	if strings.TrimSpace(token) == "" {
		utils.ErrorExit("Failed to get docker hub token", 1)
	}
	g.dockerHubToken = token
}

func (g *Gcr) regImageList() []string {

	if g.dockerHubToken == "" {
		g.hubToken()
	}

	publicImageNames := g.regPublicImageNames()

	logrus.Debugf("Number of registry images: %d", len(publicImageNames))

	return g.regPublicImageTags(publicImageNames)

}

func (g *Gcr) regPublicImageTags(imageNames []string) []string {

	var images []string
	imgNameCh := make(chan string, 20)
	imgGetWg := new(sync.WaitGroup)
	imgGetWg.Add(len(imageNames))

	for _, imageName := range imageNames {

		tmpImageName := imageName

		go func() {
			defer func() {
				g.QueryLimit <- 1
				imgGetWg.Done()
			}()

			select {
			case <-g.QueryLimit:
				var imageTags []string
				g.requestRegistryImageTags(fmt.Sprintf(HubTags, g.DockerUser, tmpImageName), &imageTags)
				for _, tag := range imageTags {
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
	return images
}

func (g *Gcr) requestRegistryImageTags(addr string, imageTags *[]string) int {

	logrus.Debugf("Registry request: %s", addr)

	if g.dockerHubToken == "" {
		g.hubToken()
	}

	req := g.buildRegistryRequest(addr)
	resp, err := g.httpClient.Do(req)
	utils.CheckAndExit(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.Warningf("Failed to request: %s", addr)
		return 0
	}

	var result struct {
		Count   int
		Next    string
		Results []struct {
			Name string
		}
	}
	b, err := ioutil.ReadAll(resp.Body)
	utils.CheckAndExit(err)
	jsoniter.Unmarshal(b, &result)

	for _, repo := range result.Results {
		*imageTags = append(*imageTags, repo.Name)
	}

	if strings.TrimSpace(result.Next) != "" {
		g.requestRegistryImageTags(result.Next, imageTags)
	}
	return result.Count
}

func (g *Gcr) regPublicImageNames() []string {
	var imageNames []string
	g.requestRegistryImageNames(fmt.Sprintf(HubRepos, g.DockerUser), &imageNames)
	return imageNames
}

func (g *Gcr) requestRegistryImageNames(addr string, imageNames *[]string) {

	logrus.Debugf("Registry request: %s", addr)

	if g.dockerHubToken == "" {
		g.hubToken()
	}

	req := g.buildRegistryRequest(addr)
	resp, err := g.httpClient.Do(req)
	utils.CheckAndExit(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.Warningf("Failed to request: %s", addr)
		return
	}

	var result struct {
		Count   int
		Next    string
		Results []struct {
			User string
			Name string
		}
	}
	b, err := ioutil.ReadAll(resp.Body)
	utils.CheckAndExit(err)
	jsoniter.Unmarshal(b, &result)

	for _, repo := range result.Results {
		*imageNames = append(*imageNames, repo.Name)
	}

	if strings.TrimSpace(result.Next) != "" {
		g.requestRegistryImageNames(result.Next, imageNames)
	}

}

func (g *Gcr) buildRegistryRequest(addr string) *http.Request {
	req, err := http.NewRequest("GET", addr, nil)
	utils.CheckAndExit(err)
	req.Header.Set("Authorization", "JWT "+g.dockerHubToken)
	return req
}

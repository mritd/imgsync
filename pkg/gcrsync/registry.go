package gcrsync

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/json-iterator/go"
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

	var images []string
	publicImageNames := g.regPublicImageNames()

	imgGetWg := new(sync.WaitGroup)
	imgGetWg.Add(len(publicImageNames))
	imgNameCh := make(chan string, 20)

	for _, imageName := range publicImageNames {
		tmpImgName := imageName
		go func() {
			defer imgGetWg.Done()

			req, err := http.NewRequest("GET", fmt.Sprintf(HubTags, g.DockerUser, tmpImgName), nil)
			utils.CheckAndExit(err)
			req.Header.Set("Authorization", "JWT "+g.dockerHubToken)

			resp, err := g.httpClient.Do(req)
			utils.CheckAndExit(err)
			defer resp.Body.Close()

			var result struct {
				Results []struct {
					Name string
				}
			}

			b, err := ioutil.ReadAll(resp.Body)
			utils.CheckAndExit(err)
			jsoniter.Unmarshal(b, &result)

			for _, tag := range result.Results {
				imgNameCh <- tmpImgName + ":" + tag.Name
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

func (g *Gcr) regPublicImageNames() []string {

	if g.dockerHubToken == "" {
		g.hubToken()
	}

	if g.dockerHubImages == nil {
		g.dockerHubImages = make(map[string]bool)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf(HubRepos, g.DockerUser), nil)
	utils.CheckAndExit(err)
	req.Header.Set("Authorization", "JWT "+g.dockerHubToken)

	resp, err := g.httpClient.Do(req)
	utils.CheckAndExit(err)
	defer resp.Body.Close()

	var result struct {
		Results []struct {
			User string
			Name string
		}
	}
	b, err := ioutil.ReadAll(resp.Body)
	utils.CheckAndExit(err)
	jsoniter.Unmarshal(b, &result)

	var imageNames []string
	for _, repo := range result.Results {
		imageNames = append(imageNames, repo.Name)
	}
	return imageNames
}

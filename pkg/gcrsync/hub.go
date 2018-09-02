package gcrsync

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/json-iterator/go"
	"github.com/mritd/gcrsync/pkg/utils"
)

func (g *Gcr) hubToken() string {
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
	return token
}

func (g *Gcr) hubImageTags(repo string) []string {

	if g.dockerHubToken == "" {
		g.hubToken()
	}

	req, err := http.NewRequest("GET", fmt.Sprintf(HubTags, g.DockerUser, repo), nil)
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

	var tags []string
	for _, tag := range result.Results {
		tags = append(tags, tag.Name)
	}
	return tags
}

func (g *Gcr) hubImages() {

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
			Name string
		}
	}
	b, err := ioutil.ReadAll(resp.Body)
	utils.CheckAndExit(err)
	jsoniter.Unmarshal(b, &result)

	for _, repo := range result.Results {
		tags := g.hubImageTags(repo.Name)
		for _, tag := range tags {
			g.dockerHubImages[repo.Name+":"+tag] = true
		}
	}
}

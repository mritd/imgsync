package core

import "time"

const (
	DefaultLimit              = 20
	DefaultHTTPTimeOut        = 10 * time.Second
	DefaultSyncTimeout        = 1 * time.Hour
	DefaultGoRequestRetry     = 3
	DefaultGoRequestRetryTime = 5 * time.Second
	DefaultDockerRepo         = "docker.io"
	ManifestDir               = "manifests"
	ChangeLog                 = "CHANGELOG-%s.md"
	DockerHubImage            = "https://hub.docker.com/v2/repositories/%s/?page_size=100"
	DockerHubTags             = "https://hub.docker.com/v2/repositories/%s/%s/tags/?page_size=100"
)

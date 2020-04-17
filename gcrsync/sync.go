package gcrsync

const (
	ChangeLog       = "CHANGELOG-%s.md"
	GcrRegistryTpl  = "gcr.io/%s/%s"
	GcrImagesTpl    = "https://gcr.io/v2/%s/tags/list"
	GcrImageTagsTpl = "https://gcr.io/v2/%s/%s/tags/list"
	DockerHubImage  = "https://hub.docker.com/v2/repositories/%s/?page_size=100"
	DockerHubTags   = "https://hub.docker.com/v2/repositories/%s/%s/tags/?page_size=100"
)

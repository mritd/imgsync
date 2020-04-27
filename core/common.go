package core

import (
	"encoding/base64"
	"time"
)

const (
	DefaultLimit              = 20
	DefaultSyncTimeout        = 10 * time.Minute
	DefaultCtxTimeout         = 5 * time.Minute
	DefaultHTTPTimeout        = 30 * time.Second
	DefaultGoRequestRetry     = 3
	DefaultGoRequestRetryTime = 5 * time.Second
	DefaultDockerRepo         = "docker.io"
	DefaultK8sRepo            = "k8s.gcr.io"
	DefaultGcrRepo            = "gcr.io"
	DefaultGcrNamespace       = "google-containers"

	GcrKubeadmImagesTpl  = "https://k8s.gcr.io/v2/tags/list"
	GcrStandardImagesTpl = "https://gcr.io/v2/%s/tags/list"

	FlannelImageName = "quay.io/coreos/flannel"

	defaultSyncRetry     = 3
	defaultSyncRetryTime = 10 * time.Second

	// DockerHubTags  = "https://hub.docker.com/v2/repositories/%s/%s/tags/?page_size=100"
	// DockerHubImage = "https://hub.docker.com/v2/repositories/%s/?page_size=100"
	// ChangeLog = "CHANGELOG-%s.md"
	// GcrStandardManifestsTpl = "https://gcr.io/v2/%s/%s/manifests/%s"
	// GcrKubeadmManifestsTpl  = "https://k8s.gcr.io/v2/%s/manifests/%s"
	// GcrStandardImageTagsTpl = "https://gcr.io/v2/%s/%s/tags/list"
	// GcrKubeadmImageTagsTpl  = "https://k8s.gcr.io/v2/%s/tags/list"

	bannerBase64    = "ZSAgZWVlZWVlZSBlZWVlZSBlZWVlZSBlICAgIGUgZWVlZWUgZWVlZQo4ICA4ICA4ICA4IDggICA4IDggICAiIDggICAgOCA4ICAgOCA4ICA4CjhlIDhlIDggIDggOGUgICAgOGVlZWUgOGVlZWU4IDhlICA4IDhlCjg4IDg4IDggIDggODggIjggICAgODggICA4OCAgIDg4ICA4IDg4Cjg4IDg4IDggIDggODhlZTggOGVlODggICA4OCAgIDg4ICA4IDg4ZTgK"
	reportHeaderTpl = `%s
========================================
>> Sync Total: %d
>> Sync Failed: %d
>> Sync Success: %d
>> Manifests CacheHit: %d
`
	reportErrorTpl = `========================================
Sync failed images:
{{range .}}{{if not .Success}}{{. | print}}: {{.Err | println}}{{end}}{{end}}`
	reportSuccessTpl = `========================================
Sync success images:
{{range .}}{{if .Success}}{{. | print}}: {{if .CacheHit}}{{"hit cache" | println}}{{else}}{{"not hit cache" | println}}{{end}}{{end}}{{end}}`
)

var (
	ManifestDir = "manifests"
	Banner, _   = base64.StdEncoding.DecodeString(bannerBase64)
)

func retry(count int, interval time.Duration, f func() error) error {
	var err error
redo:
	count--
	if err = f(); err != nil {
		if count > 0 {
			if interval > 0 {
				<-time.After(interval)
			}
			goto redo
		}
	}
	return err
}

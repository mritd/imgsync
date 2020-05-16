package core

import (
	"fmt"
	"strings"
)

type Image struct {
	Repo string
	User string
	Name string
	Tag  string

	Success  bool
	CacheHit bool
	Err      error
}

func (img *Image) String() string {
	if img.User != "" {
		return fmt.Sprintf("%s/%s/%s:%s", img.Repo, img.User, img.Name, img.Tag)
	}
	return fmt.Sprintf("%s/%s:%s", img.Repo, img.Name, img.Tag)
}

func (img *Image) MergeName() string {
	if img.User != "" {
		return fmt.Sprintf("%s_%s_%s", strings.ReplaceAll(img.Repo, "/", "_"), strings.ReplaceAll(img.User, "/", "_"), strings.ReplaceAll(img.Name, "/", "_"))
	}
	return fmt.Sprintf("%s_%s", strings.ReplaceAll(img.Repo, "/", "_"), strings.ReplaceAll(img.Name, "/", "_"))
}

type Images []*Image

func (imgs Images) Len() int           { return len(imgs) }
func (imgs Images) Less(i, j int) bool { return imgs[i].String() < imgs[j].String() }
func (imgs Images) Swap(i, j int)      { imgs[i], imgs[j] = imgs[j], imgs[i] }

type GcrResp struct {
	Child    []string               `json:"child"`
	Manifest map[string]GcrManifest `json:"manifest"`
	Name     string                 `json:"name"`
	Tags     []string               `json:"tags"`
}

type GcrManifest struct {
	ImageSizeBytes string   `json:"imageSizeBytes"`
	LayerID        string   `json:"layerId"`
	MediaType      string   `json:"mediaType"`
	Tag            []string `json:"tag"`
	TimeCreatedMS  string   `json:"timeCreatedMs"`
	TimeUploadedMS string   `json:"timeUploadedMs"`
}

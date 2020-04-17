package imgsync

import "fmt"

type Image struct {
	Repo string
	User string
	Name string
	Tag  string
}

func (img Image) String() string {
	if img.User != "" {
		return fmt.Sprintf("%s/%s/%s:%s", img.Repo, img.User, img.Name, img.Tag)
	} else {
		return fmt.Sprintf("%s/%s:%s", img.Repo, img.Name, img.Tag)
	}
}

func (img Image) MergeName() string {
	if img.User != "" {
		return fmt.Sprintf("%s_%s_%s", img.Repo, img.User, img.Name)
	} else {
		return fmt.Sprintf("%s_%s", img.Repo, img.Name)
	}
}

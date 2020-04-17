package imgsync

import "fmt"

type Image struct {
	Repo string
	User string
	Name string
	Tag  string
}

func (img Image) String() string {
	return fmt.Sprintf("%s/%s/%s:%s", img.Repo, img.User, img.Name, img.Tag)
}

func (img Image) MergeName() string {
	return fmt.Sprintf("%s_%s_%s", img.Repo, img.User, img.Name)
}

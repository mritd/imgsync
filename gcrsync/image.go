package gcrsync

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

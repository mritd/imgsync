package core

import (
	"fmt"
)

type Synchronizer interface {
	Init() error
	Images() Images
	Manifests() []string
	Sync()
}

func New(name string) (Synchronizer, error) {
	switch name {
	case "gcr":
	case "flannel":
	case "kubernetes":

	}

	return nil, fmt.Errorf("failed to create synchronizer %s: unknown synchronizer", name)
}

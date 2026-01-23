package component

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrAlreadyLatest = errors.New("already with latest build")

	RepostoryDir = fmt.Sprintf("%s/.dingo/components", func() string {
		homeDir, _ := os.UserHomeDir()
		return homeDir
	}())
)

const (
	DINGO_CLIENT     = "dingo-client"
	DINGO_DACHE      = "dingo-cache"
	DINGO_MDS        = "dingo-mds"
	DINGO_MDS_CLIENT = "dingo-mds-client"
	INSTALLED_FILE   = "installed_components.json"
	MIRROR           = "https://www.dingodb.com/dingofs"
	LASTEST_VERSION  = "latest"
)

type Component struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
	IsInstalled bool   `json:"installed"`
	IsActive    bool   `json:"active"`
	Release     string `json:"release"`
	Path        string `json:"path"`
	URL         string `json:"url"`
}

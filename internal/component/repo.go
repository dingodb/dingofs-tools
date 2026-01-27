package component

import "fmt"

type BinaryRepoData struct {
	Binary      string                  `json:"binary"`
	GeneratedAt string                  `json:"generated_at"`
	Branches    map[string]BinaryDetail `json:"branches"`
	Commits     map[string]BinaryDetail `json:"commits"`
	Tags        map[string]BinaryDetail `json:"tags"`
}

type BinaryDetail struct {
	Path      string `json:"path"`
	BuildTime string `json:"build_time"`
	Size      string `json:"size"`
	Commit    string `json:"commit,omitempty"`
}

func (b *BinaryRepoData) GetBranches() map[string]BinaryDetail {
	return b.Branches
}
func (b *BinaryRepoData) GetTags() map[string]BinaryDetail {
	return b.Tags
}

func (b *BinaryRepoData) GetCommits() map[string]BinaryDetail {
	return b.Commits
}

func (b *BinaryRepoData) GetLatest() (string, *BinaryDetail, bool) {
	latest := "v0.0.0"
	for version := range b.Tags {
		if version > latest {
			latest = version
		}
	}

	tag, ok := b.Tags[latest]
	if ok {
		return latest, &tag, true
	}

	return "", nil, false
}

func (b *BinaryRepoData) GetMain() (*BinaryDetail, bool) {
	if branch, exists := b.Branches[MAIN_VERSION]; exists {
		return &branch, true
	}

	return nil, false
}

func (b *BinaryRepoData) FindVersion(tag string) (*BinaryDetail, bool) {
	if tag, exists := b.Tags[tag]; exists {
		return &tag, true
	}

	return nil, false
}

func (b *BinaryRepoData) GetName() string {
	return b.Binary
}

func NewBinaryRepoData(name string) (*BinaryRepoData, error) {
	requestURL := URLJoin(MIRROR, fmt.Sprintf("%s.version", name))
	metadata, err := ParseFromURL(requestURL)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

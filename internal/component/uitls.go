package component

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

// input string maybe:
// dingo-mds:v1.0.0
// dingo-client
func ParseComponentVersion(input string) (string, string) {
	if strings.Contains(input, ":") {
		parts := strings.SplitN(input, ":", 2)
		return parts[0], parts[1]
	}

	return input, ""
}

func URLJoin(base string, paths ...string) string {
	u, err := url.Parse(base)
	if err != nil {
		panic(fmt.Sprintf("invalid base URL: %v", err))
	}

	for _, p := range paths {
		if p != "" {
			u.Path = path.Join(u.Path, p)
		}
	}

	return u.String()
}

func ParseBinaryRepoData(jsonData []byte) (*BinaryRepoData, error) {
	var metadata BinaryRepoData

	if err := json.Unmarshal(jsonData, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &metadata, nil
}

func ParseFromFile(filename string) (*BinaryRepoData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseBinaryRepoData(data)
}

func ParseFromURL(url string) (*BinaryRepoData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return ParseBinaryRepoData(data)
}

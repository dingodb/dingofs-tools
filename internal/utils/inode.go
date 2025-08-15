package cobrautil

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

func RemoveDuplicates(strs []string) []string {
	seen := make(map[string]bool)
	for _, str := range strs {
		seen[str] = true
	}

	result := make([]string, 0, len(seen))
	for str := range seen {
		result = append(result, str)
	}
	return result
}

// get mountPoint inode
func GetFileInode(path string) (uint64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if sst, ok := fi.Sys().(*syscall.Stat_t); ok {
		return sst.Ino, nil
	}
	return 0, nil
}

func GetInodesAsString(listFilePath string) (string, error) {
	content, err := os.ReadFile(listFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file list: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	var inodeStrings []string

	for _, line := range lines {
		filePath := strings.TrimSpace(line)
		if filePath == "" {
			continue
		}

		inodeId, err2 := GetFileInode(filePath)
		if err2 != nil {
			return "", fmt.Errorf("failed to get inode for %s: not a syscall.Stat_t", filePath)
		}
		if inodeId == 0 {
			continue
		}
		inodeStrings = append(inodeStrings, fmt.Sprintf("%d", inodeId))
	}

	inodeStrings = RemoveDuplicates(inodeStrings)
	
	return strings.Join(inodeStrings, ","), nil
}

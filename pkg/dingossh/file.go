/*
 * 	Copyright (c) 2025 dingodb.com Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package dingossh

import (
	"errors"
	"fmt"
	"os"

	"github.com/dingodb/dingofs-tools/internal/logger"
)

const (
	TEMP_DIR = "/tmp"
)

var (
	ERR_UNREACHED = errors.New("remote unreached")
)

type FileManager struct {
	sshClient *SSHClient
}

func NewFileManager(sshClient *SSHClient) *FileManager {
	return &FileManager{sshClient: sshClient}
}

func (f *FileManager) Upload(localPath, remotePath string) error {
	if f.sshClient == nil {
		return ERR_UNREACHED
	}

	err := f.sshClient.Client().Upload(localPath, remotePath)
	logger.GetLogger().Infof("UploadFile {remoteAddress: %s, localPath: %s, remotePath: %s}, error : {%v}",
		remoteAddr(f.sshClient), localPath, remotePath, err)
	return err
}

func (f *FileManager) Download(remotePath, localPath string) error {
	if f.sshClient == nil {
		return ERR_UNREACHED
	}

	err := f.sshClient.Client().Download(remotePath, localPath)
	logger.GetLogger().Infof("DownloadFile {remoteAddress: %s, remotePath: %s, localPath: %s} failed, error test -: {%v}",
		remoteAddr(f.sshClient), remotePath, localPath, err)
	return err
}

func (f *FileManager) Install(content, destPath string) error {
	file, err := os.CreateTemp(TEMP_DIR, "dingo.*.install")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	n, err := file.WriteString(content)
	if err != nil {
		return err
	} else if n != len(content) {
		return fmt.Errorf("written: expect %d bytes, actually %d bytes", len(content), n)
	}

	return os.Rename(file.Name(), destPath)
}

func (f *FileManager) InstallTmpFile(content string) (string, error) {
	file, err := os.CreateTemp(TEMP_DIR, "dingo.*.install")
	if err != nil {
		return "", err
	}

	n, err := file.WriteString(content)
	if err != nil {
		return "", err
	} else if n != len(content) {
		return "", fmt.Errorf("written: expect %d bytes, actually %d bytes", len(content), n)
	}

	return file.Name(), nil
}

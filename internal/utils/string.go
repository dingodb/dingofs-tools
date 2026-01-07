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

package utils

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

const (
	PATH_REGEX    = `^(/[^/ ]*)+/?$`
	FS_NAME_REGEX = "^([a-z0-9]+\\-?)+$"
	K_STRING_TRUE = "true"

	ROOT_PATH       = "/"
	RECYCLEBIN_PATH = "/RecycleBin"
)

func IsValidFsname(fsName string) bool {
	matched, err := regexp.MatchString(FS_NAME_REGEX, fsName)
	if err != nil || !matched {
		return false
	}
	return true
}

// rm whitespace
func RmWitespaceStr(str string) string {
	if str == "" {
		return ""
	}

	reg := regexp.MustCompile(`\s+`)
	return reg.ReplaceAllString(str, "")
}

func prompt(prompt string) string {
	if prompt != "" {
		prompt += " "
	}
	fmt.Print(color.YellowString("WARNING:"), prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(input, "\n")
}

func AskConfirmation(promptStr string, confirm string) bool {
	promptStr = promptStr + fmt.Sprintf("\nplease input [%s] to confirm:", confirm)
	ans := prompt(promptStr)
	switch strings.TrimSpace(ans) {
	case confirm:
		return true
	default:
		return false
	}
}

func IsValidPath(path string) bool {
	match, _ := regexp.MatchString(PATH_REGEX, path)
	return match
}

func GetString2Signature(date uint64, owner string) string {
	return fmt.Sprintf("%d:%s", date, owner)
}

func CalcString2Signature(in string, secretKet string) string {
	h := hmac.New(sha256.New, []byte(secretKet))
	h.Write([]byte(in))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func IsDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func IsAlpha(r rune) bool {
	return ('a' <= r && r <= 'z') || IsUpper(r)
}

func IsUpper(r rune) bool {
	return 'A' <= r && r <= 'Z'
}

func ToUnderscoredName(src string) string {
	var ret string
	for i, c := range src {
		if IsAlpha(c) {
			if c < 'a' { // upper cases
				if i != 0 && !IsUpper(rune(src[i-1])) && ret[len(ret)-1] != '-' {
					ret += "_"
				}
				ret += string(c - 'A' + 'a')
			} else {
				ret += string(c)
			}
		} else if IsDigit(c) {
			ret += string(c)
		} else if len(ret) == 0 || ret[len(ret)-1] != '_' {
			ret += "_"
		}
	}
	return ret
}

func StringList2Uint64List(strList []string) ([]uint64, error) {
	retList := make([]uint64, 0)
	for _, str := range strList {
		v, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return nil, err
		}
		retList = append(retList, v)
	}
	return retList, nil
}

func StringList2Uint32List(strList []string) ([]uint32, error) {
	retList := make([]uint32, 0)
	for _, str := range strList {
		v, err := strconv.ParseUint(str, 10, 32)
		if err != nil {
			return nil, err
		}
		retList = append(retList, uint32(v))
	}
	return retList, nil
}

func RemoveHTTPPrefix(endpoint string) string {
	re := regexp.MustCompile(`^(?i)https?://`)
	return re.ReplaceAllString(endpoint, "")
}

func IsHTTPS(endpoint string) bool {
	matched, _ := regexp.MatchString(`^(?i)https://`, endpoint)
	return matched
}

func IsSSL(host string, timeout time.Duration) bool {
	conf := &tls.Config{
		InsecureSkipVerify: true, // ignore certificate check（only check protocol）
	}
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: timeout},
		"tcp",
		host,
		conf,
	)
	if err != nil {
		return false
	}
	defer conn.Close()

	// check TLS Handshake status
	if err := conn.Handshake(); err != nil {
		return false // Handshake failed
	}

	return true // is ssl
}

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

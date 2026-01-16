// Copyright (c) 2025 dingodb.com, Inc. All Rights Reserved
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package output

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/dingodb/dingocli/internal/common"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func init() {
	log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)
	log.SetOutput(io.Discard)
}

func SetShow(show bool) {
	if show {
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(io.Discard)
	}
}

func MarshalProtoJson(message proto.Message) (interface{}, error) {
	m := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		EmitUnpopulated: true,
	}
	jsonByte, err := m.Marshal(message)
	if err != nil {
		return nil, err
	}
	var ret interface{}
	err = json.Unmarshal(jsonByte, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func ProtoMessageToJson(message proto.Message) (string, error) {
	m := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		EmitUnpopulated: true,
	}
	value, err := m.Marshal(message)
	return string(value), err
}

func ShowRpcData(request proto.Message, response proto.Message, isShow bool) {
	if isShow {
		log.SetOutput(os.Stdout)
		data, _ := ProtoMessageToJson(request)
		log.Printf("rpc request info: %s\n", data)
		data, _ = ProtoMessageToJson(response)
		log.Printf("rpc response info: %s\n", data)
	}
}

func OutputJson(result *common.OutputResult) error {
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}

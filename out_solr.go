// Copyright 2019 Oliver Szabo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"C"
	"fmt"
	"github.com/fluent/fluent-bit-go/output"
	"github.com/oleewere/go-solr-client/solr"
	"github.com/ugorji/go/codec"
	"unsafe"
)

var solrClient *solr.SolrClient

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "solr", "Solr Output")
}

//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {
	// Example to retrieve an optional configuration parameter
	//param := output.FLBPluginConfigKey(ctx, "param")
	//fmt.Printf("[flb-go] plugin parameter = '%s'\n", param)
	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	var bytesData []byte
	var h codec.MsgpackHandle
	var message interface{}

	bytesData = C.GoBytes(data, length)
	dec := codec.NewDecoderBytes(bytesData, &h)
	err := dec.Decode(&message)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Processing Data")
	fmt.Println(bytesData)
	fmt.Println(message)

	return 0
}

//export FLBPluginExit
func FLBPluginExit() int {
	return 0
}

func main() {
}

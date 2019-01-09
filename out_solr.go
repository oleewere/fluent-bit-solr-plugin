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
	"unsafe"
	"fmt"
	"github.com/fluent/fluent-bit-go/output"
)

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "gsolrout", "Solr output GO!")
}

//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {
	// Example to retrieve an optional configuration parameter
	param := output.FLBPluginConfigKey(ctx, "param")
	fmt.Printf("[flb-go] plugin parameter = '%s'\n", param)
	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
  fmt.Println("Do something")
}

func FLBPluginExit() int {
	return 0
}

func main() {
}

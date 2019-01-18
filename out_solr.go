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
	"encoding/binary"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/fluent/fluent-bit-go/output"
	"github.com/oleewere/go-buffered-processor/processor"
	"github.com/oleewere/go-solr-client/solr"
	"github.com/satori/go.uuid"
	"github.com/ugorji/go/codec"
	"log"
	"reflect"
	"sync"
	"time"
	"unsafe"
)

var solrClient *solr.SolrClient
var solrConfig solr.SolrConfig
var batchContext *processor.BatchContext
var proc SolrDataProcessor

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "solr", "Solr Output")
}

//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {
	collection := output.FLBPluginConfigKey(ctx, "Collection")
	url := output.FLBPluginConfigKey(ctx, "Url")
	urlContext := output.FLBPluginConfigKey(ctx, "Context")
	solrConfig = solr.SolrConfig{}
	if len(urlContext) > 0 {
		solrConfig.SolrUrlContext = urlContext
	} else {
		solrConfig.SolrUrlContext = "/solr"
	}
	solrConfig.Collection = collection
	solrConfig.Url = url
	log.Printf("Solr output URL: %s, Context: %s, Collection: %s", solrConfig.Url, solrConfig.SolrUrlContext, solrConfig.Collection)
	solrClient, _ = solr.NewSolrClient(&solrConfig)
	batchContext = processor.CreateDefaultBatchContext()
	batchContext.MaxBufferSize = 1
	batchContext.MaxRetries = 20
	batchContext.RetryTimeInterval = 10

	proc = SolrDataProcessor{SolrClient: solrClient, Mutex: &sync.Mutex{}}
	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	dec := NewDecoder(data, length)
	_, timestamp, record := GetRecord(dec)
	spew.Dump(timestamp)
	spew.Dump(record)

	log.Println(record)

	processor.ProcessData(record, batchContext, proc)

	return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	return 0
}

type FLBDecoder struct {
	handle *codec.MsgpackHandle
	mpdec  *codec.Decoder
}

type FLBTime struct {
	time.Time
}

func (f FLBTime) WriteExt(interface{}) []byte {
	panic("unsupported")
}

func (f FLBTime) ReadExt(i interface{}, b []byte) {
	out := i.(*FLBTime)
	sec := binary.BigEndian.Uint32(b)
	usec := binary.BigEndian.Uint32(b[4:])
	out.Time = time.Unix(int64(sec), int64(usec))
}

func (f FLBTime) ConvertExt(v interface{}) interface{} {
	return nil
}

func (f FLBTime) UpdateExt(dest interface{}, v interface{}) {
	panic("unsupported")
}

func NewDecoder(data unsafe.Pointer, length C.int) *FLBDecoder {
	var b []byte

	dec := new(FLBDecoder)
	dec.handle = new(codec.MsgpackHandle)
	dec.handle.SetExt(reflect.TypeOf(FLBTime{}), 0, &FLBTime{})

	b = C.GoBytes(data, length)
	dec.mpdec = codec.NewDecoderBytes(b, dec.handle)

	return dec
}

func GetRecord(dec *FLBDecoder) (ret int, ts interface{}, rec map[string]string) {
	var check error
	var m interface{}

	check = dec.mpdec.Decode(&m)
	if check != nil {
		return -1, 0, nil
	}

	slice := reflect.ValueOf(m)
	t := slice.Index(0).Interface()
	data := slice.Index(1)

	mapInterfaceData := data.Interface().(map[interface{}]interface{})

	mapData := make(map[string]string)

	for kData, vData := range mapInterfaceData {
		mapData[kData.(string)] = string(vData.([]uint8))
	}

	mapData["id"] = uuid.NewV4().String()

	return 0, t, mapData
}

// SolrDataProcessor type for processing Solr data
type SolrDataProcessor struct {
	Mutex      *sync.Mutex
	SolrClient *solr.SolrClient
}

// Process send gathered data to Solr
func (p SolrDataProcessor) Process(batchContext *processor.BatchContext) error {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	log.Println(batchContext.BufferData)
	_, _, err := p.SolrClient.Update(batchContext.BufferData, nil, true)
	return err
}

// HandleError handle errors during time based buffer processing (it is not used by this generator)
func (p SolrDataProcessor) HandleError(batchContext *processor.BatchContext, err error) {
	fmt.Println(err)
}

func main() {
}

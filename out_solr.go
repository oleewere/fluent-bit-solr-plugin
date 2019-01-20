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
	"github.com/fluent/fluent-bit-go/output"
	"github.com/oleewere/go-buffered-processor/processor"
	"github.com/oleewere/go-solr-client/solr"
	"github.com/satori/go.uuid"
	"github.com/ugorji/go/codec"
	"io"
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
var useEpoch string
var useBufferedProcessor string
var timeSolrField string

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "solr", "Solr Output")
}

//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {
	collection := output.FLBPluginConfigKey(ctx, "Collection")
	url := output.FLBPluginConfigKey(ctx, "Url")
	urlContext := output.FLBPluginConfigKey(ctx, "Context")
	useEpoch = output.FLBPluginConfigKey(ctx, "Epoch")
	useBufferedProcessor = output.FLBPluginConfigKey(ctx, "BufferedProcessor")
	timeSolrField = output.FLBPluginConfigKey(ctx, "TimeSolrField")
	solrConfig = solr.SolrConfig{}
	if len(urlContext) > 0 {
		solrConfig.SolrUrlContext = urlContext
	} else {
		solrConfig.SolrUrlContext = "/solr"
	}
	if len(timeSolrField) == 0 {
		timeSolrField = "logtime"
	}
	solrConfig.Collection = collection
	solrConfig.Url = url
	log.Printf("Solr output URL: %s, Context: %s, Collection: %s", solrConfig.Url, solrConfig.SolrUrlContext, solrConfig.Collection)
	solrClient, _ = solr.NewSolrClient(&solrConfig)
	batchContext = processor.CreateDefaultBatchContext()
	batchContext.MaxBufferSize = 1
	batchContext.MaxRetries = 20
	batchContext.RetryTimeInterval = 10

	// time based processing
	batchContext.ProcessTimeInterval = 60 * time.Second
	batchContext.TimeBasedProcessing = true
	if useBufferedProcessor == "true" {
		proc = SolrDataProcessor{SolrClient: solrClient, Mutex: &sync.Mutex{}}
		go processor.StartTimeBasedProcessing(batchContext, proc, 60*time.Second)
	}
	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	dec := NewDecoder(data, length)
	records := make([]map[string]string, 0)

	for {
		var m interface{}

		err := dec.mpdec.Decode(&m)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Print("Failed to decode msgpack data: %v\n", err)
			return output.FLB_ERROR
		}

		_, timestamp, record := GetRecord(m)

		if useEpoch == "true" {
			dateTimestamp := timestamp.(FLBTime).Time.Unix()
			record[timeSolrField] = fmt.Sprint(dateTimestamp)
		} else {
			record[timeSolrField] = timestamp.(FLBTime).Time.Format("2006-01-02T15:04:05.000")
		}
		if useBufferedProcessor == "true" {
			processor.ProcessData(record, batchContext, proc)
		} else {
			records = append(records, record)
		}
	}
	if useBufferedProcessor != "true" {
		solrClient.Update(records, nil, true)
	}

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

func GetRecord(m interface{}) (ret int, ts interface{}, rec map[string]string) {
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
	_, _, err := p.SolrClient.Update(batchContext.BufferData, nil, true)
	return err
}

// HandleError handle errors during time based buffer processing (it is not used by this generator)
func (p SolrDataProcessor) HandleError(batchContext *processor.BatchContext, err error) {
	log.Println(err)
}

func main() {
}

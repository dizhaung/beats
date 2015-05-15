package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/elastic/libbeat/common"
)

type BulkMsg struct {
	Ts    time.Time
	Event common.MapStr
}

func (es *Elasticsearch) Bulk(index string, doc_type string,
	params map[string]string, body chan interface{}) (*QueryResult, error) {

	path, err := MakePath(index, doc_type, "_bulk")
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for obj := range body {
		enc.Encode(obj)
	}

	url := es.Url + path
	if len(params) > 0 {
		url = url + "?" + UrlEncode(params)
	}

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, err
	}

	resp, err := es.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	obj, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result QueryResult
	err = json.Unmarshal(obj, &result)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		return &result, fmt.Errorf("ES returned an error: %s", resp.Status)
	}
	return &result, err
}

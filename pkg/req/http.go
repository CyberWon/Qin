package req

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type HTTP struct {
	URL string
}

type HTTPResult map[string]interface{}

func (h *HTTP) JSONToMap(body []byte) HTTPResult {

	var httpResult = make(HTTPResult)

	if err := json.Unmarshal(body, &httpResult); err != nil {
		log.Panic(err)
	}
	return httpResult
}

func (h *HTTP) Req(method string, data []byte) (HTTPResult, error) {
	reader := bytes.NewReader(data)
	// 忽略 ssl证书问题
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}

	// 构建request对象，并返回
	if request, err := http.NewRequest(method, h.URL, reader); err != nil {
		return nil, err
	} else {
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("accept", "application/json")
		if res, err := client.Do(request); err != nil {
			return nil, err
		} else {
			if body, err := ioutil.ReadAll(res.Body); err != nil {
				return nil, err
			} else {
				return h.JSONToMap(body), nil
			}
		}
	}
}

func (h *HTTP) Post(url string, data []byte) HTTPResult {
	h.URL = url
	httpResult, err := h.Req("POST", data)
	if err != nil {
		log.Panic(err)
	}
	return httpResult
}

package utils

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/krabt/pmsg/utils/httpclient"
)

func PostJSON(url string, reqBody, respBody any) (http.Header, error) {
	body, err := JsonEncode(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := httpclient.Post(url, httpclient.HdrValApplicationJson, body)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// "application/json"                 响应内容类型
	// "application/json; charset=utf-8"  错误响应内容类型
	contentType := resp.Header.Get("content-type")
	if !strings.EqualFold(contentType, httpclient.HdrValApplicationJson) && !strings.EqualFold(contentType, httpclient.HdrValApplicationJsonCharset) {
		return nil, fmt.Errorf("http.response.header.content-type != %s", httpclient.HdrValApplicationJson)
	}
	if err := JsonDecode(resp.Body, respBody); err != nil {
		return nil, fmt.Errorf("http.response.body json decode failed, %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return resp.Header, fmt.Errorf("invalid http.response.status: %s", resp.Status)
	}
	return resp.Header, nil
}

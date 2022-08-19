/*
 * Copyright (c) 2022, MegaEase
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/eval"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/stretchr/testify/assert"
)

func createHTTP() *HTTP {
	return &HTTP{
		DefaultProbe:    base.DefaultProbe{ProbeName: "dummy http"},
		URL:             "http://example.com",
		ContentEncoding: "text/json",
		Headers:         map[string]string{"header1": "value1", "header2": "value2"},
		Body:            "{ \"key1\": \"value1\", \"key2\": \"value2\" }",
		TextChecker: probe.TextChecker{
			Contain:    "good",
			NotContain: "bad",
		},
		Evaluator: eval.Evaluator{
			Variables: []eval.Variable{
				{
					Name:  "name",
					Type:  eval.String,
					Query: "//name",
					Value: nil,
				},
			},
			DocType:    eval.JSON,
			Expression: "name == 'EaseProbe'",
		},
		User: "user",
		Pass: "pass",
		TLS: global.TLS{
			CA:   "ca.crt",
			Cert: "cert.crt",
			Key:  "key.key",
		},
	}
}
func TestHTTPConfig(t *testing.T) {
	h := createHTTP()
	err := h.Config(global.ProbeSettings{})
	assert.Error(t, err)

	//TLS config success
	var gtls *global.TLS
	monkey.PatchInstanceMethod(reflect.TypeOf(gtls), "Config", func(_ *global.TLS) (*tls.Config, error) {
		return &tls.Config{}, nil
	})

	h.URL = "@$186example.com"
	err = h.Config(global.ProbeSettings{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid URI")
	h.URL = "https://example.com"

	h.SuccessCode = append(h.SuccessCode, []int{900, 999, 1000})
	err = h.Config(global.ProbeSettings{})
	assert.NoError(t, err)
	assert.Equal(t, [][]int{{0, 499}}, h.SuccessCode)

	err = h.Config(global.ProbeSettings{})
	assert.NoError(t, err)
	assert.NotNil(t, h.TLS)
	assert.Equal(t, "GET", h.Method)

	h.Proxy = "http://uername:password@proxy.example.com:8080"
	err = h.Config(global.ProbeSettings{})
	assert.NoError(t, err)

	var e *eval.Evaluator
	monkey.PatchInstanceMethod(reflect.TypeOf(e), "Config", func(_ *eval.Evaluator) error {
		return fmt.Errorf("Eval Config Error")
	})
	err = h.Config(global.ProbeSettings{})
	assert.Error(t, err)

	h.Proxy = "\nexample.com"
	err = h.Config(global.ProbeSettings{})
	assert.Error(t, err)

	monkey.UnpatchAll()
}

func TestTextCheckerConfig(t *testing.T) {
	h := createHTTP()
	h.TextChecker = probe.TextChecker{
		Contain:    "",
		NotContain: "",
		RegExp:     true,
	}
	h.TLS = global.TLS{}

	err := h.Config(global.ProbeSettings{})
	assert.NoError(t, err)

	h.Contain = `[a-zA-z]\d+`
	err = h.Config(global.ProbeSettings{})
	assert.NoError(t, err)
	assert.Equal(t, `[a-zA-z]\d+`, h.TextChecker.Contain)

	h.NotContain = `(?=.*word1)(?=.*word2)`
	err = h.Config(global.ProbeSettings{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or unsupported Perl syntax")
}

func TestHTTPDoProbe(t *testing.T) {
	// clear request
	h := createHTTP()
	var gtls *global.TLS
	monkey.PatchInstanceMethod(reflect.TypeOf(gtls), "Config", func(_ *global.TLS) (*tls.Config, error) {
		return &tls.Config{}, nil
	})
	err := h.Config(global.ProbeSettings{})
	assert.NoError(t, err)

	var client *http.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(ioutil.NopCloser(nil)),
		}, nil
	})
	monkey.Patch(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
		return []byte(`{ "name": "EaseProbe", "status": "good"}`), nil
	})

	s, m := h.DoProbe()
	assert.True(t, s)
	assert.Contains(t, m, "200")

	// evaluate error
	h.Evaluator.Expression = "name == 'N/A'"
	s, m = h.DoProbe()
	assert.False(t, s)
	assert.Contains(t, m, "Expression is evaluated to false")

	h.Evaluator.Variables[0].Query = "///name"
	s, m = h.DoProbe()
	assert.False(t, s)
	assert.Contains(t, m, "Evaluation Error")

	// response does not contain good string
	monkey.Patch(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
		return []byte("bad"), nil
	})
	s, m = h.DoProbe()
	assert.False(t, s)
	assert.Contains(t, m, "good")

	// response does contain the bad string
	h.Contain = ""
	s, m = h.DoProbe()
	assert.False(t, s)
	assert.Contains(t, m, "bad")

	// response is 503
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 503,
			Body:       io.NopCloser(ioutil.NopCloser(nil)),
		}, nil
	})
	s, m = h.DoProbe()
	assert.False(t, s)
	assert.Contains(t, m, "503")

	// io read failure
	monkey.Patch(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
		return nil, fmt.Errorf("read error")
	})
	s, m = h.DoProbe()
	assert.False(t, s)
	assert.Contains(t, m, "read error")

	// request failure
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("request error")
	})
	s, m = h.DoProbe()
	assert.False(t, s)
	assert.Contains(t, m, "request error")

	// http new request failure
	monkey.Patch(http.NewRequest, func(method, url string, body io.Reader) (*http.Request, error) {
		return nil, fmt.Errorf("new request error")
	})
	s, m = h.DoProbe()
	assert.False(t, s)
	assert.Contains(t, m, "new request error")

	monkey.UnpatchAll()

	var d *net.Dialer
	monkey.PatchInstanceMethod(reflect.TypeOf(d), "DialContext", func(*net.Dialer, context.Context, string, string) (net.Conn, error) {
		return &net.TCPConn{}, nil
	})

	h.TLS = global.TLS{
		CA:   "",
		Cert: "",
		Key:  "",
	}
	h.URL = "http://127.0.0.1:1234"
	err = h.Config(global.ProbeSettings{})
	assert.NoError(t, err)

	s, m = h.DoProbe()
	assert.False(t, s)
	assert.NotContains(t, m, "200")

	monkey.UnpatchAll()
}

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

package dingtalk

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/report"
	"github.com/stretchr/testify/assert"
)

func assertError(t *testing.T, err error, msg string, contain bool) {
	assert.Error(t, err)
	if contain {
		assert.Contains(t, err.Error(), msg)
	} else {
		assert.Equal(t, msg, err.Error())
	}
}

func TestDingTalk(t *testing.T) {
	conf := &NotifyConfig{
		SignSecret: "secret",
	}
	err := conf.Config(global.NotifySettings{})
	assert.NoError(t, err)
	assert.Equal(t, report.Markdown, conf.NotifyFormat)
	assert.Equal(t, "dingtalk", conf.Kind())

	var client *http.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(strings.NewReader(`{"errmsg": "ok", "errcode": 0}`))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})
	err = conf.SendDingtalkNotification("title", "message")
	assert.NoError(t, err)

	// bad response
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(strings.NewReader(`{"errmsg": "error", "errcode": 1}`))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})
	err = conf.SendDingtalkNotification("title", "message")
	assertError(t, err, "Error response from Dingtalk [200]", true)

	// bad json
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(strings.NewReader(`{"errmsg": "error", "errcode = 1}`))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})
	err = conf.SendDingtalkNotification("title", "message")
	assertError(t, err, "Error response from Dingtalk [200]", true)

	// bad ioutil.ReadAll
	monkey.Patch(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
		return nil, errors.New("read error")
	})
	err = conf.SendDingtalkNotification("title", "message")
	assertError(t, err, "read error", false)

	// bad http do
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		return nil, errors.New("http do error")
	})
	err = conf.SendDingtalkNotification("title", "message")
	assertError(t, err, "http do error", false)

	// bad http.NewRequest
	monkey.Patch(http.NewRequest, func(method, url string, body io.Reader) (*http.Request, error) {
		return nil, errors.New("http.NewRequest error")
	})
	err = conf.SendDingtalkNotification("title", "message")
	assertError(t, err, "http.NewRequest error", false)

	monkey.UnpatchAll()

}

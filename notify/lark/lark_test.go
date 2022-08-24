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

package lark

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

func assertError(t *testing.T, err error, msg string) {
	assert.Error(t, err)
	assert.Equal(t, msg, err.Error())
}
func TestLark(t *testing.T) {
	conf := &NotifyConfig{}
	err := conf.Config(global.NotifySettings{})
	assert.NoError(t, err)
	assert.Equal(t, report.Lark, conf.NotifyFormat)
	assert.Equal(t, "lark", conf.Kind())

	var client *http.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(strings.NewReader(`{"StatusCode": 0}`))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})
	err = conf.SendLark("title", "message")
	assert.NoError(t, err)

	// bad response
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(strings.NewReader(`{"StatusCode": "100"}`))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})
	err = conf.SendLark("title", "message")
	assertError(t, err, "Error response from Lark - code [0] - msg []")

	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(strings.NewReader(`{"StatusCode": "100", "code": 10, "msg": "lark error"}`))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})
	err = conf.SendLark("title", "message")
	assertError(t, err, "Error response from Lark - code [10] - msg [lark error]")

	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(strings.NewReader(`bad : json format`))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})
	err = conf.SendLark("title", "message")
	assertError(t, err, "Error response from Lark [200] - [bad : json format]")

	monkey.Patch(ioutil.ReadAll, func(_ io.Reader) ([]byte, error) {
		return nil, errors.New("read error")
	})
	err = conf.SendLark("title", "message")
	assertError(t, err, "read error")

	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		return nil, errors.New("http error")
	})
	err = conf.SendLark("title", "message")
	assertError(t, err, "http error")

	monkey.Patch(http.NewRequest, func(_ string, _ string, _ io.Reader) (*http.Request, error) {
		return nil, errors.New("new request error")
	})
	err = conf.SendLark("title", "message")
	assertError(t, err, "new request error")

	monkey.UnpatchAll()

}

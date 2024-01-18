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

package nexmo

import (
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"

	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/notify/sms/conf"
	"github.com/stretchr/testify/assert"
)

func assertError(t *testing.T, err error, msg string) {
	t.Helper()
	assert.Error(t, err)
	assert.Equal(t, msg, err.Error())
}

func testNotify(t *testing.T, provider conf.Provider) {
	var client *http.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := io.NopCloser(strings.NewReader(`ok`))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})

	err := provider.Notify("title", "text")
	assert.NoError(t, err)

	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := io.NopCloser(strings.NewReader(`Internal Server Error`))
		return &http.Response{
			StatusCode: 500,
			Body:       r,
		}, nil
	})
	err = provider.Notify("title", "text")
	assertError(t, err, "Error response from SMS with request body <From=&To=&api_key=&api_secret=&text=text> [500] - [Internal Server Error]")

	monkey.Patch(io.ReadAll, func(_ io.Reader) ([]byte, error) {
		return nil, errors.New("read error")
	})
	err = provider.Notify("title", "text")
	assertError(t, err, "read error")

	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		return nil, errors.New("http do error")
	})
	err = provider.Notify("title", "text")
	assertError(t, err, "http do error")

	monkey.Patch(http.NewRequest, func(_ string, _ string, _ io.Reader) (*http.Request, error) {
		return nil, errors.New("http new request error")
	})
	err = provider.Notify("title", "text")
	assertError(t, err, "http new request error")

	monkey.UnpatchAll()
}

func TestTwilio(t *testing.T) {
	opt := conf.Options{}
	p := New(opt)
	testNotify(t, p)
}

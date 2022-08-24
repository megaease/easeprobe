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

package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func newDummyResult(name string) probe.Result {
	var total int64 = 100
	up := rand.Int63n(total)
	down := rand.Int63n(total - up)

	r := probe.Result{
		Name:             name,
		Endpoint:         "http://endpoint:8080",
		StartTime:        time.Now(),
		StartTimestamp:   time.Now().Unix(),
		RoundTripTime:    100 * time.Millisecond,
		Status:           probe.StatusUp,
		PreStatus:        probe.StatusDown,
		Message:          "dummy message",
		LatestDownTime:   time.Now(),
		RecoveryDuration: 20 * time.Second,
		Stat: probe.Stat{
			Since: time.Now(),
			Total: total,
			Status: map[probe.Status]int64{
				probe.StatusUp:      up,
				probe.StatusDown:    down,
				probe.StatusUnknown: total - up - down,
			},
			UpTime:   time.Duration(up) * time.Second,
			DownTime: time.Duration(down) * time.Second,
		},
	}
	probe.SetResultData(name, &r)
	return r
}

type ProbeFuncType func() (bool, string)

type dummyProber struct {
	base.DefaultProbe
}

func (d *dummyProber) Config(g global.ProbeSettings) error {
	d.DefaultProbe.Config(g, d.ProbeKind, d.ProbeTag, d.ProbeName, "endpoint", d.DoProbe)
	return nil
}
func (d *dummyProber) DoProbe() (bool, string) {
	return rand.Int()%2 == 0, "hello world"
}

func getProbers(n int) []probe.Prober {
	ps := []probe.Prober{}
	for i := 1; i <= n; i++ {
		ps = append(ps, newDummyProber(fmt.Sprintf("prober-%02d", i)))
	}
	setResultData(ps)
	return ps
}
func newDummyProber(name string) probe.Prober {
	r := newDummyResult(name)
	return &dummyProber{
		DefaultProbe: base.DefaultProbe{
			ProbeKind:   "dummy",
			ProbeTag:    "tag",
			ProbeName:   name,
			ProbeResult: &r,
		},
	}
}
func setResultData(probes []probe.Prober) {
	for _, p := range probes {
		probe.SetResultData(p.Name(), p.Result())
	}
}

func TestDiscordConfig(t *testing.T) {
	conf := &NotifyConfig{}
	err := conf.Config(global.NotifySettings{})
	assert.NoError(t, err)
	assert.Equal(t, "discord", conf.Kind())
	assert.Equal(t, conf.Username, global.GetEaseProbe().Name)
	assert.Equal(t, conf.Avatar, global.GetEaseProbe().IconURL)
	assert.Equal(t, conf.Thumbnail, global.GetEaseProbe().IconURL)
}

func TestDiscordDryNotify(t *testing.T) {
	conf := &NotifyConfig{}
	conf.Dry = true
	conf.NotifyName = "dummyDiscord"

	err := conf.Config(global.NotifySettings{})
	assert.NoError(t, err)

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	r := newDummyResult("dummy")
	p := getProbers(4)

	buf.Reset()
	conf.Notify(r)
	assert.Contains(t, buf.String(), "[discord / dummyDiscord] Dry notify")

	buf.Reset()
	conf.NotifyStat(p)
	assert.Contains(t, buf.String(), "[discord / dummyDiscord] Dry notify")
	assert.Contains(t, buf.String(), "**Overall SLA Report (1/1)**")

	monkey.Patch(json.Marshal, func(v interface{}) ([]byte, error) {
		return []byte(""), errors.New("marshal error")
	})
	buf.Reset()
	conf.Notify(r)
	assert.Contains(t, buf.String(), "[discord / dummyDiscord] JSON Marshal Error")

	buf.Reset()
	conf.NotifyStat(p)
	assert.Contains(t, buf.String(), "[discord / dummyDiscord] JSON Marshal Error")

	monkey.UnpatchAll()
}

func TestDiscordNotify(t *testing.T) {
	conf := &NotifyConfig{}
	conf.Dry = false
	conf.NotifyName = "dummyDiscord"
	conf.Retry.Times = 1

	err := conf.Config(global.NotifySettings{})
	assert.NoError(t, err)

	r := newDummyResult("dummy")
	p := getProbers(20)

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	var client *http.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(strings.NewReader(``))
		return &http.Response{
			StatusCode: 204,
			Body:       r,
		}, nil
	})

	buf.Reset()
	conf.Notify(r)
	assert.Contains(t, buf.String(), "[discord / dummyDiscord / Notification] - dummy - successfully sent!")

	buf.Reset()
	conf.NotifyStat(p)
	assert.Contains(t, buf.String(), "[discord / dummyDiscord / SLA] - successfully sent part [1/2]")

	// error response
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(strings.NewReader(`no permission`))
		return &http.Response{
			StatusCode: 403,
			Body:       r,
		}, nil
	})
	buf.Reset()
	r.Status = probe.StatusDown
	conf.Notify(r)
	assert.Contains(t, buf.String(), "no permission")

	// ioutil.ReadAll Error
	monkey.Patch(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
		return nil, errors.New("read error")
	})
	buf.Reset()
	conf.Notify(r)
	assert.Contains(t, buf.String(), "read error")

	// http.Client.Do error
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		return nil, errors.New("http request error")
	})
	conf.Retry.Times = 2
	conf.Retry.Interval = time.Millisecond
	buf.Reset()
	conf.NotifyStat(p)
	assert.Contains(t, buf.String(), "http request error")

	// http.NewRequest error
	monkey.Patch(http.NewRequest, func(method, url string, body io.Reader) (*http.Request, error) {
		return nil, errors.New("http request error")
	})
	buf.Reset()
	conf.Notify(r)
	assert.Contains(t, buf.String(), "http request error")

	// json.Marshal error
	monkey.Patch(json.Marshal, func(v interface{}) ([]byte, error) {
		return nil, errors.New("marshal error")
	})
	buf.Reset()
	conf.Notify(r)
	assert.Contains(t, buf.String(), "marshal error")

	monkey.UnpatchAll()
}

func TestNewField(t *testing.T) {
	conf := &NotifyConfig{}
	r := newDummyResult("dummy")
	f := conf.NewField(r, true)
	assert.Equal(t, "dummy", f.Name)

	f = conf.NewField(r, false)
	assert.Equal(t, "-------------------- dummy --------------------", f.Name)
}

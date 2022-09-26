package ringcentral

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

func TestRingCentral(t *testing.T) {
	conf := &NotifyConfig{}
	conf.NotifyName = "dummy"
	err := conf.Config(global.NotifySettings{})
	assert.NoError(t, err)
	assert.Equal(t, "ringcentral", conf.Kind())
	assert.Equal(t, report.Text, conf.NotifyFormat)

	var client http.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(&client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(strings.NewReader(`{"status":"OK"}`))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})
	err = conf.SendRingCentral("title", "message")
	assert.NoError(t, err)

	monkey.PatchInstanceMethod(reflect.TypeOf(&client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(strings.NewReader(`{"status": "error","message": "Your request was accepted, however a post was not generated","error": "Webhook not found!"}`))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})
	err = conf.SendRingCentral("title", "message")
	assertError(t, err, "Non-ok response returned from RingCentral {\"status\": \"error\",\"message\": \"Your request was accepted, however a post was not generated\",\"error\": \"Webhook not found!\"}")

	monkey.Patch(ioutil.ReadAll, func(_ io.Reader) ([]byte, error) {
		return nil, errors.New("read error")
	})
	err = conf.SendRingCentral("title", "message")
	assertError(t, err, "read error")

	monkey.PatchInstanceMethod(reflect.TypeOf(&client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		return nil, errors.New("http do error")
	})
	err = conf.SendRingCentral("title", "message")
	assertError(t, err, "http do error")

	monkey.Patch(http.NewRequest, func(method string, url string, body io.Reader) (*http.Request, error) {
		return nil, errors.New("new request error")
	})
	err = conf.SendRingCentral("title", "message")
	assertError(t, err, "new request error")

	monkey.UnpatchAll()
}

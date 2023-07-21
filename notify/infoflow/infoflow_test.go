package infoflow

import (
	"errors"
	"io"
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
func TestInfoflow(t *testing.T) {
	conf := &NotifyConfig{}
	err := conf.Config(global.NotifySettings{})
	assert.NoError(t, err)
	assert.Equal(t, report.Infoflow, conf.NotifyFormat)
	assert.Equal(t, "infoflow", conf.Kind())

	var client *http.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := io.NopCloser(strings.NewReader(`{"StatusCode": 0}`))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})
	err = conf.SendInfoflow("title", "message")
	assert.NoError(t, err)

	// bad response
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := io.NopCloser(strings.NewReader(`{"StatusCode": "100"}`))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})
	err = conf.SendInfoflow("title", "message")
	assertError(t, err, "Error response from Infoflow - code [0] - msg []")

	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := io.NopCloser(strings.NewReader(`{"StatusCode": "100", "code": 10, "msg": "infoflow error"}`))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})
	err = conf.SendInfoflow("title", "message")
	assertError(t, err, "Error response from Infoflow - code [10] - msg [infoflow error]")

	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		r := io.NopCloser(strings.NewReader(`bad : json format`))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	})
	err = conf.SendInfoflow("title", "message")
	assertError(t, err, "Error response from Infoflow [200] - [bad : json format]")

	monkey.Patch(io.ReadAll, func(_ io.Reader) ([]byte, error) {
		return nil, errors.New("read error")
	})
	err = conf.SendInfoflow("title", "message")
	assertError(t, err, "read error")

	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Do", func(_ *http.Client, req *http.Request) (*http.Response, error) {
		return nil, errors.New("http error")
	})
	err = conf.SendInfoflow("title", "message")
	assertError(t, err, "http error")

	monkey.Patch(http.NewRequest, func(_ string, _ string, _ io.Reader) (*http.Request, error) {
		return nil, errors.New("new request error")
	})
	err = conf.SendInfoflow("title", "message")
	assertError(t, err, "new request error")

	monkey.UnpatchAll()

}

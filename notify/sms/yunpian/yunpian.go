package yunpian

import (
	"fmt"
	"github.com/megaease/easeprobe/notify/sms/conf"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Kind is the type of driver
const Kind string = "Yunpian"

type Yunpian struct {
	conf.Options `yaml:",inline"`
}

// New create a yunpian sms provider
func New(opt conf.Options) *Yunpian {
	return &Yunpian{
		Options: opt,
	}
}

// Kind return the type of Notify
func (c Yunpian) Kind() string {
	return Kind
}

// Notify return the type of Notify
func (c Yunpian) Notify(title, text string) error {
	api := c.Url

	form := url.Values{}
	form.Add("apikey", c.Key)
	form.Add("mobile", c.Mobile)
	form.Add("text", c.Sign+text)

	log.Debugf("[%s] - API %s - Form %s", c.Kind(), api, form)
	req, err := http.NewRequest(http.MethodPost, api, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Close = true
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error response from Sms [%d] - [%s]", resp.StatusCode, string(buf))
	}
	return nil
}

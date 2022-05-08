package nexmo

import (
	"fmt"
	"github.com/megaease/easeprobe/notify/sms/conf"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Kind is the type of Provider
const Kind string = "Nexmo"

type Nexmo struct {
	conf.Options `yaml:",inline"`
}

// New create a Nexmo sms provider
func New(opt conf.Options) *Nexmo {
	return &Nexmo{
		Options: opt,
	}
}

// Kind return the type of Notify
func (c Nexmo) Kind() string {
	return Kind
}

// Notify return the type of Notify
func (c Nexmo) Notify(title, text string) error {

	api := c.Url

	form := url.Values{}
	form.Add("From", c.From)
	form.Add("To", c.Mobile)
	form.Add("text", text)
	form.Add("api_key", c.Key)
	form.Add("api_secret", c.Secret)

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

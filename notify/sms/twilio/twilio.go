package twilio

import (
	"fmt"
	"github.com/megaease/easeprobe/notify/sms/conf"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Kind is the type of provider
const Kind string = "Twilio"

type Twilio struct {
	conf.Options `yaml:",inline"`

	Sid    string `yaml:"sid"`
	Token  string `yaml:"token"`
	From   string `yaml:"from"`
	ApiUrl string `yaml:"api_url"`
}

// New create a Twilio sms provider
func New(opt conf.Options) *Twilio {
	return &Twilio{
		Options: opt,
	}
}

// Kind return the type of Notify
func (c Twilio) Kind() string {
	return Kind
}

// Notify return the type of Notify
func (c Twilio) Notify(title, text string) error {
	api := c.ApiUrl + c.Sid + "/Messages.json"

	form := url.Values{}
	form.Add("From", c.From)
	form.Add("To", c.Mobile)
	form.Add("text", text)

	log.Debugf("[%s] - API %s - Form %s", c.Kind(), api, form)
	req, err := http.NewRequest(http.MethodPost, api, strings.NewReader(form.Encode()))
	req.SetBasicAuth(c.Sid, c.Token)
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

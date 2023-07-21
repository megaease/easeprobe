package infoflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"
	"github.com/megaease/easeprobe/report"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the infoflow notification configuration
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`
	WebhookURL         string `yaml:"webhook"   json:"webhook" jsonschema:"required,format=uri,title=Webhook URL,description=The Infoflow Robot Webhook URL"`
	GroupID            string `yaml:"group_id"  json:"group_id" jsonschema:"required,format=uri,title=GroupId,description=The Infoflow Robot GroupID"`
}

// Config configures the infoflow notification
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.NotifyKind = "infoflow"
	c.NotifyFormat = report.Infoflow
	c.NotifySendFunc = c.SendInfoflow
	c.DefaultNotify.Config(gConf)
	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.NotifyKind, c.NotifyName, c)
	return nil
}

// SendInfoflow is the wrapper for SendInfoflowNotification
func (c *NotifyConfig) SendInfoflow(title, msg string) error {
	return c.SendInfoflowNotification(msg)
}

// SendInfoflowNotification will post to an 'Robot Webhook' url in Infoflow Apps. It accepts
// some text and the Infoflow robot will send it in group.
func (c *NotifyConfig) SendInfoflowNotification(msg string) error {
	msgContent := fmt.Sprintf(`
	{
		"message":{
			"header":{
				"toid":[%s]
			},
			"body":[
				%s
			]
		}
	}`, c.GroupID, msg)

	req, err := http.NewRequest(http.MethodPost, c.WebhookURL, bytes.NewBuffer([]byte(msgContent)))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Close = true

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	ret := make(map[string]interface{})
	err = json.Unmarshal(buf, &ret)
	if err != nil {
		return fmt.Errorf("Error response from Infoflow [%d] - [%s]", resp.StatusCode, string(buf))
	}
	if statusCode, ok := ret["StatusCode"].(float64); !ok || statusCode != 0 {
		code, _ := ret["code"].(float64)
		msg, _ := ret["msg"].(string)
		return fmt.Errorf("Error response from Infoflow - code [%d] - msg [%v]", int(code), msg)
	}
	return nil
}

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

package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// SNSNotifyConfig is the AWS SNS notification configuration
type SNSNotifyConfig struct {
	Options  `yaml:",inline"`
	TopicARN string          `yaml:"arn"`
	client   *sns.SNS        `yaml:"-"`
	context  context.Context `yaml:"-"`
}

// Kind return the type of Notify
func (c *SNSNotifyConfig) Kind() string {
	return "AWS-SNS"
}

// Config configures the slack notification
func (c *SNSNotifyConfig) Config(gConf global.NotifySettings) error {
	if c.Dry {
		log.Infof("Notification [%s] - [%s]  is running on Dry mode!", c.Kind(), c.Name)
	}

	if err := c.Options.Config(gConf); err != nil {
		return err
	}
	c.client = sns.New(c.session)
	c.context = context.Background()

	log.Infof("[%s] configuration: %+v", c.Kind(), c)
	return nil
}

// Notify write the message into the slack
func (c *SNSNotifyConfig) Notify(result probe.Result) {
	if c.Dry {
		c.DryNotify(result)
		return
	}
	json := result.SlackBlockJSON()
	c.SendNotificationWithRetry("Notification", json)
}

// NotifyStat write the all probe stat message to slack
func (c *SNSNotifyConfig) NotifyStat(probers []probe.Prober) {
	if c.Dry {
		c.DryNotifyStat(probers)
		return
	}
	json := probe.StatSlackBlockJSON(probers)
	c.SendNotificationWithRetry("SLA", json)

}

// DryNotify just log the notification message
func (c *SNSNotifyConfig) DryNotify(result probe.Result) {
	log.Infof("[%s / %s] - %s", c.Kind(), c.Name, result.SlackBlockJSON())
}

// DryNotifyStat just log the notification message
func (c *SNSNotifyConfig) DryNotifyStat(probers []probe.Prober) {
	log.Infof("[%s / %s] - %s", c.Kind(), c.Name, probe.StatSlackBlockJSON(probers))
}

// SendNotificationWithRetry send the SNS notification with retry
func (c *SNSNotifyConfig) SendNotificationWithRetry(tag string, msg string) {

	fn := func() error {
		log.Debugf("[%s - %s] - %s", c.Kind(), tag, msg)
		return c.SendNotification(msg)
	}

	err := global.DoRetry(c.Kind(), c.Name, tag, c.Retry, fn)
	probe.LogSend(c.Kind(), c.Name, tag, "", err)
}

// SendNotification sends the message to SNS
func (c *SNSNotifyConfig) SendNotification(msg string) error {
	ctx, cancel := context.WithTimeout(c.context, c.Timeout)
	defer cancel()

	res, err := c.client.PublishWithContext(ctx, &sns.PublishInput{
		Message:  &msg,
		TopicArn: aws.String(c.TopicARN),
	})
	if err != nil {
		return err
	}
	log.Debugf("[%s / %s] Message ID = %s", c.Kind(), c.Name, *res.MessageId)
	return nil
}

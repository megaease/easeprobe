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
	return c.MyKind
}

// Config configures the slack notification
func (c *SNSNotifyConfig) Config(gConf global.NotifySettings) error {
	c.MyKind = "AWS-SNS"
	if c.Format == 0 {
		c.Format = probe.Text
	}
	c.SendFunc = c.SendSNS

	if err := c.Options.Config(gConf); err != nil {
		return err
	}

	c.client = sns.New(c.session)
	c.context = context.Background()

	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.MyKind, c.Name, c)
	return nil
}

// SendSNS is the warp function of SendSNSNotification
func (c *SNSNotifyConfig) SendSNS(title, msg string) error {
	return c.SendSNSNotification(msg)
}

// SendSNSNotification sends the message to SNS
func (c *SNSNotifyConfig) SendSNSNotification(msg string) error {
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

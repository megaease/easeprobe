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
	"github.com/megaease/easeprobe/report"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the AWS SNS notification configuration
type NotifyConfig struct {
	Options  `yaml:",inline"`
	Format   report.Format   `yaml:"format,omitempty" json:"format,omitempty" jsonschema:"type=string,enum=text,enum=html,enum=markdown,enum=json,title=Format of the Notification,description=Format of the notification,default=text"`
	TopicARN string          `yaml:"arn" json:"arn" jsonschema:"title=Topic ARN,description=The ARN of the SNS topic"`
	client   *sns.SNS        `yaml:"-" json:"-"`
	context  context.Context `yaml:"-" json:"-"`
}

// Config configures the slack notification
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.NotifyKind = "aws-sns"
	if c.Format == report.Unknown {
		c.Format = report.Text
	}
	c.DefaultNotify.NotifyFormat = c.Format
	c.NotifySendFunc = c.SendSNS

	if err := c.Options.Config(gConf); err != nil {
		return err
	}

	c.client = sns.New(c.session)
	c.context = context.Background()

	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.NotifyKind, c.NotifyName, c)
	return nil
}

// SendSNS is the warp function of SendSNSNotification
func (c *NotifyConfig) SendSNS(title, msg string) error {
	return c.SendSNSNotification(msg)
}

// SendSNSNotification sends the message to SNS
func (c *NotifyConfig) SendSNSNotification(msg string) error {
	ctx, cancel := context.WithTimeout(c.context, c.Timeout)
	defer cancel()

	res, err := c.client.PublishWithContext(ctx, &sns.PublishInput{
		Message:  &msg,
		TopicArn: aws.String(c.TopicARN),
	})
	if err != nil {
		return err
	}
	log.Debugf("[%s / %s] Message ID = %s", c.Kind(), c.NotifyName, *res.MessageId)
	return nil
}

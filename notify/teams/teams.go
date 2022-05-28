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

package teams

import (
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"
	"github.com/megaease/easeprobe/report"

	log "github.com/sirupsen/logrus"

	goteamsnotify "github.com/atc0005/go-teams-notify/v2"
)

// NotifyConfig is the teams notification configuration
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`
	WebhookURL         string `yaml:"webhook"`

	client goteamsnotify.API
}

// Config configures the teams notification
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.NotifyKind = "teams"
	c.NotifyFormat = report.MarkdownSocial
	c.NotifySendFunc = c.SendTeamsMessage
	c.DefaultNotify.Config(gConf)

	c.client = goteamsnotify.NewClient()
	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.NotifyKind, c.NotifyName, c)
	return nil
}

// SendTeamsMessage sends the message to the teams channel
func (c *NotifyConfig) SendTeamsMessage(title, msg string) error {
	msgCard := goteamsnotify.NewMessageCard()
	msgCard.Title = title
	msgCard.Text = msg

	return c.client.Send(c.WebhookURL, msgCard)
}

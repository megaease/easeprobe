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

// Package notify contains the notify implementation.
package notify

import (
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/aws"
	"github.com/megaease/easeprobe/notify/dingtalk"
	"github.com/megaease/easeprobe/notify/discord"
	"github.com/megaease/easeprobe/notify/email"
	"github.com/megaease/easeprobe/notify/lark"
	"github.com/megaease/easeprobe/notify/log"
	"github.com/megaease/easeprobe/notify/ringcentral"
	"github.com/megaease/easeprobe/notify/shell"
	"github.com/megaease/easeprobe/notify/slack"
	"github.com/megaease/easeprobe/notify/sms"
	"github.com/megaease/easeprobe/notify/teams"
	"github.com/megaease/easeprobe/notify/telegram"
	"github.com/megaease/easeprobe/notify/wecom"
	"github.com/megaease/easeprobe/probe"
)

// Config is the notify configuration
type Config struct {
	Log         []log.NotifyConfig         `yaml:"log,omitempty" json:"log,omitempty" jsonschema:"title=Log Notification,description=Log Notification Configuration"`
	Email       []email.NotifyConfig       `yaml:"email,omitempty" json:"email,omitempty" jsonschema:"title=Email Notification,description=Email Notification Configuration"`
	Slack       []slack.NotifyConfig       `yaml:"slack,omitempty" json:"slack,omitempty" jsonschema:"title=Slack Notification,description=Slack Notification Configuration"`
	Discord     []discord.NotifyConfig     `yaml:"discord,omitempty" json:"discord,omitempty" jsonschema:"title=Discord Notification,description=Discord Notification Configuration"`
	Telegram    []telegram.NotifyConfig    `yaml:"telegram,omitempty" json:"telegram,omitempty" jsonschema:"title=Telegram Notification,description=Telegram Notification Configuration"`
	AwsSNS      []aws.NotifyConfig         `yaml:"aws_sns,omitempty" json:"aws_sns,omitempty" jsonschema:"title=AWS SNS Notification,description=AWS SNS Notification Configuration"`
	Wecom       []wecom.NotifyConfig       `yaml:"wecom,omitempty" json:"wecom,omitempty" jsonschema:"title=WeCom Notification,description=WeCom Notification Configuration"`
	Dingtalk    []dingtalk.NotifyConfig    `yaml:"dingtalk,omitempty" json:"dingtalk,omitempty" jsonschema:"title=DingTalk Notification,description=DingTalk Notification Configuration"`
	Lark        []lark.NotifyConfig        `yaml:"lark,omitempty" json:"lark,omitempty" jsonschema:"title=Lark Notification,description=Lark Notification Configuration"`
	Sms         []sms.NotifyConfig         `yaml:"sms,omitempty" json:"sms,omitempty" jsonschema:"title=SMS Notification,description=SMS Notification Configuration"`
	Teams       []teams.NotifyConfig       `yaml:"teams,omitempty" json:"teams,omitempty" jsonschema:"title=Teams Notification,description=Teams Notification Configuration"`
	Shell       []shell.NotifyConfig       `yaml:"shell,omitempty" json:"shell,omitempty" jsonschema:"title=Shell Notification,description=Shell Notification Configuration"`
	RingCentral []ringcentral.NotifyConfig `yaml:"ringcentral,omitempty" json:"ringcentral,omitempty" jsonschema:"title=RingCentral Notification,description=RingCentral Notification Configuration"`
}

// Notify is the configuration of the Notify
type Notify interface {
	Kind() string
	Name() string
	Channels() []string
	Config(global.NotifySettings) error
	Notify(probe.Result)
	NotifyStat([]probe.Prober)

	DryNotify(probe.Result)
	DryNotifyStat([]probe.Prober)
}

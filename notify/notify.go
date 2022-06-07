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

package notify

import (
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/aws"
	"github.com/megaease/easeprobe/notify/dingtalk"
	"github.com/megaease/easeprobe/notify/discord"
	"github.com/megaease/easeprobe/notify/email"
	"github.com/megaease/easeprobe/notify/lark"
	"github.com/megaease/easeprobe/notify/log"
	"github.com/megaease/easeprobe/notify/shell"
	"github.com/megaease/easeprobe/notify/slack"
	"github.com/megaease/easeprobe/notify/sms"
	"github.com/megaease/easeprobe/notify/teams"
	"github.com/megaease/easeprobe/notify/telegram"
	"github.com/megaease/easeprobe/notify/wecom"
	"github.com/megaease/easeprobe/probe"
)

//Config is the notify configuration
type Config struct {
	Log      []log.NotifyConfig      `yaml:"log"`
	Email    []email.NotifyConfig    `yaml:"email"`
	Slack    []slack.NotifyConfig    `yaml:"slack"`
	Discord  []discord.NotifyConfig  `yaml:"discord"`
	Telegram []telegram.NotifyConfig `yaml:"telegram"`
	AwsSNS   []aws.SNSNotifyConfig   `yaml:"aws_sns"`
	Wecom    []wecom.NotifyConfig    `yaml:"wecom"`
	Dingtalk []dingtalk.NotifyConfig `yaml:"dingtalk"`
	Lark     []lark.NotifyConfig     `yaml:"lark"`
	Sms      []sms.NotifyConfig      `yaml:"sms"`
	Teams    []teams.NotifyConfig    `yaml:"teams"`
	Shell    []shell.NotifyConfig    `yaml:"shell"`
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

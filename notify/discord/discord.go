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

// Package discord is the notification for Discord
package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/report"
	log "github.com/sirupsen/logrus"
)

// Refer to:
// - Documents: https://birdie0.github.io/discord-webhooks-guide/index.html
// - Using https://discohook.org/ to preview

// Limitation - https://discordjs.guide/popular-topics/embeds.html#embed-limits
// Embed titles are limited to 256 characters
// Embed descriptions are limited to 4096 characters
// There can be up to 25 fields
// A field's name is limited to 256 characters and its value to 1024 characters
// The footer text is limited to 2048 characters
// The author name is limited to 256 characters
// The sum of all characters from all embed structures in a message must not exceed 6000 characters
// 10 embeds can be sent per message

// Thumbnail use thumbnail in the embed. You can set only url of the thumbnail.
// There is no way to set width/height of the picture.
type Thumbnail struct {
	URL string `json:"url"`
}

// Fields allows you to use multiple title + description blocks in embed.
// fields is an array of field objects. Each object includes three values:
type Fields struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

// Footer allows you to add footer to embed. footer is an object which includes two values:
//   - text - sets name for author object. Markdown is disabled here!!!
//   - icon_url - sets icon for author object. Requires text value.
type Footer struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}

// Author is an object which includes three values:
// - name - sets name.
// - url - sets link. Requires name value. If used, transforms name into hyperlink.
// - icon_url - sets avatar. Requires name value.
type Author struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	IconURL string `json:"icon_url"`
}

// Embed is custom embeds for message sent by webhook.
// embeds is an array of embeds and can contain up to 10 embeds in the same message.
type Embed struct {
	Author      Author    `json:"author"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Color       int       `json:"color"`
	Description string    `json:"description"`
	Timestamp   string    `json:"timestamp"` //"YYYY-MM-DDTHH:MM:SS.MSSZ"
	Thumbnail   Thumbnail `json:"thumbnail"`
	Fields      []Fields  `json:"fields"`
	Footer      Footer    `json:"footer"`
}

// Discord is the struct for all of the discrod json.
type Discord struct {
	Username  string  `json:"username"`
	AvatarURL string  `json:"avatar_url"`
	Content   string  `json:"content"`
	Embeds    []Embed `json:"embeds"`
}

// NotifyConfig is the slack notification configuration
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`
	Username           string `yaml:"username,omitempty" json:"username,omitempty" jsonschema:"title=Username,description=Discord Username for the notification"`
	WebhookURL         string `yaml:"webhook" json:"webhook" jsonschema:"format=uri,title=Webhook URL,description=Discord Webhook URL for the notification"`
	Avatar             string `yaml:"avatar,omitempty" json:"avatar,omitempty" jsonschema:"format=uri,title=Avatar,description=Discord Avatar for the notification,example=https://example.com/avatar.png"`
	Thumbnail          string `yaml:"thumbnail,omitempty" json:"thumbnail,omitempty" jsonschema:"format=uri,title=Thumbnail,description=Discord Thumbnail for the notification,example=https://example.com/thumbnail.png"`
}

// Config configures the log files
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.NotifyKind = "discord"
	c.DefaultNotify.Config(gConf)

	if len(strings.TrimSpace(c.Username)) <= 0 {
		c.Username = global.GetEaseProbe().Name
	}

	if len(strings.TrimSpace(c.Avatar)) <= 0 {
		c.Avatar = global.GetEaseProbe().IconURL
	}

	if len(strings.TrimSpace(c.Thumbnail)) <= 0 {
		c.Thumbnail = global.GetEaseProbe().IconURL
	}

	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.NotifyKind, c.NotifyName, c)
	return nil
}

// NewDiscord new a discord object from a result
func (c *NotifyConfig) NewDiscord(result probe.Result) Discord {
	discord := Discord{
		Username:  c.Username,
		AvatarURL: c.Avatar,
		Content:   "",
		Embeds:    []Embed{},
	}

	// using https://www.spycolor.com/ to pick color
	color := 1091331 //"#10a703" - green
	if result.Status != probe.StatusUp {
		color = 10945283 // "#a70303" - red
	}

	rtt := result.RoundTripTime.Round(time.Millisecond)
	description := fmt.Sprintf("%s %s - ⏱ %s\n```%s```",
		result.Status.Emoji(), result.Endpoint, rtt, result.Message)

	discord.Embeds = append(discord.Embeds, Embed{
		Author:      Author{},
		Title:       result.Title(),
		URL:         "",
		Color:       color,
		Description: description,
		Timestamp:   result.StartTime.UTC().Format(time.RFC3339),
		Thumbnail:   Thumbnail{URL: c.Thumbnail},
		Fields:      []Fields{},
		Footer: Footer{
			Text:    global.FooterString(),
			IconURL: global.GetEaseProbe().IconURL,
		},
	})
	return discord
}

// Notify write the message into the slack
func (c *NotifyConfig) Notify(result probe.Result) {
	if c.Dry {
		c.DryNotify(result)
		return
	}

	discord := c.NewDiscord(result)
	tag := "Notification"

	fn := func() error {
		return c.SendDiscordNotification(discord, tag)
	}

	err := global.DoRetry(c.Kind(), c.NotifyName, tag, c.Retry, fn)
	report.LogSend(c.Kind(), c.NotifyName, tag, result.Name, err)
}

// NewEmbed new a embed object from a result
func (c *NotifyConfig) NewEmbed() Embed {
	return Embed{
		Author:      Author{},
		Title:       "",
		URL:         "",
		Color:       239, // #0000ef - blue
		Description: "",
		Timestamp:   "",
		Thumbnail:   Thumbnail{},
		Fields:      []Fields{},
		Footer:      Footer{},
	}
}

// NewField new a Field object from a result
func (c *NotifyConfig) NewField(result probe.Result, inline bool) Fields {
	message := "%s\n" +
		"**Availability**\n>\t" + " **Up**:  `%s`  **Down** `%s`  -  **SLA**: `%.2f %%`" +
		"\n**Probe Times**\n>\t**Total** : `%d` ( %s )" +
		"\n**Latest Probe**\n>\t%s | %s" +
		"\n>\t`%s ` \n\n"

	timeFmt := global.GetTimeFormat()
	timeLoc := global.GetTimeLocation()
	desc := fmt.Sprintf(message, result.Endpoint,
		report.DurationStr(result.Stat.UpTime), report.DurationStr(result.Stat.DownTime), result.SLAPercent(),
		result.Stat.Total, report.SLAStatusText(result.Stat, report.Markdown),
		result.StartTime.In(timeLoc).Format(timeFmt), result.Status.Emoji()+" "+result.Status.String(),
		result.Message)

	line := ""
	name := result.Name
	if !inline {
		len := (45 - len(result.Name)) / 2
		if len > 0 {
			line = strings.Repeat("-", len)
		}
		name = fmt.Sprintf("%s %s %s", line, result.Name, line)
	}
	return Fields{
		Name:   name,
		Value:  desc,
		Inline: inline,
	}
}

// NewEmbeds return a discord with multiple Embed
func (c *NotifyConfig) NewEmbeds(probers []probe.Prober) []Discord {
	var discords []Discord

	//every page has 12 probe result
	const pageCnt = 12
	total := len(probers)
	//calculate how many page we need
	pages := total / pageCnt
	if total%pageCnt > 0 {
		pages++
	}

	for p := 0; p < pages; p++ {
		discord := Discord{
			Username:  c.Username,
			AvatarURL: c.Avatar,
			Content:   fmt.Sprintf("**Overall SLA Report (%d/%d)**", p+1, pages),
			Embeds:    []Embed{},
		}

		discord.Embeds = append(discord.Embeds, c.NewEmbed())

		//calculate the current page start and end position
		start := p * pageCnt
		end := (p + 1) * pageCnt
		if len(probers) < end {
			end = len(probers)
		}
		for i := start; i < end; i++ {
			//discord.Embeds = append(discord.Embeds, c.NewEmbed(*probers[i].Result()))
			discord.Embeds[0].Fields = append(discord.Embeds[0].Fields,
				c.NewField(*probers[i].Result(), true))
		}
		discords = append(discords, discord)
	}

	return discords
}

// NotifyStat write the all probe stat message to slack
func (c *NotifyConfig) NotifyStat(probers []probe.Prober) {
	if c.Dry {
		c.DryNotifyStat(probers)
		return
	}
	tag := "SLA"
	discords := c.NewEmbeds(probers)
	total := len(discords)
	for idx, discord := range discords {

		fn := func() error {
			return c.SendDiscordNotification(discord, tag)
		}

		err := global.DoRetry(c.Kind(), c.NotifyName, tag, c.Retry, fn)
		if err != nil {
			log.Errorf("[%s / %s / %s] - failed to send part [%d/%d]! (%v)", c.Kind(), c.Name(), tag, idx+1, total, err)
		} else {
			log.Infof("[%s / %s / %s] - successfully sent part [%d/%d]!", c.Kind(), c.Name(), tag, idx+1, total)
		}

	}
}

// DryNotify just log the notification message
func (c *NotifyConfig) DryNotify(result probe.Result) {
	discord := c.NewDiscord(result)
	json, err := json.Marshal(discord)
	if err != nil {
		log.Errorf("[%s / %s] JSON Marshal Error : %v", c.Kind(), c.NotifyName, err)
		return
	}
	log.Infof("[%s / %s] Dry notify - %s", c.Kind(), c.NotifyName, string(json))
}

// DryNotifyStat just log the notification message
func (c *NotifyConfig) DryNotifyStat(probers []probe.Prober) {
	discord := c.NewEmbeds(probers)
	json, err := json.Marshal(discord)
	if err != nil {
		log.Errorf("[%s / %s] JSON Marshal Error : %v", c.Kind(), c.NotifyName, err)
		return
	}
	log.Infof("[%s / %s] Dry notify - %s", c.Kind(), c.NotifyName, string(json))
}

// SendDiscordNotification will post to an 'Incoming Webhook' url setup in Discrod Apps.
func (c *NotifyConfig) SendDiscordNotification(discord Discord, tag string) error {
	json, err := json.Marshal(discord)
	if err != nil {
		log.Errorf("[%s / %s / %s] - %v, err - %s", c.Kind(), c.Name(), tag, discord, err)
		return &global.ErrNoRetry{Message: err.Error()}
	}
	log.Debugf("[%s / %s/ %s] - %s", c.Kind(), c.Name(), tag, string(json))

	req, err := http.NewRequest(http.MethodPost, c.WebhookURL, bytes.NewBuffer([]byte(json)))
	if err != nil {
		return err
	}
	req.Close = true
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return &global.ErrNoRetry{Message: fmt.Sprintf("Error response from Discord with request body <%s> [%d] - [%s]", json, resp.StatusCode, string(buf))}
	}
	return nil
}

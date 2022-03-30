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

package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
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
//  - text - sets name for author object. Markdown is disabled here!!!
//  - icon_url - sets icon for author object. Requires text value.
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
	Name       string        `yaml:"name"`
	WebhookURL string        `yaml:"webhook"`
	Avatar     string        `yaml:"avatar"`
	Thumbnail  string        `yaml:"thumbnail"`
	Dry        bool          `yaml:"dry"`
	Timeout    time.Duration `yaml:"timeout"`
	Retry      global.Retry  `yaml:"retry"`
}

// Kind return the type of Notify
func (c *NotifyConfig) Kind() string {
	return "discord"
}

// Config configures the log files
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {

	if c.Dry {
		log.Infof("Notification [%s] - [%s]  is running on Dry mode!", c.Kind(), c.Name)
	}

	if len(strings.TrimSpace(c.Avatar)) <= 0 {
		c.Avatar = global.Icon
	}

	if len(strings.TrimSpace(c.Thumbnail)) <= 0 {
		c.Thumbnail = global.Icon
	}

	c.Timeout = gConf.NormalizeTimeOut(c.Timeout)
	c.Retry = gConf.NormalizeRetry(c.Retry)

	log.Infof("[%s] configuration: %+v", c.Kind(), c)

	return nil
}

// NewDiscord new a discord object from a result
func (c *NotifyConfig) NewDiscord(result probe.Result) Discord {
	discord := Discord{
		Username:  global.Prog,
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
		Footer:      Footer{Text: "Probed at", IconURL: global.Icon},
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
		json, err := json.Marshal(discord)
		if err != nil {
			log.Debugf("[%s - %s] - %v", c.Kind(), tag, discord)
		} else {
			log.Debugf("[%s - %s ] - %s", c.Kind(), tag, string(json))
		}
		return c.SendDiscordNotification(discord)
	}

	err := global.DoRetry(c.Kind(), c.Name, tag, c.Retry, fn)
	probe.LogSend(c.Kind(), c.Name, tag, result.Name, err)
}

// NewEmbed new a embed object from a result
func (c *NotifyConfig) NewEmbed(result probe.Result) Embed {
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

	desc := fmt.Sprintf(message, result.Endpoint,
		result.Stat.UpTime.Round(time.Second), result.Stat.DownTime.Round(time.Second), result.SLA(),
		result.Stat.Total, probe.StatStatusText(result.Stat, probe.Makerdown),
		result.StartTime.UTC().Format(result.TimeFormat), result.Status.Emoji()+" "+result.Status.String(),
		result.Message)

	line := ""
	if !inline {
		len := (45 - len(result.Name)) / 2
		if len > 0 {
			line = strings.Repeat("-", len)
		}
	}
	return Fields{
		Name:   fmt.Sprintf("%s %s %s", line, result.Name, line),
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
			Username:  global.Prog,
			AvatarURL: c.Avatar,
			Content:   fmt.Sprintf("**Overall SLA Report (%d/%d)**", p+1, pages),
			Embeds:    []Embed{},
		}

		discord.Embeds = append(discord.Embeds, Embed{})

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
			json, err := json.Marshal(discord)
			if err != nil {
				log.Debugf("[%s - %s ] - %v", c.Kind(), tag, discord)
			} else {
				log.Debugf("[%s - %s ] - %s", c.Kind(), tag, string(json))
			}
			return c.SendDiscordNotification(discord)
		}

		err := global.DoRetry(c.Kind(), c.Name, tag, c.Retry, fn)
		if err != nil {
			log.Errorf("[%s / %s / %s] - failed to send part [%d/%d]! (%v)", c.Kind(), c.Name, tag, idx+1, total, err)
		} else {
			log.Infof("[%s / %s / %s] - successfully sent part [%d/%d]!", c.Kind(), c.Name, tag, idx+1, total)
		}

	}
}

// DryNotify just log the notification message
func (c *NotifyConfig) DryNotify(result probe.Result) {
	discord := c.NewDiscord(result)
	json, err := json.Marshal(discord)
	if err != nil {
		log.Errorf("error : %v", err)
		return
	}
	log.Infof("[%s / %s ] Dry notify - %s", c.Kind(), c.Name, string(json))
}

// DryNotifyStat just log the notification message
func (c *NotifyConfig) DryNotifyStat(probers []probe.Prober) {
	discord := c.NewEmbeds(probers)
	json, err := json.Marshal(discord)
	if err != nil {
		log.Errorf("error : %v", err)
		return
	}
	log.Infof("[%s / %s ] Dry notify - %s", c.Kind(), c.Name, string(json))
}

// SendDiscordNotification will post to an 'Incoming Webhook' url setup in Discrod Apps.
func (c *NotifyConfig) SendDiscordNotification(discord Discord) error {
	json, err := json.Marshal(discord)
	if err != nil {
		return err
	}
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

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return fmt.Errorf("Error response from Discord [%d] - [%s]", resp.StatusCode, string(buf))
	}
	return nil
}

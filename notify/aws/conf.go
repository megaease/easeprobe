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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/megaease/easeprobe/global"
)

// Credentials is AWS access id and access token
type Credentials struct {
	ID     string `yaml:"id"`
	Secret string `yaml:"key"`
}

// Options is AWS Configuration
type Options struct {
	Name        string        `yaml:"name"`
	Region      string        `yaml:"region"`
	Endpoint    string        `yaml:"endpoint"`
	Credentials Credentials   `yaml:"credential"`
	Profile     string        `yaml:"profile"`
	Dry         bool          `yaml:"dry"`
	Timeout     time.Duration `yaml:"timeout"`
	Retry       global.Retry  `yaml:"retry"`

	session *session.Session `yaml:"-"`
}

// Config config a AWS configuration
func (conf *Options) Config(gConf global.NotifySettings) error {

	conf.Timeout = gConf.NormalizeTimeOut(conf.Timeout)
	conf.Retry = gConf.NormalizeRetry(conf.Retry)

	session, err := session.NewSessionWithOptions(
		session.Options{
			Config: aws.Config{
				Credentials: credentials.NewStaticCredentials(
					conf.Credentials.ID,
					conf.Credentials.Secret,
					"",
				),
				Region:   aws.String(conf.Region),
				Endpoint: aws.String(conf.Endpoint),
			},
			Profile: conf.Profile,
		},
	)

	if err != nil {
		return err
	}
	conf.session = session
	return nil

}

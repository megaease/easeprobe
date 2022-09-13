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

// Package aws is the AWS notification package
package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"
)

// Credentials is AWS access id and access token
type Credentials struct {
	ID     string `yaml:"id" json:"id" jsonschema:"required,title=AWS Access Key ID,description=AWS Access Key ID"`
	Secret string `yaml:"key" json:"key" jsonschema:"required,title=AWS Access Key Secret,description=AWS Access Key Secret"`
}

// Options is AWS Configuration
type Options struct {
	base.DefaultNotify `yaml:",inline"`
	Region             string      `yaml:"region" json:"region" jsonschema:"required,title=AWS Region ID,description=AWS Region ID,example=\"us-west-2\""`
	Endpoint           string      `yaml:"endpoint" json:"endpoint" jsonschema:"required,title=AWS Endpoint,description=AWS Endpoint,example=\"https://sns.us-west-2.amazonaws.com\""`
	Credentials        Credentials `yaml:"credential" json:"credential" jsonschema:"required,title=AWS Credential,description=AWS Credential"`
	Profile            string      `yaml:"profile,omitempty" json:"profile,omitempty" jsonschema:"title=AWS Profile,description=AWS Profile"`

	session *session.Session `yaml:"-"`
}

// Config config a AWS configuration
func (conf *Options) Config(gConf global.NotifySettings) error {

	conf.DefaultNotify.Config(gConf)

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

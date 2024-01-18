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
	"errors"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/report"
	"github.com/stretchr/testify/assert"
)

func TestSNSConfig(t *testing.T) {
	conf := &NotifyConfig{}
	err := conf.Config(global.NotifySettings{})
	assert.NoError(t, err)
	assert.Equal(t, "aws-sns", conf.Kind())
	assert.Equal(t, report.Text, conf.Format)
	assert.NotNil(t, conf, conf.client)
	assert.NotNil(t, conf, conf.context)

	monkey.Patch(session.NewSessionWithOptions, func(options session.Options) (*session.Session, error) {
		return nil, errors.New("session error")
	})
	err = conf.Config(global.NotifySettings{})
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "session error")

	monkey.PatchInstanceMethod(reflect.TypeOf(conf.client), "PublishWithContext", func(_ *sns.SNS, ctx context.Context, input *sns.PublishInput, opts ...request.Option) (*sns.PublishOutput, error) {
		id := "123"
		num := "ab"
		return &sns.PublishOutput{
			MessageId:      &id,
			SequenceNumber: &num,
		}, nil
	})
	err = conf.SendSNSNotification("test")
	assert.NoError(t, err)

	monkey.PatchInstanceMethod(reflect.TypeOf(conf.client), "PublishWithContext", func(_ *sns.SNS, ctx context.Context, input *sns.PublishInput, opts ...request.Option) (*sns.PublishOutput, error) {
		return nil, errors.New("publish error")
	})

	err = conf.SendSNS("title", "msg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "publish error")

	monkey.UnpatchAll()
}

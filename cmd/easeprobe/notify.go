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

package main

import (
	"github.com/megaease/easeprobe/conf"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify"
	log "github.com/sirupsen/logrus"
)

func configNotifiers(notifies []notify.Notify) {
	gNotifyConf := global.NotifySettings{
		TimeFormat: conf.Get().Settings.TimeFormat,
		Retry:      conf.Get().Settings.Notify.Retry,
	}
	for _, n := range notifies {
		err := n.Config(gNotifyConf)
		if err != nil {
			log.Errorf("error: %v", err)
			continue
		}
		log.Infof("Successfully setup the notify channel: %s", n.Kind())
	}
}

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

package web

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/megaease/easeprobe/conf"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/report"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
)

var probers *[]probe.Prober

func slaHTML(w http.ResponseWriter, req *http.Request) {
	interval := conf.Get().Settings.HTTPServer.AutoRefreshTime
	refresh := fmt.Sprintf("%d", interval.Milliseconds())
	html := []byte(report.SLAHTML(*probers) + report.AutoRefreshJS(refresh))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(html)
}

func slaJSON(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(report.SLAJSON(*probers)))
}

// SetProbers set the probers
func SetProbers(p []probe.Prober) {
	probers = &p
}

// Server is the http server
func Server() {

	c := conf.Get()
	host := c.Settings.HTTPServer.IP
	port := c.Settings.HTTPServer.Port

	// Configure the http server
	if len(host) > 0 && net.ParseIP(host) == nil {
		host = global.DefaultHTTPServerIP
	}
	p, err := strconv.Atoi(port)
	if err != nil || p <= 1024 || p > 65535 {
		log.Warnf("[Web] Invalid port number: %s, use the default value: %s", port, global.DefaultHTTPServerPort)
		port = global.DefaultHTTPServerPort
	} else {
		port = c.Settings.HTTPServer.Port
	}

	// Configure the auto refresh time of the SLA page
	if c.Settings.HTTPServer.AutoRefreshTime == 0 {
		interval := global.DefaultProbeInterval
		// find the minimum probe interval
		for p := range *probers {
			if interval > (*probers)[p].Interval() {
				interval = (*probers)[p].Interval()
			}
		}
		c.Settings.HTTPServer.AutoRefreshTime = interval
		log.Debugf("[Web] Auto refresh interval time: %s", interval)
	}

	// Prepare the router
	r := chi.NewRouter()

	filename := c.Settings.HTTPServer.AccessLog.File
	if len(filename) > 0 {
		log.Infof("[Web] Access Log output file: %s", filename)
		logger := c.Settings.HTTPServer.AccessLog.Logger
		r.Use(NewStructuredLogger(logger))
	}

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware.StripSlashes)

	r.Get("/", slaHTML)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/sla", slaJSON)
	})

	r.NotFound(slaHTML)

	server, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		log.Fatalf("[Web] Failed to start the http server: %s", err)
	}
	log.Infof("[Web] HTTP server is listening on %s:%s", host, port)

	// Start the http server
	go func() {
		if err := http.Serve(server, r); err != nil {
			log.Errorf("[Web] HTTP server error: %s", err)
		}
	}()

}

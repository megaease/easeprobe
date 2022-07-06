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
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/megaease/easeprobe/conf"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/report"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
)

var probers *[]probe.Prober

func getRefreshInterval(refersh string) time.Duration {
	interval := conf.Get().Settings.HTTPServer.AutoRefreshTime
	if strings.TrimSpace(refersh) == "" {
		return interval
	}
	r, err := time.ParseDuration(refersh)
	if err != nil {
		log.Errorf("[Web] Invalid refresh time: %s", err)
		return interval
	}
	return r
}

func getStatus(status string) *probe.Status {
	if status == "" {
		return nil
	}
	var s probe.Status
	s.Status(status)
	return &s
}

func getFloat(f string, _default float64) float64 {
	if f == "" {
		return _default
	}
	flt, err := strconv.ParseFloat(f, 64)
	if err != nil {
		log.Debugf("[Web] Invalid float value: %s", err)
		return _default
	}
	return flt
}

func getInt(i string, _default int) int {
	if i == "" {
		return _default
	}
	it, err := strconv.Atoi(i)
	if err != nil {
		log.Debugf("[Web] Invalid int value: %s", err)
		return _default
	}
	return it
}

func getFilter(req *http.Request) (*report.SLAFilter, error) {
	filter := &report.SLAFilter{}

	filter.Name = strings.TrimSpace(req.URL.Query().Get("name"))
	filter.Kind = strings.TrimSpace(req.URL.Query().Get("kind"))
	filter.Endpoint = strings.TrimSpace(req.URL.Query().Get("ep"))
	filter.Status = getStatus(req.URL.Query().Get("status"))
	filter.Message = strings.TrimSpace(req.URL.Query().Get("msg"))
	filter.SLAGreater = getFloat(req.URL.Query().Get("gte"), 0)
	filter.SLALess = getFloat(req.URL.Query().Get("lte"), 100)
	filter.PageNum = getInt(req.URL.Query().Get("pg"), 1)
	filter.PageSize = getInt(req.URL.Query().Get("sz"), global.DefaultPageSize)

	if err := filter.Check(); err != nil {
		log.Errorf(err.Error())
		return nil, err
	}
	return filter, nil
}

func slaHTML(w http.ResponseWriter, req *http.Request) {
	interval := getRefreshInterval(req.URL.Query().Get("refresh"))

	filter, err := getFilter(req)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	refresh := fmt.Sprintf("%d", interval.Milliseconds())
	html := []byte(report.SLAHTMLFilter(*probers, filter) + report.AutoRefreshJS(refresh))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(html)
}

func slaJSON(w http.ResponseWriter, req *http.Request) {
	filter, err := getFilter(req)
	if err != nil {
		buf, e := json.Marshal(err)
		if e != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(buf)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_probers := filter.Filter(*probers)
	w.Write([]byte(report.SLAJSON(_probers)))
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
		c.Settings.HTTPServer.AutoRefreshTime = global.DefaultProbeInterval
	}
	log.Debugf("[Web] Auto refresh interval time: %s", c.Settings.HTTPServer.AutoRefreshTime)

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

	r.Get("/metrics", promhttp.Handler().ServeHTTP)

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

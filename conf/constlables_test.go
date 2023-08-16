package conf

import (
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/megaease/easeprobe/probe/http"
	"github.com/megaease/easeprobe/probe/tcp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMergeConstLabels(t *testing.T) {

	ps := []probe.Prober{
		&http.HTTP{
			DefaultProbe: base.DefaultProbe{
				Labels: prometheus.Labels{"service": "service_a"},
			},
		},
		&tcp.TCP{
			DefaultProbe: base.DefaultProbe{
				Labels: prometheus.Labels{"host": "host_b"},
			},
		},
	}

	MergeConstLabels(ps)

	assert.Equal(t, 2, len(ps[0].LabelMap()))
	assert.Equal(t, "service_a", ps[0].LabelMap()["service"])
	assert.Equal(t, "", ps[0].LabelMap()["host"])

	assert.Equal(t, 2, len(ps[1].LabelMap()))
	assert.Equal(t, "", ps[1].LabelMap()["service"])
	assert.Equal(t, "host_b", ps[1].LabelMap()["host"])
}

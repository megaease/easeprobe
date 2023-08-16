package conf

import (
	"github.com/megaease/easeprobe/metric"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/megaease/easeprobe/probe/http"
	"github.com/megaease/easeprobe/probe/tcp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMergeConstLabels(t *testing.T) {

	ps := []probe.Prober{
		&http.HTTP{
			DefaultProbe: base.DefaultProbe{
				Labels: metric.LabelMap{"service": "service_a"},
			},
		},
		&tcp.TCP{
			DefaultProbe: base.DefaultProbe{
				Labels: metric.LabelMap{"host": "host_b"},
			},
		},
	}

	MergeConstLabels(ps)

	assert.Equal(t, 2, len(ps[0].LabelMap()))
	assert.Equal(t, "service_a", ps[0].LabelMap()["service"])
	assert.Equal(t, "", ps[0].LabelMap()["host"])

	assert.Equal(t, 2, len(ps[1].LabelMap()))
	assert.Equal(t, "", ps[0].LabelMap()["service"])
	assert.Equal(t, "host_b", ps[0].LabelMap()["host"])
}

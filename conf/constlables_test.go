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
				Labels: []metric.Label{{"service", "service_a"}},
			},
		},
		&tcp.TCP{
			DefaultProbe: base.DefaultProbe{
				Labels: []metric.Label{{"host", "host_b"}},
			},
		},
	}

	MergeConstLabels(ps)

	assert.Equal(t, "service_a", ps[0].Label()[0].Value)
	assert.Equal(t, "host_b", ps[0].Label()[1].Value)

	assert.Equal(t, "host_a", ps[1].Label()[0].Value)
	assert.Equal(t, "service_b", ps[1].Label()[1].Value)
}

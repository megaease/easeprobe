package conf

import (
	"github.com/megaease/easeprobe/metric"
	"github.com/megaease/easeprobe/probe"
)

var constLabels = make(map[string]bool)

// MergeConstLabels merge const labels from all probers.
// Prometheus requires all metric  of the same name have the same set of labels in a collector
func MergeConstLabels(ps []probe.Prober) {
	for _, p := range ps {
		for _, k := range p.Label() {
			constLabels[k.Name] = true
		}
	}

	for _, p := range ps {
		buildConstLabels(p)
	}
}
func buildConstLabels(p probe.Prober) {
	ls := p.Label()
LA:
	for k, _ := range constLabels {
		for _, l := range ls {
			if k == l.Name {
				continue LA
			}
		}

		ls = append(ls, metric.Label{Name: k, Value: ""})
	}

	p.SetLabel(ls)
}

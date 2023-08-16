package conf

import (
	"github.com/megaease/easeprobe/probe"
)

var constLabels = make(map[string]bool)

// MergeConstLabels merge const labels from all probers.
// Prometheus requires all metric  of the same name have the same set of labels in a collector
func MergeConstLabels(ps []probe.Prober) {
	for _, p := range ps {
		for k, _ := range p.LabelMap() {
			constLabels[k] = true
		}
	}

	for _, p := range ps {
		buildConstLabels(p)
	}
}
func buildConstLabels(p probe.Prober) {
	ls := p.LabelMap()
	if ls == nil {
		ls = make(map[string]string)
	}

	for k, _ := range constLabels {
		if _, ok := ls[k]; !ok {
			ls[k] = ""
		}
	}
}

package conf

import (
	"github.com/megaease/easeprobe/probe"
)

// MergeConstLabels merge const labels from all probers.
//
//	Prometheus requires all metric  of the same name have the same set of labels in a collector
func MergeConstLabels(ps []probe.Prober) {
	var constLabels = make(map[string]bool)
	for _, p := range ps {
		for k, _ := range p.LabelMap() {
			constLabels[k] = true
		}
	}

	for _, p := range ps {
		buildConstLabels(p, constLabels)
	}
}

func buildConstLabels(p probe.Prober, constLabels map[string]bool) {
	ls := p.LabelMap()
	if ls == nil {
		ls = make(map[string]string)
		p.SetLabelMap(ls)
	}

	for k, _ := range constLabels {
		if _, ok := ls[k]; !ok {
			ls[k] = ""
		}
	}
}

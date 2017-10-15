package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/DNS-OARC/ripeatlas"
	"github.com/DNS-OARC/ripeatlas/measurement"
	"github.com/r3ek0/atlas_exporter/dns"
	"github.com/r3ek0/atlas_exporter/http"
	"github.com/r3ek0/atlas_exporter/metric"
	"github.com/r3ek0/atlas_exporter/ntp"
	"github.com/r3ek0/atlas_exporter/ping"
	"github.com/r3ek0/atlas_exporter/probe"
	"github.com/r3ek0/atlas_exporter/sslcert"
	"github.com/r3ek0/atlas_exporter/traceroute"
	"github.com/prometheus/common/log"
)

var cache *probe.ProbeCache

func initCache() {
	cache = probe.NewCache(time.Duration(*cacheTtl) * time.Second)
	startCacheCleanupFunc(time.Duration(*cacheCleanUp) * time.Second)
}

func startCacheCleanupFunc(d time.Duration) {
	go func() {
		for {
			select {
			case <-time.After(d):
				log.Infoln("Cleaning up cache...")
				r := cache.CleanUp()
				log.Infof("Items removed: %d", r)
			}
		}
	}()
}

func getMeasurement(id string) ([]metric.MetricExporter, error) {
	a := ripeatlas.Atlaser(ripeatlas.NewHttp())
	c, err := a.MeasurementLatest(ripeatlas.Params{"pk": id})

	if err != nil {
		return nil, err
	}

	res := make([]metric.MetricExporter, 0)
	ch := make(chan metric.MetricExporter)

	count := 0
	for r := range c {
		if r.ParseError != nil {
			return nil, err
		}

		go getMetricExporter(r, ch)
		count++
	}

	for i := 0; i < count; i++ {
		select {
		case m := <-ch:
			if m != nil && (!*filterInvalidResults || m.IsValid()) {
				res = append(res, m)
			}
		case <-time.After(60 * time.Second):
			return nil, errors.New(fmt.Sprintln("Timeout exceeded!"))
		}
	}

	return res, nil
}

func getProbeList() []probe.Probe {
	return cache.GetAll()
}

func getMetricExporter(r *measurement.Result, out chan metric.MetricExporter) {
	var m metric.MetricExporter

	switch r.Type() {
	case "ping":
		m = ping.FromResult(r)
	case "traceroute":
		m = traceroute.FromResult(r)
	case "ntp":
		m = ntp.FromResult(r)
	case "dns":
		m = dns.FromResult(r)
	case "http":
		m = http.FromResult(r)
	case "sslcert":
		m = sslcert.FromResult(r)
	default:
		log.Errorf("Type %s is not yet supported\n", r.Type())
	}

	if m != nil {
		setAsnForMetricExporter(r, m)
	}

	out <- m
}
func setAsnForMetricExporter(r *measurement.Result, m metric.MetricExporter) {
	p, found := cache.Get(r.PrbId())

	if !found {
		var err error
		p, err = probe.Get(r.PrbId())

		if err != nil {
			log.Errorf("Could not get information for probe %d: %v\n", r.PrbId(), err)
			return
		}

		cache.Add(r.PrbId(), p)
	}

	if r.Af() == 4 {
		m.SetAsn(p.Asn4)
	} else {
		m.SetAsn(p.Asn6)
	}
}

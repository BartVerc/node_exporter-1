// Copyright 2016 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"flag"
	"io/ioutil"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
)

var (
	summaryfile = flag.String(
		"collector.puppet.summaryfile",
		"/var/lib/puppet/state/last_run_summary.yaml",
		"Path where puppet stores its summary file.")
)

type puppetCollector struct {
	configVersion     *prometheus.Desc
	resources         *prometheus.Desc
	timeOfLastRun     *prometheus.Desc
	lastRunStagesTime *prometheus.Desc
	events            *prometheus.Desc
}

type summaryConfig struct {
	Config float64 `yaml:"config"`
}

type summary struct {
	Resources map[string]float64 `yaml:"resources"`
	Time      map[string]float64 `yaml:"time"`
	Events    map[string]float64 `yaml:"events"`
	Version   summaryConfig      `yaml:"version"`
}

func init() {
	Factories["puppet"] = NewPuppetCollector
}

func NewPuppetCollector() (Collector, error) {
	return &puppetCollector{
		configVersion: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "puppet_catalog_version"),
			"Catalog version of Puppet.", nil, nil),
		resources: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "puppet_resources"),
			"Summary of Puppet resources", []string{"status"}, nil),
		lastRunStagesTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "puppet_time_per_stage_seconds"),
			"The time per stage of a Puppet run.", []string{"stage"}, nil),
		timeOfLastRun: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "puppet_time_of_last_run_seconds"),
			"Timestamp of the last puppet run in Unixtime", nil, nil),
		events: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "puppet_events"),
			"Summary of Puppet events.", []string{"status"}, nil),
	}, nil
}

func (c *puppetCollector) Update(ch chan<- prometheus.Metric) error {
	yamlFile, err := ioutil.ReadFile(*summaryfile)
	if err != nil {
		return err
	}

	var summary summary
	err = yaml.Unmarshal(yamlFile, &summary)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(c.configVersion, prometheus.GaugeValue,
		summary.Version.Config)

	for status, number := range summary.Resources {
		ch <- prometheus.MustNewConstMetric(c.resources, prometheus.GaugeValue,
			number, status)
	}

	for stage, time := range summary.Time {
		if stage == "last_run" {
			ch <- prometheus.MustNewConstMetric(c.timeOfLastRun,
				prometheus.GaugeValue, time)
		} else if stage != "total" {
			ch <- prometheus.MustNewConstMetric(c.lastRunStagesTime,
				prometheus.GaugeValue, time, stage)
		}
	}

	for status, number := range summary.Events {
		if status != "total" {
			ch <- prometheus.MustNewConstMetric(c.events, prometheus.GaugeValue,
				number, status)
		}
	}
	return nil
}

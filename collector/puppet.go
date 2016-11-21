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
	fail_count          *prometheus.Desc
	time_since_last_run *prometheus.Desc
}

type summaryResources struct {
	Failed int `yaml:"failed"`
}

type summaryTime struct {
	LastRun int `yaml:"last_run"`
}

type summaryEvents struct {
	Failure int `yaml:"failure"`
}

type summary struct {
	Resources summaryResources `yaml:"resources"`
	Time      summaryTime      `yaml:"time"`
	Events    summaryEvents    `yaml:"events"`
}

func init() {
	Factories["puppet"] = NewPuppetCollector
}

func NewPuppetCollector() (Collector, error) {
	return &puppetCollector{
		fail_count: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "Puppet_fail_count"),
			"The total Puppet fail_count.", nil, nil),
		time_since_last_run: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "Puppet_time_since_last_run"),
			"The time since the last Puppet run.", nil, nil),
	}, nil
}

func (c *puppetCollector) Update(ch chan<- prometheus.Metric) error {

	if err := c.collectFailCount(ch); err != nil {
		return err
	}
	return nil
}

func (c *puppetCollector) collectFailCount(ch chan<- prometheus.Metric) error {
	yamlFile, err := ioutil.ReadFile(*summaryfile)

	if err != nil {
		return err
	}

	var summary summary

	err = yaml.Unmarshal(yamlFile, &summary)
	if err != nil {
		return err
	}

	fail_count := float64(summary.Events.Failure + summary.Resources.Failed)
	time_since_last_run := float64(summary.Time.LastRun)

	ch <- prometheus.MustNewConstMetric(c.fail_count, prometheus.GaugeValue, fail_count)
	ch <- prometheus.MustNewConstMetric(c.time_since_last_run, prometheus.GaugeValue, time_since_last_run)
	return nil
}

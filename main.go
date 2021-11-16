package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	sonargo "github.com/magicsong/sonargo/sonar"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "sonarqube"

var (
	// Metrics
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last sonar query successful.",
		nil, nil,
	)

	healthStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "health_status"),
		"SonarQube Health Status",
		nil, nil,
	)
)

type Exporter struct {
	sonarEndpoint, sonarUsername, sonarPassword string
}

func NewExporter(sonarEndpoint string, sonarUsername string, sonarPassword string) *Exporter {
	return &Exporter{
		sonarEndpoint: sonarEndpoint,
		sonarUsername: sonarUsername,
		sonarPassword: sonarPassword,
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- healthStatus
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	health, err := e.GatherSonarMetrics()

	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0,
		)
		log.Println(err)
		return
	}
	ch <- prometheus.MustNewConstMetric(
		up, prometheus.GaugeValue, 1,
	)

	ch <- prometheus.MustNewConstMetric(
		healthStatus, prometheus.GaugeValue, health,
	)
}

func (e *Exporter) GatherSonarMetrics() (float64, error) {
	client, err := sonargo.NewClient(e.sonarEndpoint+"/api", e.sonarUsername, e.sonarPassword)
	if err != nil {
		return 0.0, err
	}

	v, _, err := client.System.Health()
	if err != nil {
		return 0.0, err
	}

	fmt.Println(v.Health)

	if v.Health == "GREEN" {
		return 1.0, nil
	}

	return 0.0, nil
}

func main() {
	sonarEndpoint := os.Getenv("SONAR_URL")
	sonarUsername := os.Getenv("SONAR_USER")
	sonarPassword := os.Getenv("SONAR_PASSWORD")

	exporter := NewExporter(sonarEndpoint, sonarUsername, sonarPassword)
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}

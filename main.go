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

	activityStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "activity_status"),
		"SonarQube Activity Status",
		[]string{"metric"}, nil,
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
	ch <- activityStatus
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	health, err := e.GatherSonarHealth()

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

	as, err := e.GatherSonarActivityStatus()

	if err != nil {
		log.Println(err)
	} else {
		ch <- prometheus.MustNewConstMetric(
			activityStatus, prometheus.GaugeValue, float64(as.Pending), "pending",
		)
		ch <- prometheus.MustNewConstMetric(
			activityStatus, prometheus.GaugeValue, float64(as.Failing), "failing",
		)
		ch <- prometheus.MustNewConstMetric(
			activityStatus, prometheus.GaugeValue, float64(as.InProgress), "inProgress",
		)
	}
}

func (e *Exporter) GatherSonarHealth() (float64, error) {
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

func (e *Exporter) GatherSonarActivityStatus() (*ActivityStatus, error) {
	client, err := NewClient(e.sonarEndpoint, e.sonarUsername, e.sonarPassword)
	if err != nil {
		return nil, err
	}

	return client.ActivityStatus()
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

package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/junkpiano/sonarqube-exporter/internal"

	"github.com/junkpiano/gosq"

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

	generalStats = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "general_stats"),
		"SonarQube General Statistics",
		[]string{"metric"}, nil,
	)

	codeDemographics = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "code_demographics"),
		"SonarQube Code Demographics",
		[]string{"lang"}, nil,
	)

	projectCountDemographics = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "project_count_demographics"),
		"SonarQube Project Count Demographics",
		[]string{"lang"}, nil,
	)

	searchStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "search_status"),
		"SonarQube Search State",
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
	ch <- generalStats
	ch <- codeDemographics
	ch <- projectCountDemographics
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	re, err := e.GatherSystemInfo()

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
		healthStatus, prometheus.GaugeValue, internal.ConvertStatusToFloat(re.Health),
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

	if err != nil {
		log.Println(err)
	} else {
		ch <- prometheus.MustNewConstMetric(
			generalStats, prometheus.GaugeValue, float64(re.Statistics.UserCount), "UserCount",
		)
		ch <- prometheus.MustNewConstMetric(
			generalStats, prometheus.GaugeValue, float64(re.Statistics.ProjectCount), "ProjectCount",
		)
		ch <- prometheus.MustNewConstMetric(
			generalStats, prometheus.GaugeValue, float64(re.Statistics.Ncloc), "NCLoC",
		)

		for _, n := range re.Statistics.NclocByLanguage {
			ch <- prometheus.MustNewConstMetric(
				codeDemographics, prometheus.GaugeValue, float64(n.Ncloc), n.Language,
			)
		}

		for _, n := range re.Statistics.ProjectCountByLanguage {
			ch <- prometheus.MustNewConstMetric(
				projectCountDemographics, prometheus.GaugeValue, float64(n.Count), n.Language,
			)
		}
	}

	ch <- prometheus.MustNewConstMetric(
		searchStatus, prometheus.GaugeValue, internal.ConvertStatusToFloat(re.Search_State.State), "Health",
	)

	ch <- prometheus.MustNewConstMetric(
		searchStatus, prometheus.GaugeValue, float64(re.Search_State.CPU_Usage), "CPU Usage",
	)

	sComps := strings.Split(re.Search_State.Disk_Available, " ")

	if len(sComps) > 0 {
		if val, err := strconv.ParseFloat(sComps[0], 64); err == nil {
			ch <- prometheus.MustNewConstMetric(
				searchStatus, prometheus.GaugeValue, float64(val), "Disk Available",
			)
		}
	}
}

func (e *Exporter) GatherSonarActivityStatus() (*gosq.ActivityStatus, error) {
	client, err := gosq.NewClient(e.sonarEndpoint, e.sonarUsername, e.sonarPassword)
	if err != nil {
		return nil, err
	}

	return client.ActivityStatus()
}

func (e *Exporter) GatherSystemInfo() (*gosq.SystemInfo, error) {
	client, err := gosq.NewClient(e.sonarEndpoint, e.sonarUsername, e.sonarPassword)
	if err != nil {
		return nil, err
	}

	return client.SystemInfo()
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

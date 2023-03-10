package main

import (
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/dkorunic/iSMC/output"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type SensorsCollector struct{}

func createNewDesc(catalog, description string, value interface{}) *prometheus.Desc {
	var unit = getUnit(value)

	var help = catalog + " " + description
	var variableLabels []string = []string{}
	var constLabels prometheus.Labels = prometheus.Labels{}

	re := regexp.MustCompile("[0-9]+")
	var idx = re.FindAllStringIndex(description, -1)
	if len(idx) == 1 && idx[0][0] > 0 {
		// log.Println(description, description[idx[0][0]:idx[0][1]], description[:idx[0][0]]+description[idx[0][1]:])
		help = description[:idx[0][0]] + description[idx[0][1]:]

		// variableLabels = append(variableLabels, "index")
		constLabels["index"] = description[idx[0][0]:idx[0][1]]
	}

	help = strings.TrimSpace(help)
	help = strings.Replace(help, "  ", " ", -1)

	var fqName = strings.ToLower(help)
	fqName = strings.Replace(fqName, " ", "_", -1)
	fqName = strings.Replace(fqName, ".", "_", -1)
	fqName = strings.Replace(fqName, "-", "_", -1)
	fqName = strings.Replace(fqName, "(", "", -1)
	fqName = strings.Replace(fqName, ")", "", -1)
	fqName = strings.Replace(fqName, "/", "_", -1)
	fqName = "sensor_" + fqName + unit

	log.Print(fqName, help, variableLabels, constLabels)

	return prometheus.NewDesc(
		fqName,
		help,
		variableLabels,
		constLabels)
}

func getUnit(value interface{}) string {
	if v, ok := value.(string); ok {
		if idx := strings.Index(v, " "); idx != -1 {

			switch v[idx:] {
			case " A":
				return "_amperes"
			case " V":
				return "_volts"
			case " W":
				return "_watt"
			case " °C":
				return "_celsius"
			case " rpm":
				return "_rpm"
			default:
				log.Print(value, v[idx:])
			}
		}
	}

	return ""
}

func getGaugeValue(value interface{}) float64 {

	switch v := value.(type) {
	case int:
		return float64(v)
	case float64:
		return v
	case bool:
		if v {
			return 1
		} else {
			return 0
		}
	case string:
		if idx := strings.Index(v, " "); idx != -1 {
			v = v[:idx]
		}

		if s, err := strconv.ParseFloat(v, 64); err == nil {
			return s
		}
	default:
		log.Printf("I don't know about type %T!\n", v)
	}

	return 0
}

// Describe implements prometheus.Collector.
func (l *SensorsCollector) Describe(ch chan<- *prometheus.Desc) {
}

// Collect implements prometheus.Collector.
func (l *SensorsCollector) Collect(ch chan<- prometheus.Metric) {

	for catalog, catalogValue := range output.GetAll() {
		// log.Printf("Catalog: %s\n", catalog)

		if catalogValue, ok := catalogValue.(map[string]interface{}); ok {

			for description, details := range catalogValue {
				// log.Printf("description: %s\n", description)
				// log.Printf("details: %s\n", details.(map[string]interface{})["value"])

				var value = details.(map[string]interface{})["value"]

				ch <- prometheus.MustNewConstMetric(createNewDesc(catalog, description, value),
					prometheus.GaugeValue,
					getGaugeValue(value),
				)
			}
		}
	}
}

func main() {
	collector := &SensorsCollector{}
	prometheus.MustRegister(collector)

	pattern := "/metrics"
	addr := "localhost:9101"
	log.Printf("staring server listening at %s%s", addr, pattern)

	http.Handle(pattern, promhttp.Handler())
	err := http.ListenAndServe(addr, nil)
	log.Fatal(err)
}

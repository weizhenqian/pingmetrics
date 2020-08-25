package basic

import (
	"strings"
	"strconv"
)

func IsContain(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

func NewMetricValue(metric string, val interface{}, tags ...string) *MetricValue {
	mv := MetricValue{
		Metric: metric,
		Value:  val,
	}

	size := len(tags)

	if size > 0 {
		mv.Tags = strings.Join(tags, ",")
	}

	return &mv
}

func GaugeValue(metric string, val interface{}, tags ...string) *MetricValue {
	val,_ =  strconv.ParseFloat(val,64)
	return NewMetricValue(metric, val, tags...)
}

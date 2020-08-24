package basic

type MetricValue struct {
	Endpoint  string      `json:"endpoint"`
	Metric    string      `json:"metric"`
	Value     interface{} `json:"value"`
	Step      int64       `json:"step"`
	Tags      string      `json:"tags"`
	Timestamp int64       `json:"timestamp"`
}
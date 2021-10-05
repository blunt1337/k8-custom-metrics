package main

import (
	"encoding/json"
	"net/http"
)

func (p *MetricsProvider) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		p.updateMetrics(resp, req)
		return
	case "GET":
		resp.WriteHeader(http.StatusOK)
		resp.Header().Set("Content-Type", "application/json")
		json.NewEncoder(resp).Encode(p.metrics)
		return
	}

	resp.WriteHeader(http.StatusNotFound)
}

func (p *MetricsProvider) updateMetrics(resp http.ResponseWriter, req *http.Request) {
	var data struct {
		Namespace string           `json:"namespace"`
		Name      string           `json:"name"`
		Metrics   map[string]int64 `json:"metrics"`
	}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(err.Error()))
		return
	}

	for metricName, value := range data.Metrics {
		// Get metric handler
		metric, ok := p.metrics[metricName]
		if !ok {
			metric = &Metric{}
			metric.Init()
			p.metrics[metricName] = metric
		}

		// Set value
		metric.SetValue(data.Namespace+"/"+data.Name, value)
	}

	resp.WriteHeader(http.StatusOK)
	return
}

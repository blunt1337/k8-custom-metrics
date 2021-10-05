package main

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
	"time"
)

type MetricValue struct {
	Value *resource.Quantity `json:"value"`
	Time  metav1.Time        `json:"time"`
}

type Metric struct {
	valuesLock sync.RWMutex
	values     map[string]*MetricValue
}

type Metrics map[string]*Metric

func NewMetrics() Metrics {
	return make(Metrics)
}

func (m *Metric) Init() {
	m.values = make(map[string]*MetricValue)
}

func (m *Metric) GetValue(name string) (*MetricValue, bool) {
	m.valuesLock.RLock()
	defer m.valuesLock.RUnlock()
	res, ok := m.values[name]
	return res, ok
}

func (m *Metric) SetValue(name string, value int64) {
	mvalue := &MetricValue{
		Value: resource.NewQuantity(value, resource.DecimalExponent),
		Time:  metav1.Time{time.Now()},
	}

	m.valuesLock.RLock()
	defer m.valuesLock.RUnlock()
	m.values[name] = mvalue
}

func (m *Metric) MarshalJSON() ([]byte, error) {
	m.valuesLock.RLock()
	defer m.valuesLock.RUnlock()
	return json.Marshal(&m.values)
}

package main

import (
	"context"
	"fmt"

	apierr "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/metrics/pkg/apis/custom_metrics"
	"k8s.io/metrics/pkg/apis/external_metrics"

	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider/helpers"
)

type MetricsProvider struct {
	client      dynamic.Interface
	mapper      apimeta.RESTMapper
	metrics     Metrics
	metricsList []provider.CustomMetricInfo
}

func NewProvider(client dynamic.Interface, mapper apimeta.RESTMapper) *MetricsProvider {
	p := &MetricsProvider{
		client:      client,
		mapper:      mapper,
		metrics:     NewMetrics(),
		metricsList: []provider.CustomMetricInfo{},
	}

	// Generate metric infos
	for name := range p.metrics {
		p.metricsList = append(p.metricsList, provider.CustomMetricInfo{
			GroupResource: schema.GroupResource{Group: "", Resource: "pods"},
			Metric:        name,
			Namespaced:    true,
		})
	}

	return p
}

// ListAllMetrics lists available metrics
func (p *MetricsProvider) ListAllMetrics() []provider.CustomMetricInfo {
	fmt.Printf("\nListAllMetrics\n")
	return p.metricsList
}

// Get a metric value by it's name
func (p *MetricsProvider) GetMetricByName(ctx context.Context, name types.NamespacedName, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValue, error) {
	fmt.Printf("\nGetMetricByName\n%#v\n%#v\n%#v\n%#v\n", ctx, name, info, metricSelector)

	value, err := p.valueFor(info, name)
	if err != nil {
		return nil, err
	}

	return p.metricFor(value, name, info, metricSelector)
}

func (p *MetricsProvider) GetMetricBySelector(ctx context.Context, namespace string, selector labels.Selector, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValueList, error) {
	fmt.Printf("\nGetMetricBySelector\n%#v\n%#v\n%#v\n%#v\n%#v\n", ctx, namespace, selector, info, metricSelector)
	return p.metricsFor(namespace, selector, info, metricSelector)
}

// valueFor is a helper function to get just the value of a specific metric
func (p *MetricsProvider) valueFor(info provider.CustomMetricInfo, name types.NamespacedName) (*MetricValue, error) {
	info, _, err := info.Normalized(p.mapper)
	if err != nil {
		return nil, err
	}

	metric, found := p.metrics[info.Metric]
	if !found {
		return nil, provider.NewMetricNotFoundForError(info.GroupResource, info.Metric, name.Name)
	}

	value, found := metric.GetValue(name.String())
	if !found {
		return nil, provider.NewMetricNotFoundForError(info.GroupResource, info.Metric, name.Name)
	}

	return value, nil
}

// metricFor is a helper function which formats a value, metric, and object info into a MetricValue which can be returned by the metrics API
func (p *MetricsProvider) metricFor(value *MetricValue, name types.NamespacedName, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValue, error) {
	objRef, err := helpers.ReferenceFor(p.mapper, name, info)
	if err != nil {
		return nil, err
	}

	metric := &custom_metrics.MetricValue{
		DescribedObject: objRef,
		Metric: custom_metrics.MetricIdentifier{
			Name: info.Metric,
		},
		Timestamp: value.Time,
		Value:     *value.Value,
	}

	metricSelStr := metricSelector.String()
	if len(metricSelStr) > 0 {
		sel, err := metav1.ParseToLabelSelector(metricSelStr)
		if err != nil {
			return nil, err
		}
		metric.Metric.Selector = sel
	}

	return metric, nil
}

// metricsFor is a wrapper used by GetMetricBySelector to format several metrics which match a resource selector
func (p *MetricsProvider) metricsFor(namespace string, selector labels.Selector, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValueList, error) {
	names, err := helpers.ListObjectNames(p.mapper, p.client, namespace, selector, info)
	if err != nil {
		return nil, err
	}

	res := make([]custom_metrics.MetricValue, 0, len(names))
	for _, name := range names {
		namespacedName := types.NamespacedName{Name: name, Namespace: namespace}
		value, err := p.valueFor(info, namespacedName)
		if err != nil {
			if apierr.IsNotFound(err) {
				continue
			}
			return nil, err
		}

		metric, err := p.metricFor(value, namespacedName, info, metricSelector)
		if err != nil {
			return nil, err
		}
		res = append(res, *metric)
	}

	return &custom_metrics.MetricValueList{
		Items: res,
	}, nil
}

// Unimplemented functions

func (p *MetricsProvider) GetExternalMetric(ctx context.Context, namespace string, metricSelector labels.Selector, info provider.ExternalMetricInfo) (*external_metrics.ExternalMetricValueList, error) {
	return nil, nil
}

func (p *MetricsProvider) ListAllExternalMetrics() []provider.ExternalMetricInfo {
	return []provider.ExternalMetricInfo{}
}

package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
	basecmd "sigs.k8s.io/custom-metrics-apiserver/pkg/cmd"
)

type CustomAdapter struct {
	basecmd.AdapterBase
}

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	cmd := &CustomAdapter{}
	cmd.Flags().AddGoFlagSet(flag.CommandLine) // make sure we get the klog flags
	cmd.Flags().Parse(os.Args)

	// Authorize /custom-metrics url
	cmd.CustomMetricsAdapterServerOptions.Authorization.AlwaysAllowPaths = append(cmd.CustomMetricsAdapterServerOptions.Authorization.AlwaysAllowPaths, "/custom-metrics")

	provider := cmd.makeProviderOrDie()
	cmd.WithCustomMetrics(provider)

	server, err := cmd.Server()
	if err != nil {
		klog.Fatalf("unable to run custom metrics adapter: %v", err)
	}

	// Our custom API
	server.GenericAPIServer.Handler.NonGoRestfulMux.Handle("/custom-metrics", provider)

	klog.Infof("Starting adapter...")

	// Run kubernetes metric client
	if err := cmd.Run(wait.NeverStop); err != nil {
		klog.Fatalf("unable to run custom metrics adapter: %v", err)
	}
}

func (a *CustomAdapter) makeProviderOrDie() *MetricsProvider {
	client, err := a.DynamicClient()
	if err != nil {
		klog.Fatalf("unable to construct dynamic client: %v", err)
	}

	mapper, err := a.RESTMapper()
	if err != nil {
		klog.Fatalf("unable to construct discovery REST mapper: %v", err)
	}

	return NewProvider(client, mapper)
}

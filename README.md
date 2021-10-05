# k8-custrom-metrics

Simple custom metrics api server for kubernetes. Just store metrics that can be fetched by kubernetes' metrics server.

Start with `kubectl apply -f k8-config.yaml`.

Push some metrics with your pods:
```
curl --request POST --insecure 'https://custom-metrics-apiserver.kube-system.svc.cluster.local/custom-metrics' \
--header 'Content-Type: application/json' \
--data-raw '{
	"namespace": "<NAMESPACE>",
	"name": "<POD_NAME>",
	"metrics": {
        "http-sebsocket-count": 1337,
        "http-request-count": 1338
    }
}'
```

Debug metrics with: `kubectl get -n metrics --raw /apis/custom.metrics.k8s.io/v1beta1/namespaces/<NAMESPACE>/pod/<POD_NAME>/<METRIC_NAME>`
Or with `curl --insecure 'https://custom-metrics-apiserver.kube-system.svc.cluster.local/custom-metrics'`
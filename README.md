# Kubectl Resource Plugin

A plugin to access Kubernetes resource requests, limits, and usage.

## Install

`go get -u github.com/howardjohn/kubectl-resources`

## Usage

```
kubectl resources --help
Plugin to access Kubernetes resource requests, limits, and usage.

Usage:
  kubectl-resources [flags]

Flags:
  -h, --help               help for kubectl-resources
  -n, --namespace string   namespace to query. If not set, all namespaces are included
  -c, --show-containers    include container level details
  -d, --show-nodes         include node names
  -v, --verbose            show full resource names
```

```
kubectl resources
NAMESPACE       POD             CPU USE CPU REQ CPU LIM MEM USE MEM REQ MEM LIM
default         details-v1      6m      110m    2000m   36Mi    39Mi    1000Mi
default         productpage-v1  12m     110m    2000m   71Mi    39Mi    1000Mi
default         ratings-v1      5m      110m    2000m   34Mi    39Mi    1000Mi
default         reviews-v1      6m      110m    2000m   117Mi   39Mi    1000Mi
default         reviews-v2      7m      110m    2000m   106Mi   39Mi    1000Mi
default         reviews-v3      6m      110m    2000m   114Mi   39Mi    1000Mi
default         shell-f2g7f     4m      20m     2500m   24Mi    164Mi   1465Mi
default         shell-p6ggs     5m      20m     2500m   24Mi    164Mi   1465Mi
```

`kubectl resources` will by default shorten pod and node names. This can be disabled with `-v`.
In the above example, the `details-v1` deployment has just one replica, so only the deployment name is shown.
For `shell`, which has multiple replicas, just the replicaset's identifier is shown.
For nodes, only the unique segment of the name is shown.

To show more details, you can add the `-c` flag to show containers, and the `-d` flag to show nodes.

```
kubectl resources -v -c -d
NAMESPACE  POD                             CONTAINER       NODE                              CPU USE CPU REQ CPU LIM MEM USE  MEM REQ MEM LIM
default    details-v1-5cb65fd66c-6mtt2     details         gke-default-pool-f93f0b08-zwt6    1m      100m    -       10Mi     -       -
default    details-v1-5cb65fd66c-6mtt2     istio-proxy     gke-default-pool-f93f0b08-zwt6    5m      10m     2000m   25Mi     39Mi    1000Mi
```

# Kubectl Resource Plugin

[![Build status](https://badge.buildkite.com/cf86c7994aac947617af9b5a26cd4377f75f62d4f5a0529efa.svg)](https://buildkite.com/john-howard/build)

A plugin to access Kubernetes resource requests, limits, and usage.

## Install

You can download and install kubectl-resources from [release](https://github.com/howardjohn/kubectl-resources/releases/latest).

Also You can install kubectl-resources by `go install`:
``` shell
go install github.com/howardjohn/kubectl-resources
```

## Usage

```
Plugin to access Kubernetes resource requests, limits, and usage.

Usage:
  kubectl-resources [flags]

Flags:
  -A, --all-namespaces                 If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.
  -b, --by string                      column to aggregate on. Default is pod (default "POD")
  -c, --color                          show colors for pods using excessive resources (default true)
  -h, --help                           help for kubectl-resources
  -n, --namespace string               If present, the namespace scope for this CLI request
  -l, --selector string                Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
  -d, --show-nodes                     include node names
  -v, --verbose                        show full resource names
  -w, --warnings                       only show resources using excessive resources
```

Example output:

```
$ kubectl resources
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

You can also aggregate info at different levels. By default this is at the pod level, but can also be namespace, node, or container.

For example, to see resource utilization by namespace

```
$ kubectl resources --by namespace
NAMESPACE     CPU USE  CPU REQ  CPU LIM  MEM USE  MEM REQ  MEM LIM
default       3m       110m     2100m    29Mi     144Mi    1152Mi
istio-system  70m      2140m    14800m   570Mi    3641Mi   8934Mi
kube-system   82m      1831m    3222m    1312Mi   1203Mi   2365Mi
              155m     4081m    20122m   1912Mi   4989Mi   12452Mi
```

## Known Issues

While init containers [play a role in resource allocation](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/#resources), they are not accounted for in this tool.

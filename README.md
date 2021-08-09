# Morphling

Morphling is an auto-configuration framework for
machine learning model serving (inference) on Kubernetes.  Check the [website](http://kubedl.io/tuning/intro/) for details.

## Overview

Morphling tunes the optimal configurations for your ML/DL model serving deployments.
It searches the best container-level configurations (e.g., resource allocations and runtime parameters) by empirical trials, where a few configurations are sampled for performance evaluation. 

![Stack](docs/img/stack.png)

## Features
Key benefits include:

- Automated tuning workflows hidden behind simple APIs.
- Out of the box ML model serving stress-test clients.
- Cloud agnostic and tested on [AWS](https://aws.amazon.com/), [Alicloud](https://us.alibabacloud.com/), etc. 
- ML framework agnostic and generally support popular frameworks, including [TensorFlow](https://github.com/tensorflow/tensorflow), [PyTorch](https://github.com/pytorch/pytorch), etc. 
- Equipped with various and customizable hyper-parameter tuning algorithms.  

## Getting started

#### Install CRDs

From git root directory, run

```commandline
kubectl apply -f config/crd/bases
```


#### Install Morphling Components
     
 ```commandline
 kubectl create namespace morphling-system
 
 kubectl apply -f manifests/configmap
 kubectl apply -f manifests/controllers
 kubectl apply -f manifests/pv
 kubectl apply -f manifests/mysql-db
 kubectl apply -f manifests/db-manager
 ```
By default, Morphling will be installed under `morphling-system` namespace.

The official Morphling component images are hosted under [docker hub](https://hub.docker.com/r/kubedl).

Check if all components are running successfully:
```commandline
kubectl get deployment -n morphling-system
```

Expected output:
```commandline
NAME                   READY   UP-TO-DATE   AVAILABLE   AGE
morphling-controller   1/1     1            1           10m
morphling-db-manager   1/1     1            1           10m
morphling-mysql        1/1     1            1           10m
```

#### Uninstall Morphling controller

```bash
kubectl delete namespace morphling-system
```

#### Delete CRDs
```bash
kubectl delete crd profilingexperiments.tuning.kubedl.io samplings.tuning.kubedl.io trials.tuning.kubedl.io
```

## Running Examples

This example demonstrates how to tune the configuration for a [resnet50](https://www.tensorflow.org/api_docs/python/tf/keras/applications/ResNet50) model deployed with [Tensorflow Serving](https://www.tensorflow.org/tfx/guide/serving) under Morphling.

For demonstration, we choose _two_ configurations to tune: 
the first one the CPU cores (resource allocation), and the second one is maximum serving batch size (runtime parameter). 
We use grid search for configuration sampling.

#### Submit the configuration tuning experiment

```bash
kubectl -n morphling-system apply -f https://raw.githubusercontent.com/alibaba/morphling/master/example/experiment/experiment-resnet50-grid.yaml
```

#### Monitor the status of the configuration tuning experiment
```bash
kubectl get -n morphling-system pe
kubectl describe -n morphling-system pe
```
#### Monitor sampling trials (performance test)
```bash
kubectl -n morphling-system get trial
```

#### Get the searched optimal configuration
```bash
kubectl -n morphling-system get pe
```

Expected output:
```bash
NAME                  STATE       AGE   OBJECT NAME   OPTIMAL OBJECT VALUE   OPTIMAL PARAMETERS
resnet50-experiment   Succeeded   12m   qps           15                     [map[category:resource name:cpu value:4] map[category:env name:BATCH_SIZE value:32]]
```

#### Delete the tuning experiment

```bash
kubectl -n morphling-system delete pe --all
```

#### Other Examples 
Check the [Quick Start](docs/quick_start.md) for more examples.

##  Workflow
See [Morphling Workflow](./docs/workflow-design.md) to check how Morphling tunes ML serving 
configurations automatically in a Kubernetes-native way.

## Developer Guide

#### Build the controller manager binary

```bash
make manager
```
#### Run the tests

```bash
make test
```
#### Generate manifests, e.g., CRD, RBAC YAML files, etc.

```bash
make manifests
```
#### Build the component docker images, e.g., Morphling controller, DB-Manager

```bash
make docker-build
```

#### Push the component docker images

```bash
make docker-push
```

To develop/debug Morphling controller manager locally, please check the [debug guide](./docs/debug_guide.md).

## Community

If you have any questions or want to contribute, GitHub issues or pull requests are warmly welcome.

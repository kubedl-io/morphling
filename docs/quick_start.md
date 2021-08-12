# Quick Start


This example demonstrates how to tune the configuration for a [mobilenet](https://www.tensorflow.org/api_docs/python/tf/keras/applications/mobilenet) model deployed with [Tensorflow Serving](https://www.tensorflow.org/tfx/guide/serving) under Morphling.

## Tuning a ProfilingExperiment using Random Search
In this example, apart from CPU cores and batch size shown in [README](../README.md), we also include
GPU memory sharing as an additional tunable configuration.
Besides, we use random search for configuration sampling.

```commandline
kubectl -n morphling-system apply -f https://raw.githubusercontent.com/alibaba/morphling/main/examples/experiment/experiment-mobilenet-grid.yaml
```

```yaml
# kubectl  -n morphling-system get pe mobilenet-experiment-grid -o yaml
apiVersion: "tuning.kubedl.io/v1alpha1"
kind: ProfilingExperiment
metadata:
  namespace: morphling-system
  name: mobilenet-experiment-grid
spec:
  objective:
    type: maximize
    objectiveMetricName: qps
  algorithm:
    algorithmName: grid
  parallelism: 1
  maxNumTrials: 18
  clientTemplate:
    spec:
      template:
        spec:
          containers:
            - name: pi
              image: kubedl/morphling-http-client:demo
              env:
                - name: TF_CPP_MIN_LOG_LEVEL
                  value: "3"
                - name: MODEL_NAME
                  value: "mobilenet"
              resources:
                requests:
                  cpu: 800m
                  memory: "1800Mi"
                limits:
                  cpu: 800m
                  memory: "1800Mi"
              command: [ "python3" ]
              args: ["morphling_client.py", "--model", "mobilenet", "--printLog", "True", "--num_tests", "10"]

              imagePullPolicy: IfNotPresent
          restartPolicy: Never
      backoffLimit: 1

  servicePodTemplate:
    template:
      spec:
        containers:
          - name: service-container
            image: kubedl/morphling-tf-model:demo  #-cv
            imagePullPolicy: IfNotPresent
            env:
              - name: MODEL_NAME
                value: "mobilenet"
            resources:
              requests:
                cpu: 1
                memory: "1800Mi"
                # nvidia.com/gpu: "1"
              limits:
                cpu: 1
                memory: "1800Mi"
                # nvidia.com/gpu: "1"
            ports:
              - containerPort: 8500

  tunableParameters:
    - category: "resource"
      parameters:
        - parameterType: discrete
          name: "cpu"
          feasibleSpace:
            list:
              #              - "500m"
              #              - "1500m"
              - "2000m"
        - parameterType: discrete
          name: "memory"
          feasibleSpace:
            list:
              - "200Mi"
              #              - "1000Mi"
              - "2000Mi"
    - category: "env"
      parameters:
        - parameterType: discrete
          name: "BATCH_SIZE"
          feasibleSpace:
            list:
              - "1"
              - "2"
```
### List current trials

```commandline
kubectl -n morphling-system get trial
```

#### Get the searched optimal configuration
```bash
kubectl -n morphling-system get pe
```

### Delete ProfilingExperiment

```commandline
kubectl delete -n morphling-system pe --all
```

## Performance Evaluation for a Single Trial 
Morphling also support launching a single trial to test the configuration performance.

```commandline
kubectl -n morphling-system apply -f example/trial/experiment-resnet50.yaml
```

### List current trials

```commandline
kubectl -n morphling-system get trial
```

### Delete the trial

```commandline
kubectl delete -n morphling-system trial --all
```
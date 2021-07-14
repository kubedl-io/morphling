# Quick Start


This example demonstrates how to tune the configuration for a [resnet50](https://www.tensorflow.org/api_docs/python/tf/keras/applications/ResNet50) model deployed with [Tensorflow Serving](https://www.tensorflow.org/tfx/guide/serving) under Morphling.

## Tuning a ProfilingExperiment using Random Search
In this example, apart from CPU cores and batch size shown in [README](../README.md), we also include
GPU memory sharing as an additional tunable configuration.
Besides, we use random search for configuration sampling.

```commandline
kubectl -n morphling-system apply -f https://raw.githubusercontent.com/alibaba/morphling/master/experiment/experiment-resnet50-random.yaml
```

```yaml
# kubectl  -n morphling-system get pe resnet50-experiment-random -o yaml
apiVersion: "tuning.kubedl.io/v1alpha1"
kind: ProfilingExperiment
metadata:
  namespace: morphling-system
  name: resnet50-experiment-random
spec:
  requestTemplate: "https://ss1.bdstatic.com/70cFuXSh_Q1YnxGkpoWK1HF6hhy/it/u=2153705155,1396952620&fm=11&gp=0.jpg"
  objective:
    type: maximize
    objectiveMetricName: qps
  algorithm:
    algorithmName: random
  parallelism: 1
  maxNumTrials: 3
  clientTemplate:
    metadata:
      name: "resnet50-client"
      namespace: "default"
    spec:
      template:
        spec:
          imagePullSecrets:
            - name: harborsecretkey
          containers:
            - name: pi
              image: kubedl/morphling-http-client
              env:
                - name: TF_CPP_MIN_LOG_LEVEL
                  value: "3"
                - name: MODEL_NAME
                  value: "resnet50"
              resources:
                requests:
                  cpu: 4
                limits:
                  cpu: 4
              command: [ "python3" ]
              args: ["delphin_client.py", "--model", "resnet50", "--printLog", "True"]

              imagePullPolicy: IfNotPresent
          restartPolicy: Never
      backoffLimit: 4

  servicePodTemplate:
    metadata:
      name: "resnet-pod"
      namespace: "default"
    template:
      spec:
        imagePullSecrets:
          - name: harborsecretkey
        containers:
          - name: resnet-container
            image: kubedl/morphling-tf-model
            imagePullPolicy: IfNotPresent
            env:
              - name: MODEL_NAME
                value: "resnet50"
            resources:
              requests:
                cpu: 1
                nvidia.com/gpu: "1"
              limits:
                cpu: 1
                nvidia.com/gpu: "1"
            ports:
              - containerPort: 8500

  tunableParameters:
    - category: "resource"
      parameters:
        - parameterType: int
          name: "cpu"
          feasibleSpace:
            min: "1"
            max: "5"
            step: "1"
    - category: "env"
      parameters:
        - parameterType: double
          name: "GPU_MEM"
          feasibleSpace:
            min: "0.20"
            max: "1.01"
            step: "0.20"
        - parameterType: discrete
          name: "BATCH_SIZE"
          feasibleSpace:
            list:
              - "1"
              - "2"
              - "4"
              - "8"
              - "16"
              - "32"
              - "64"
              - "128"
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
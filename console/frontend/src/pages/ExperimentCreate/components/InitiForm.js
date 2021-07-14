const initialParameter = {
  name: "demo-experiment",
  namespace: "morphling-system",
  algorithmName: "random",
  objectiveType: "maximize",
  objectiveName: "qps",
  maxTrials: 3,
  parallelism: 1,
  tuningParameters: [{
    key: '1',
    category: 'Resource',
    name: 'CPU',
    type: 'double',
    max: '2',
    min: '1',
    step: '0.1',
    list: []
  },
    {
      key: '2',
      category: 'Env',
      name: 'BATCH_SIZE',
      type: 'int',
      max: '2',
      min: '1',
      step: '1',
      list: []
    }]


};

const initialYaml = ''

const initialClientYaml = `metadata:
  name: "resnet50-client"
  namespace: "default"
spec:
  template:
    spec:
      containers:
      - name: pi
        image: kubedl/morphling-http-client
        env:
          - name: TF_CPP_MIN_LOG_LEVEL
            value: "3"
          - name: MODEL_NAME
            value: "resnet50"
        command: [ "python3" ]
        args: ["morphling_client.py", "--model", "resnet50", "--printLog", "True"]

        imagePullPolicy: IfNotPresent
      restartPolicy: Never
  backoffLimit: 4`

const initialServiceYaml = 'metadata:\n' +
  '  name: "resnet-pod"\n' +
  '  namespace: "default"\n' +
  'template:\n' +
  '  spec:\n' +
  '    containers:\n' +
  '      - name: resnet-container\n' +
  '        image: registry.cn-hangzhou.aliyuncs.com/delphin/resnet-model:aws #kubedl/morphling-tf-model:resnet50 #gcr.io/tensorflow-serving/resnet\n' +
  '        imagePullPolicy: IfNotPresent\n' +
  '        env:\n' +
  '          - name: MODEL_NAME\n' +
  '            value: "resnet50"\n' +
  '        resources:\n' +
  '          requests:\n' +
  '            cpu: 1\n' +
  '            # nvidia.com/gpu: "1"\n' +
  '          limits:\n' +
  '            cpu: 1\n' +
  '            # nvidia.com/gpu: "1"\n' +
  '        ports:\n' +
  '          - containerPort: 8500'
export {
  initialClientYaml,
  initialServiceYaml,
  initialYaml,
  initialParameter
}

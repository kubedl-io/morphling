function getFakeCaptcha(req, res) {
  return res.json('captcha-xxx');
} // 代码中会兼容本地 service mock 以及部署站点的静态数据

export default {
  // 支持值为 Object 和 Array
  'GET /api/v1alpha1/data/total': (req, res) => {
    res.send({
      code: '200',
      data: {
        totalCPU: 604000,
        totalMemory: 2747247054848,
        totalGPU: 16000
      },
    });
  },

  'POST /api/v1alpha1/experiment/submitYaml': (req, res) => {
    res.send({
      code: '200',
      data: {"ee": "ee"},
    });
  },
  'POST /api/v1alpha1/experiment/submitPars': (req, res) => {
    res.send({
      code: '200',
      data: {"ee": "ee"},
    });
  },


  'GET /api/v1alpha1/data/request/Running': (req, res) => {
    res.send({
      code: '200',
      data: {
        requestCPU: 22400,
        requestMemory: 40170946560,
        requestGPU: 2000
      },
    });
  },


  'GET /api/v1alpha1/data/nodeInfos': (req, res) => {
    res.send({
      code: '200',
      data: {
        items: [
          {
            nodeName: 'cn-hangzhou.192.174.35',
            instanceType: 'ecs.gn5-c8g1.2xlarge',
            totalCPU: 96000,
            totalMemory: 776498278400,
            totalGPU: 8000,
            requestCPU: 400,
            requestMemory: 1493172224,
            requestGPU: 0,
            gpuType: 'T4',

          },
          {
            nodeName: 'cn-beijing.192.174.35',
            instanceType: 'ecs.gn5-c8g1.2xlarge',
            totalCPU: 96000,
            totalMemory: 776498278400,
            totalGPU: 8000,
            requestCPU: 400,
            requestMemory: 1493172224,
            requestGPU: 0,
            gpuType: 'V100',
          },
          {
            nodeName: 'cn-shanghai.192.174.35',
            instanceType: 'ecs.gn5-c8g1.2xlarge',
            totalCPU: 96000,
            totalMemory: 776498278400,
            totalGPU: 8000,
            requestCPU: 400,
            requestMemory: 1493172224,
            requestGPU: 0,
            gpuType: '2080',
          },
        ],
      },
    });
  },

  'GET /api/v1alpha1/data/config': (req, res) => {
    res.send({
      code: '200',
      data: {
        namespace: "morphling-system",
        "http-client-image": "kubedl/morphling-http-client",
        "http-client.yaml": "unknown",
        "pai-tf-cpu-images": "unknown",
        "pai-tf-gpu-images": "unknown",
      },
    })
  },

  'GET /api/v1alpha1/data/namespaces': (req, res) => {
    res.send({
      code: '200',
      data: [ "default", "kube-system", "morphling-system"],
    })
  },

  'GET /api/v1alpha1/data/algorithmNames': (req, res) => {
    res.send({
      code: '200',
      data: ["random", "Bayesian Opt", "grid"],
    })
  },


  'GET /api/v1alpha1/experiment/list': (req, res) => {
    res.send({
      code: '200',
      data: {
        "peInfos": [
          {
            name: 'resnet50-experiment-random',
            UserName: 'user-1',
            UserId: 123456789,
            peStatus: 'Created',
            namespace: 'default',
            createTime: "2021-01-02 15:04:05",
            endTime: "2021-01-02 15:08:05",
            durationTime: "4m0s",
          },

        ],
        "total": 1,
      },
    });
  },

  'GET /api/v1alpha1/experiment/detail': (req, res) => {
    res.send({
      code: '200',
      data: {
        "peInfo":
          {
            name: 'resnet50-experiment-randomde',
            UserName: 'user-1',
            UserId: 123456789,
            peStatus: 'Created',
            namespace: 'default',
            createTime: "2021-01-02 15:04:05",
            endTime: "2021-01-02 15:08:05",
            durationTime: "4m0s",
            trialsTotal: 3,
            trialsSucceeded: 3,
            algorithmName: "random",
            maxNumTrials: 3,
            objective: "maximize qps",
            parallelism: 1,
            specs: [{name: "111", jobStatus: "eeee"}],
            parameters: [
              {category: "Resource", name: "CPU", type: "Int", space: "max: 5, min: 1, step: 1"},
              {category: "Env", name: "Batch Size", type: "Int", space: "[1,2, 4, 8, 16, 32]"},
            ],
            trials: [
              {
                name: "resnet50-experiment-random-dpg9qlps",
                Status: "Succeeded",
                objectiveName: "qps",
                objectiveValue: "3",
                createTime: "2021-01-02 15:04:05",
                parameterSamples: {CPU: 1, "Batch Size": 1},
                isOptimal: false,
              },
              {
                name: "resnet50-experiment-random-rer3frg4",
                Status: "Succeeded",
                objectiveName: "qps",
                objectiveValue: "4",
                createTime: "2021-01-02 15:04:05",
                parameterSamples: {CPU: 1, "Batch Size": 2},
                isOptimal: false,
              },
              {
                name: "resnet50-experiment-random-oemyu879",
                Status: "Succeeded",
                objectiveName: "qps",
                objectiveValue: "6",
                createTime: "2021-01-02 15:04:05",
                parameterSamples: {CPU: 1, "Batch Size": 3},
                isOptimal: true,
              },
            ],
            currentOptimalTrial: [{
              objectiveName: "qps",
              objectiveValue: "6",
              parameterSamples: {CPU: 1, "Batch Size": 3},
            }]
            // parameter_entries: ["CPU", "Batch Size"]
          },
        "total": 1,
      },
    });
  },

};

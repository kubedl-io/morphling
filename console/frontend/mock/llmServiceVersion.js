export default {
  'GET /api/v1alpha1/llm-service-version': (req, res) => {
    res.send({
      code: '200',
      data: [
        {
          id: 1,
          modelName: 'microsoft/phi-2',
          version: 'v1.0.0',
          status: 'COMPLETED',
          createdAt: '2024-03-20 10:00:00',
          testedAt: '2024-03-20 12:00:00',
          objectiveMetricName: 'qps',
          bestValue: 0.98
        },
        {
          id: 2,
          modelName: 'GPT2',
          version: 'v1.0.0',
          status: 'PENDING',
          createdAt: '2024-03-20 10:00:00',
          objectiveMetricName: 'qps',
        }
      ],
    });
  },
};
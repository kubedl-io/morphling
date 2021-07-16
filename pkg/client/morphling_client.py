from __future__ import print_function

import tensorflow as tf
import threading
from tensorflow_serving.apis import predict_pb2
from tensorflow_serving.apis import prediction_service_pb2_grpc
import os
import grpc
import api_pb2
import api_pb2_grpc
import rfc3339
import time
import math
import numpy as np
from queue import Queue


models = {
    'densenet121': [224,224],
    'densenet169': [224,224],
    'densenet201': [224,224],
    'efficientnetb0': [224,224],
    'efficientnetb1': [240,240],
    'efficientnetb2': [260,260],
    'efficientnetb3': [300,300],
    'efficientnetb4': [380,380],
    'efficientnetb5': [456,456],
    'efficientnetb6': [528,528],
    'efficientnetb7': [600,600],
    'inceptionresnetv2': [299,299],
    'inceptionv3': [299,299],
    'mobilenet': [224,224],
    'mobilenetv2': [224,224],
    'nasnetlarge': [331,331],
    'nasnetmobile': [224,224],
    'resnet101': [224,224],
    'resnet152': [224,224],
    'resnet50': [224,224],
    'resnet101v2': [224,224],
    'resnet152v2': [224,224],
    'resnet50v2': [224,224],
    'vgg16': [224,224],
    'vgg19': [224,224],
    'xception': [299,299]
}


with tf.device("/cpu:0"):
    tf.get_logger().setLevel('ERROR')

    # The image URL is the location of the image we should send to the server
    # IMAGE_URL = os.environ['RequestTemplate']

    tf.compat.v1.app.flags.DEFINE_integer('concurrency', 100000, 'maximum number of concurrent inference requests')
    tf.compat.v1.app.flags.DEFINE_integer('num_tests', 3, 'Number of test images per test')
    tf.compat.v1.app.flags.DEFINE_integer('batch_size', os.environ['BATCH_SIZE'], 'Number of test images per query')
    tf.compat.v1.app.flags.DEFINE_integer('qps', 10, 'QPS initial value')
    tf.compat.v1.app.flags.DEFINE_string('server', os.environ['ServiceName'], 'PredictionService host:port')
    tf.compat.v1.app.flags.DEFINE_string('image', '', 'path to imxage in JPEG format')
    tf.compat.v1.app.flags.DEFINE_string('model', os.environ['MODEL_NAME'], 'model name')
    tf.compat.v1.app.flags.DEFINE_string('signature', 'serving_default', 'signature name')
    tf.compat.v1.app.flags.DEFINE_string('inputs', 'inputs', 'signatureDef for inputs')
    tf.compat.v1.app.flags.DEFINE_string('outputs', 'predictions', 'signatureDef for outputs')
    tf.compat.v1.app.flags.DEFINE_enum('task', default='cv', enum_values=['cv', 'nlp'], help='which type of task')
    tf.compat.v1.app.flags.DEFINE_bool('printLog', False, 'whether to print temp results')
    FLAGS = tf.compat.v1.app.flags.FLAGS

    # dl_request = requests.get(IMAGE_URL, stream=True)
    # dl_request.raise_for_status()
    # data = dl_request.content
    if FLAGS.task == 'cv':
        with open("./image.jpg", 'rb') as f:
            data = f.read()
        data = tf.image.decode_jpeg(data)
        data = tf.image.convert_image_dtype(data, dtype=tf.float32)
        data = tf.image.resize(data, size=models[FLAGS.model])
        data = tf.expand_dims(data, axis=0)
    elif FLAGS.task == 'nlp':
        data = tf.convert_to_tensor(["This is a test!"])
    data = tf.concat([data] * FLAGS.batch_size, axis=0)
    if FLAGS.printLog:
        print("Input data shape: ", data.shape)
        print("FLAGS.batch_size: ", FLAGS.batch_size)
    timeout = 100  # 100 seconds
    manager_server = "morphling-db-manager:6799"
    channel_manager = grpc.insecure_channel(manager_server)
    timeout_in_seconds = 10
    channel = grpc.insecure_channel(FLAGS.server)
    stub = prediction_service_pb2_grpc.PredictionServiceStub(channel)


    def predict(test_mode=False):
        request = predict_pb2.PredictRequest()
        request.model_spec.name = FLAGS.model  # 'resnet50'
        request.model_spec.signature_name = FLAGS.signature
        request.inputs[FLAGS.inputs].CopyFrom(tf.make_tensor_proto(data, shape=list(data.shape)))
        result = stub.Predict(request, timeout)  # 100 seconds
        if test_mode:
            print(result)
        response = np.array(
            result.outputs[FLAGS.outputs].float_val)
        prediction = np.argmax(response)
        if test_mode:
            print('Prediction:', prediction)


    class Item:
        """An item that we queue for processing by the thread pool."""

        def __init__(self, request):
            self.request = request
            self.start = time.time()


    class QueueRunner:
        def __init__(self, threads):
            self.threads = threads
            self.tasks = Queue()
            self.workers = []
            self._error = 0
            self._done = 0
            self._rt = []
            self._condition = threading.Condition()

            for _ in range(self.threads):
                worker = threading.Thread(target=self.handle_tasks, args=(self.tasks,))
                worker.daemon = True
                self.workers.append(worker)
                worker.start()

        def run_one_item(self, item):
            try:
                result = stub.Predict(item.request, timeout)  # 100 seconds
                response = np.array(
                    result.outputs[FLAGS.outputs].float_val)
                _prediction = np.argmax(response)
            except Exception as e:
                if FLAGS.printLog:
                    print(e)
                with self._condition:
                    self._error += 1
            finally:
                response_time = time.time() - item.start
                with self._condition:
                    self._rt.append(response_time)
                    self._done += 1
                    self._condition.notify()

        def get_pass_items(self):
            return self._done - self._error

        def get_error_rate(self):
            return self._error / float(self._done)

        def get_avg_rt(self):
            return np.mean(self._rt)

        def get_tail_rt(self):
            pass

        def handle_tasks(self, tasks_queue):
            while True:
                item = tasks_queue.get()
                if item is None:
                    tasks_queue.task_done()
                    break
                self.run_one_item(item)
                tasks_queue.task_done()

        def enqueue(self, request):
            self.tasks.put(Item(request))

        def finish(self):
            for _ in self.workers:
                self.tasks.put(None)
            for worker in self.workers:
                worker.join()


    def do_inference(num_tests, qps):
        """Tests PredictionService with concurrent requests.

        Args:
            qps: Expected queries per second
            hostport: Host:port address of the PredictionService.
            concurrency: Maximum number of concurrent requests.
            num_tests: Number of test images to use.

        Returns:
            The QPS and classification error rate.
        """
        error_rate_list = []
        rt_list = []
        qps_list = []

        # prepare the request
        request = predict_pb2.PredictRequest()
        request.model_spec.name = FLAGS.model  # 'resnet50'
        request.model_spec.signature_name = FLAGS.signature
        request.inputs[FLAGS.inputs].CopyFrom(tf.make_tensor_proto(data, shape=list(data.shape)))

        for i in range(num_tests):
            runner = QueueRunner(4)

            start_time = time.time()
            time_wait = 1.0 / qps
            for j in range(max(math.ceil(qps), 4)):
                runner.enqueue(request=request)
                time.sleep(time_wait)

            runner.finish()
            total_time = time.time() - start_time
            error_rate_list.append(runner.get_error_rate())
            rt_list.append(runner.get_avg_rt())
            qps_list.append(runner.get_pass_items() / total_time)

        return np.mean(error_rate_list), np.mean(rt_list), np.mean(qps_list)


    def main(_):
        assert FLAGS.num_tests <= 10000
        assert FLAGS.server != ""

        # warmup
        warmup_samples = 10
        for _ in range(warmup_samples):
            predict()

        qps_max = None
        qps_current = FLAGS.qps
        qps_previous = 0
        rt_slo = 1.0

        error_rate, rt, qps_real = do_inference(num_tests=FLAGS.num_tests, qps=qps_current)

        if FLAGS.printLog:
            print('\nPreparing QPS: %s, QPS_real: %s, Inference error rate: %s%%, RT: %s' %
                  (qps_current, qps_real, error_rate * 100, np.mean(rt)))

        # service is not available
        if error_rate > 0.01:
            qps_previous = 0
        else:
            while True:
                if qps_current < 1:
                    break
                error_rate, rt, qps_real = do_inference(num_tests=FLAGS.num_tests, qps=qps_current)

                if FLAGS.printLog:
                    print('\nQPS: %s, QPS_real: %s, Inference error rate: %s%%, RT: %s' %
                          (qps_current, qps_real, error_rate * 100, np.mean(rt)))

                qps_current = qps_real
                if qps_previous > 0 and 1.1 > qps_current / qps_previous > 0.9:
                    break
                if error_rate > 0.1:
                    break
                if rt < rt_slo:
                    if qps_max:
                        qps_previous = qps_current
                        qps_current = (qps_max + qps_previous) / 2
                    else:
                        qps_previous = qps_current
                        qps_current *= 2
                elif rt >= rt_slo:
                    qps_max = qps_current
                    qps_current = (qps_max + qps_previous) / 2

                if FLAGS.printLog:
                    print('\n next QPS: %s' % qps_current)

        qps_previous = int(qps_previous * FLAGS.batch_size)
        print(qps_previous)

        mls = []
        ml = api_pb2.MetricLog(
            time_stamp=rfc3339.rfc3339(time.time()),
            metric=api_pb2.Metric(
                name='qps',
                value=str(qps_previous)
            )
        )
        mls.append(ml)

        observation_log = api_pb2.ObservationLog(metric_logs=mls)

        stub = api_pb2_grpc.ManagerStub(channel_manager)
        result = stub.ReportObservationLog(api_pb2.ReportObservationLogRequest(
            trial_name=os.environ['TrialName'],
            observation_log=observation_log
        ), timeout=timeout_in_seconds)
        if FLAGS.printLog:
            print(result)


    if __name__ == '__main__':
        tf.compat.v1.app.run()
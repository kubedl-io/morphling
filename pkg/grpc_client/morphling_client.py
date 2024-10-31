import os
import time
import grpc
import json
import math
import numpy as np
import threading
from queue import Queue

import predict_pb2
import predict_pb2_grpc
import api_pb2 as db_pb2
import api_pb2_grpc as db_pb2_grpc
import threading

# Set constants and parameters
BATCH_SIZE = int(os.getenv("BATCH_SIZE", 1))
CONCURRENCY = 4  # Number of concurrent request threads
NUM_TESTS = int(os.getenv("NUM_TESTS", 3))
INPUT_LENGTH = os.getenv("INPUT_LENGTH", "middle")
SERVICE_NAME = os.getenv("ServiceName")

def create_channel():
    # wait for the service to be ready
    while True:
        try:
            channel = grpc.insecure_channel(SERVICE_NAME)
            grpc.channel_ready_future(channel).result(timeout=10)
            print(f"Successfully connected to {SERVICE_NAME}")
            return channel
        except Exception as e:
            print(f"Connection failed {e}")
            time.sleep(10)

def create_stub():
    try:
        channel = create_channel()
        stub = predict_pb2_grpc.PredictorStub(channel)
    except Exception as e:
        print(f"Failed to create channel: {e}")
    return stub

# Prepare virtual inputs to get input size
short_input = "This is a test"
middle_input = "This is a test, how are you today?"
long_input = "This is a test, how are you today? Can you tell me what is the weather in Beijing?"
timeout = 5000
print("creating stub")
stub = create_stub()
print("stub created")

# Define request generation function
def generate_batch_request(input_length, batch_size):
    request = predict_pb2.PredictRequest()
    if input_length == "short":
        selected_input = short_input
    elif input_length == "middle":
        selected_input = middle_input
    else:
        selected_input = long_input
    request.input_data = json.dumps({"text": [selected_input] * batch_size}).encode('utf-8')
    return request

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
            result = stub.Predict(item.request, timeout)
            print("predict send in:", time.strftime("%Y-%m-%d %H:%M:%S"))
        except Exception as e:
            print("error: ", str(e))
            with self._condition:
                self._error += 1
        finally:
            response_time = time.time() - item.start
            print("response_time: ", response_time)
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
        num_tests: Number of test to use.

    Returns:
        The QPS and classification error rate.
    """
    error_rate_list = []
    rt_list = []
    qps_list = []

    # prepare the request
    request = generate_batch_request(input_length=INPUT_LENGTH, batch_size=BATCH_SIZE)

    for _ in range(num_tests):
        runner = QueueRunner(4)

        start_time = time.time()
        time_wait = 1.0 / qps
        for _ in range(max(math.ceil(qps), 4)):
            runner.enqueue(request=request)
            time.sleep(time_wait)

        runner.finish()
        total_time = time.time() - start_time
        error_rate_list.append(runner.get_error_rate())
        rt_list.append(runner.get_avg_rt())
        qps_list.append(runner.get_pass_items() / total_time)

    return np.mean(error_rate_list), np.mean(rt_list), np.mean(qps_list)

# Main function
def main():
    db_name = "morphling-db-manager"
    db_namespace = os.environ["DBNamespace"]
    db_port = os.environ["DBPort"]
    manager_server = "%s.%s:%s" % (
        db_name,
        db_namespace,
        db_port,
    ) 
    channel_manager = grpc.insecure_channel(manager_server)

    qps_max = None
    qps_current = 10 # initial qps
    qps_previous = 0
    rt_slo = 1.0

    # service is available
    while True:
        if qps_current < 1:
            break
        error_rate, rt, qps_real = do_inference(
            num_tests=NUM_TESTS, qps=qps_current
        )
        
        print(
            "\nQPS: %s, QPS_real: %s, Inference error rate: %s%%, RT: %s"
            % (qps_current, qps_real, error_rate * 100, np.mean(rt))
        )

        qps_current = qps_real
        if qps_previous > 0 and 1.2 > qps_current / qps_previous > 0.8:
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
            
        print("\n next QPS: %s" % qps_current)

    qps_previous = int(qps_previous * BATCH_SIZE)

    mls = []
    ml = db_pb2.KeyValue(key="qps", value=str(qps_previous))
    mls.append(ml)

    print(mls)
    stub_ = db_pb2_grpc.DBStub(channel_manager)
    result = stub_.SaveResult(
        db_pb2.SaveResultRequest(
            trial_name=os.environ["TrialName"],
            namespace=os.environ["Namespace"],
            results=mls,
        ),
        timeout=10,
    )
    print(result)
    
if __name__ == "__main__":
    main()
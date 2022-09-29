# make sure you use grpc version 1.39.0 or later,
# because of https://github.com/grpc/grpc/issues/15880 that affected earlier versions
import grpc
from locust import events, User, task
from locust.exception import LocustError
from locust.user.task import LOCUST_STATE_STOPPING
import numpy as np
import gevent
import time
import os

# patch grpc so that it uses gevent instead of asyncio
import grpc.experimental.gevent as grpc_gevent
grpc_gevent.init_gevent()
failratio_limit = float(os.getenv("FAIL_RATIO", 0.2))

class GrpcClient:
    def __init__(self, environment, stub, output):
        self.env = environment
        self._stub_class = stub.__class__
        self._stub = stub
        self._outputs = output

    def __getattr__(self, name):
        func = self._stub_class.__getattribute__(self._stub, name)

        def wrapper(*args, **kwargs):
            request_meta = {
                "request_type": "grpc",
                "name": name,
                "start_time": time.time(),
                "response_length": 0,
                "exception": None,
                "context": None,
                "response": None,
            }
            start_perf_counter = time.perf_counter()
            try:
                result = func(*args, **kwargs)
                response = np.array(result.outputs[self._outputs].float_val)
                prediction = np.argmax(response)
                print("Prediction:", prediction)
                request_meta["response"] = result
            except grpc.RpcError as e:
                request_meta["exception"] = e
            request_meta["response_time"] = (time.perf_counter() - start_perf_counter) * 1000
            self.env.events.request.fire(**request_meta)
            return request_meta["response"]

        return wrapper

class GrpcUser(User):
    abstract = True
    stub_class = None

    def __init__(self, environment):
        super().__init__(environment)
        for attr_value, attr_name in ((self.stub_class, "stub_class"),):
            if attr_value is None:
                raise LocustError(f"You must specify the {attr_name}.")
        self._channel = grpc.insecure_channel(self.FLAGS.server)
        self._channel_closed = False
        stub = self.stub_class(self._channel)
        self.client = GrpcClient(environment, stub, self.FLAGS.outputs)

    def stop(self, force=False):
       self._channel_closed = True
       time.sleep(1)

# Add a FailRatio Listener
from locust.runners import STATE_STOPPING, STATE_STOPPED, STATE_CLEANUP, MasterRunner, LocalRunner
def checker(environment):
    while not environment.runner.state in [STATE_STOPPING, STATE_STOPPED, STATE_CLEANUP]:
        time.sleep(1)
        if environment.runner.stats.total.fail_ratio > failratio_limit:
            print(f"fail ratio was {environment.runner.stats.total.fail_ratio}, quitting")
            environment.runner.quit()
            if environment.web_ui != None:
                environment.web_ui.stop()
            return

@events.init.add_listener
def on_locust_init(environment, **_kwargs):
    # dont run this on workers, we only care about the aggregated numbers
    if isinstance(environment.runner, MasterRunner) or isinstance(environment.runner, LocalRunner):
        gevent.spawn(checker, environment)
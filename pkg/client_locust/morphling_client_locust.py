from __future__ import print_function

import math
import os
import threading
import time
import numpy as np

import grpc
import api_pb2
import api_pb2_grpc

import invokust

# ResultDB Settings
db_name = "morphling-db-manager"
db_namespace = os.environ["DBNamespace"]
db_port = os.environ["DBPort"]
manager_server = "%s.%s:%s" % (
    db_name,
    db_namespace,
    db_port,
)  # "morphling-db-manager.morphling-system:6799"
channel_manager = grpc.insecure_channel(manager_server)
timeout_in_seconds = 10
rt_slo = 1.0 # currently slo is not guaranteed
batch_size = int(os.getenv("BATCH_SIZE", 1))
printlog = os.getenv("PRINTLOG", 'False').lower() in ('true', '1', 't')

# Locust Settings
settings = invokust.create_settings(
    locustfile=os.getenv("LOCUST_LOCUSTFILE","locustfile.py"),
    num_users=os.getenv("LOCUST_NUM_USERS", 10),
    spawn_rate=os.getenv("LOCUST_SPAWN_RATE", 10),
    run_time=os.getenv("LOCUST_RUN_TIME", 15),
    metrics_export=os.getenv("LOCUST_METRICS_EXPORT", 'False').lower() in ('true', '1', 't'),
    loglevel=os.getenv("LOCUST_LOGLEVEL", "INFO")
    )

def do_inference():
    """Tests PredictionService with concurrent requests.

    Returns:
        The QPS and classification error rate.
    """
    loadtest = invokust.LocustLoadTest(settings)
    loadtest.run()
    stats = loadtest.stats()
    mean_error_rate = stats['num_requests_fail']/stats['num_requests'] if stats['num_requests'] > 0 else 0 
    stats = stats["requests"]['grpc_Predict']
    return   mean_error_rate, stats['median_response_time'], stats['total_rps'] * batch_size

def main():
    error_rate, rt, qps_real = do_inference()

    if printlog:
        print(
            "\nQPS_real: %s, Inference error rate: %s%%, RT: %s"
            % (qps_real, error_rate * 100, np.mean(rt))
        )
    mls = []
    ml = api_pb2.KeyValue(key="qps", value=str(qps_real))
    mls.append(ml)

    stub_ = api_pb2_grpc.DBStub(channel_manager)
    result = stub_.SaveResult(
        api_pb2.SaveResultRequest(
            trial_name=os.environ["TrialName"],
            namespace=os.environ["Namespace"],
            results=mls,
        ),
        timeout=timeout_in_seconds,
    )
    if printlog:
        print(result)

if __name__ == "__main__":
    main()

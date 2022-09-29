# make sure you use grpc version 1.39.0 or later,
# because of https://github.com/grpc/grpc/issues/15880 that affected earlier versions
import os
import grpc
from locust import task

import tensorflow as tf
from tensorflow_serving.apis import predict_pb2, prediction_service_pb2_grpc
from locust_grpc import GrpcUser

reserved_models = {
    "densenet121": [224, 224],
    "densenet169": [224, 224],
    "densenet201": [224, 224],
    "efficientnetb0": [224, 224],
    "efficientnetb1": [240, 240],
    "efficientnetb2": [260, 260],
    "efficientnetb3": [300, 300],
    "efficientnetb4": [380, 380],
    "efficientnetb5": [456, 456],
    "efficientnetb6": [528, 528],
    "efficientnetb7": [600, 600],
    "inceptionresnetv2": [299, 299],
    "inceptionv3": [299, 299],
    "mobilenet": [224, 224],
    "mobilenetv2": [224, 224],
    "nasnetlarge": [331, 331],
    "nasnetmobile": [224, 224],
    "resnet101": [224, 224],
    "resnet152": [224, 224],
    "resnet50": [224, 224],
    "resnet101v2": [224, 224],
    "resnet152v2": [224, 224],
    "resnet50v2": [224, 224],
    "vgg16": [224, 224],
    "vgg19": [224, 224],
    "xception": [299, 299],
}

def parse_flags():
    with tf.device("/cpu:0"):
        tf.get_logger().setLevel("ERROR")

        tf.compat.v1.app.flags.DEFINE_integer(
            "concurrency", 100000, "maximum number of concurrent inference requests"
        )
        tf.compat.v1.app.flags.DEFINE_integer(
            "num_tests", 3, "Number of test images per test"
        )
        tf.compat.v1.app.flags.DEFINE_integer(
            "batch_size", os.environ["BATCH_SIZE"], "Number of test images per query"
        )
        tf.compat.v1.app.flags.DEFINE_string(
            "server", os.environ["ServiceName"], "PredictionService host:port"
        )
        tf.compat.v1.app.flags.DEFINE_string("image", "", "path to imxage in JPEG format")
        tf.compat.v1.app.flags.DEFINE_string(
            "model", os.environ["MODEL_NAME"], "model name"
        )
        tf.compat.v1.app.flags.DEFINE_string(
            "signature", "serving_default", "signature name"
        )
        tf.compat.v1.app.flags.DEFINE_string("inputs", "inputs", "signatureDef for inputs")
        tf.compat.v1.app.flags.DEFINE_string(
            "outputs", "predictions", "signatureDef for outputs"
        )
        tf.compat.v1.app.flags.DEFINE_enum(
            "task", default="cv", enum_values=["cv", "nlp"], help="which type of task"
        )
        tf.compat.v1.app.flags.DEFINE_bool(
            "printLog", True, "whether to print temp results"
        )
        return tf.compat.v1.app.flags.FLAGS

class TensorflowGrpcUser(GrpcUser):
    FLAGS = parse_flags()
    assert FLAGS.num_tests <= 10000
    assert FLAGS.server != ""
    stub_class = prediction_service_pb2_grpc.PredictionServiceStub

    @task
    def predict(self):
        if not self._channel_closed:
            if self.FLAGS.task == "cv":
                with open("./image.jpg", "rb") as f:
                    data = f.read()
                data = tf.image.decode_jpeg(data)
                data = tf.image.convert_image_dtype(data, dtype=tf.float32)
                data = tf.image.resize(data, reserved_models[self.FLAGS.model])
                data = tf.expand_dims(data, axis=0)
            elif self.FLAGS.task == "nlp":
                data = tf.convert_to_tensor(["This is a test!"])
            data = tf.concat([data] * self.FLAGS.batch_size, axis=0)
            request = predict_pb2.PredictRequest()
            request.model_spec.name = self.FLAGS.model # 'resnet50'
            request.model_spec.signature_name = self.FLAGS.signature
            request.inputs[self.FLAGS.inputs].CopyFrom(
                tf.make_tensor_proto(data, shape=list(data.shape))
            )
            timeout = 100 # second
            self.client.Predict(request, timeout)



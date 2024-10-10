FROM python:3.6

ENV TARGET_DIR /morphling
ENV SUGGESTION_DIR cmd/algorithm/grid
ENV PYTHONUNBUFFERED 0

RUN if [ "$(uname -m)" = "ppc64le" ] || [ "$(uname -m)" = "aarch64" ]; then \
    apt-get -y update && \
    apt-get -y install gfortran libopenblas-dev liblapack-dev && \
    pip install cython 'numpy>=1.13.3'; \
    fi
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.28 && \
    if [ "$(uname -m)" = "ppc64le" ]; then \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-ppc64le; \
    elif [ "$(uname -m)" = "aarch64" ]; then \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-arm64; \
    else \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64; \
    fi && \
    chmod +x /bin/grpc_health_probe

ADD ./pkg/ ${TARGET_DIR}/pkg/
ADD ./api/ ${TARGET_DIR}/api/
ADD ./${SUGGESTION_DIR}/ ${TARGET_DIR}/${SUGGESTION_DIR}/
WORKDIR  ${TARGET_DIR}/${SUGGESTION_DIR}
RUN pip install --no-cache-dir -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple

RUN chgrp -R 0 ${TARGET_DIR} \
  && chmod -R g+rwX ${TARGET_DIR}

ENV PYTHONPATH ${TARGET_DIR}:${TARGET_DIR}/api/v1alpha1/grpc_proto/grpc_algorithm/python3:${TARGET_DIR}/api/v1alpha1/grpc_proto/health/python

ENTRYPOINT ["python", "main.py"]

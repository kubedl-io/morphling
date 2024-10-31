FROM nvidia/cuda:11.8.0-cudnn8-runtime-ubuntu22.04

# Install Python and pip
RUN apt-get update && apt-get install -y python3-pip

# Set Python mirror
RUN pip3 config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple

# Copy .whl file into the container
ARG VLLM_FILE
ARG CUDA_VERSION
COPY ${VLLM_FILE} /tmp/

# Install .whl file
RUN pip install /tmp/${VLLM_FILE} --extra-index-url https://download.pytorch.org/whl/cu${CUDA_VERSION}

# Remove .whl file to reduce image size
RUN rm /tmp/${VLLM_FILE}

# Set working directory
WORKDIR /workspace

# Install dependencies
COPY requirements.txt /workspace/
RUN pip install --no-cache-dir -r /workspace/requirements.txt

# Expose port 8500 for gRPC
EXPOSE 8500
EXPOSE 8501

# Copy the rest of the files but not .kubedl_model_cache
COPY ./*.py /workspace/

# Set environment variables
ENV PYTHONUNBUFFERED=1 
ENV HF_ENDPOINT=https://hf-mirror.com

# Create model cache directory
RUN mkdir -p /workspace/.kubedl_model_cache
ENV HF_HOME=/workspace/.kubedl_model_cache

# Set ENTRYPOINT to execute the Python script
ENTRYPOINT ["python3", "main.py"]
import os
import time
import torch
import json
import torch.quantization
import grpc
from vllm import LLM
from concurrent import futures
import predict_pb2, predict_pb2_grpc
from huggingface_hub import snapshot_download
from transformers import AutoModelForCausalLM, AutoTokenizer, BitsAndBytesConfig

global model
global tokenizer
global inference_framework

def load_model_with_cache(model_name, cache_dir, data_type="fp16"):
    snapshot_download(model_name, cache_dir=cache_dir, local_files_only=False, max_workers=16, )
    model = None
    tokenizer = None
    if inference_framework == "torch":
        model = AutoModelForCausalLM.from_pretrained(
            model_name,
            trust_remote_code=True,
            cache_dir=cache_dir,
            quantization_config=BitsAndBytesConfig(load_in_8bit=True) if data_type == "int8" else None,
        )
        tokenizer = AutoTokenizer.from_pretrained(model_name, cache_dir=cache_dir)
    elif inference_framework == "vllm":
        model = LLM(
            model=model_name,
            trust_remote_code=True,
            download_dir=cache_dir,
            quantization="bitsandbytes" if data_type == "int8" else None,
            load_format="bitsandbytes" if data_type == "int8" else "auto"
        )
    return model, tokenizer

class PredictionService(predict_pb2_grpc.PredictorServicer):
    def Predict(self, request, context):
        result = []
        print("receive in:", time.strftime("%Y-%m-%d %H:%M:%S"))
        try:
            request_data = json.loads(request.input_data)
            prompts = [str(text) for text in request_data["text"]]
            if inference_framework == "torch":
                tokenizer.pad_token_id = tokenizer.eos_token_id
                input_data = tokenizer.batch_encode_plus(
                    prompts,
                    padding=True,
                    truncation=True,
                    max_length=200,
                    return_tensors="pt"
                ).to(model.device)
                outputs = model.generate(
                    input_ids=input_data.input_ids, 
                    attention_mask=input_data.attention_mask,
                    max_new_tokens=50
                )
                outputs = tokenizer.decode(outputs[0], skip_special_tokens=True)
                print("Decoding completed in:", time.strftime("%Y-%m-%d %H:%M:%S"))
            elif inference_framework == "vllm":
                outputs = model.generate(prompts)
                for output in outputs:
                    prompt = output.prompt
                    generated_text = output.outputs[0].text
                    result.append({
                        "text": output.outputs[0].text
                    })
                    print(f"Prompt: {prompt!r}, Generated text: {generated_text!r}")
            else:
                raise ValueError(f"Unsupported inference framework: {inference_framework}")
            
            response = predict_pb2.PredictResponse()
            response.output_data = json.dumps(result).encode('utf-8')
            return response
        except Exception as e:
            print("Error during inference: ", str(e))
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return predict_pb2.PredictResponse()


if __name__ == "__main__":
    model_name = os.getenv("MODEL_NAME", "huggyllama/llama-7b")
    inference_framework = os.getenv("INFERENCE_FRAMEWORK", "vllm")
    data_type = os.getenv("DTYPE", "int8")
    cache_dir = os.getenv("HF_HOME", ".kubedl_model_cache")
    
    model, tokenizer = load_model_with_cache(model_name, cache_dir=cache_dir, data_type=data_type)
    print("Model loaded")
    
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    predict_pb2_grpc.add_PredictorServicer_to_server(PredictionService(), server)
    server.add_insecure_port('0.0.0.0:8500')
    server.start()
    print("gRPC server started, listening on port 8500")
    server.wait_for_termination()
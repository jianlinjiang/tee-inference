# build llama cpp
ARG UBUNTU_VERSION=24.04
FROM ubuntu:$UBUNTU_VERSION AS build

RUN apt-get update && \
    apt-get install -y build-essential git cmake libcurl4-openssl-dev

WORKDIR /app
RUN git clone --depth 1 https://github.com/ggml-org/llama.cpp.git
WORKDIR /app/llama.cpp
RUN cmake -S . -B build -DCMAKE_BUILD_TYPE=Release -DGGML_NATIVE=OFF -DGGML_BACKEND_DL=ON -DGGML_CPU_ALL_VARIANTS=ON \
    && cmake --build build -j $(nproc)

RUN mkdir -p /app/lib && \
    find build -name "*.so" -exec cp {} /app/lib \;
    
RUN mkdir -p /app/full \
    && cp build/bin/* /app/full \
    && cp *.py /app/full \
    && cp -r gguf-py /app/full \
    && cp -r requirements /app/full \
    && cp requirements.txt /app/full \
    && cp .devops/tools.sh /app/full/tools.sh

RUN mkdir -p /app/lib && \
    find build -name "*.so" -exec cp {} /app/lib \;

FROM golang:1.24 AS build2
WORKDIR /app
COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o tee-inference

FROM ubuntu:$UBUNTU_VERSION AS base


RUN apt-get update \
    && apt-get install -y libgomp1 curl \
    && apt autoremove -y \
    && apt clean -y \
    && rm -rf /tmp/* /var/tmp/* \
    && find /var/cache/apt/archives /var/lib/apt/lists -not -name lock -type f -delete \
    && find /var/cache -type f -delete


WORKDIR /app
COPY --from=build /app/lib/ /app
COPY --from=build /app/full/llama-server /app
COPY --from=build2 /app/tee-inference /app
RUN mkdir /app/models
COPY ./Qwen3-1.7B-Q4_K_M.gguf /app/models/Qwen3-1.7B-Q4_K_M.gguf
EXPOSE 8085
LABEL tee.launch_policy.allow_env_override="PORT,MODEL"
LABEL "tee.launch_policy.allow_cmd_override"="true"
ENTRYPOINT ["/app/tee-inference"]
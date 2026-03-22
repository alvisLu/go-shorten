# Stage 1: Build
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates gcc g++ musl-dev cmake make git

# Clone whisper.cpp and build as static library
RUN git clone https://github.com/ggerganov/whisper.cpp.git /whisper.cpp && \
    cd /whisper.cpp && git checkout 9386f2394010 && \
    cmake -B build -DBUILD_SHARED_LIBS=OFF -DWHISPER_BUILD_TESTS=OFF -DWHISPER_BUILD_EXAMPLES=OFF && \
    cmake --build build --config Release

ENV CGO_CFLAGS="-I/whisper.cpp/include -I/whisper.cpp/ggml/include"
ENV CGO_LDFLAGS="-L/whisper.cpp/build/src -L/whisper.cpp/build/ggml/src -lwhisper -lggml -lggml-base -lggml-cpu -lstdc++ -lm"

WORKDIR /app

# Copy go.mod / go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and compile statically
COPY . .
RUN CGO_ENABLED=1 GOOS=linux \
    go build -ldflags="-s -w -extldflags '-static'" -o server ./cmd/server

# Stage 2: Final image
FROM alpine:3

RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
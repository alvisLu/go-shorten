# Stage 1: Build
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates gcc g++ musl-dev cmake make git

WORKDIR /app

# Copy go.mod to extract whisper.cpp commit hash
COPY go.mod go.sum ./
RUN go mod download

# Clone whisper.cpp and build as static library
# Extract whisper.cpp commit hash from go.mod pseudo-version (e.g. v0.0.0-<timestamp>-<commit>)
RUN WHISPER_COMMIT=$(grep 'ggerganov/whisper.cpp' go.mod | awk '{print $2}' | awk -F'-' '{print $NF}') && \
    git clone https://github.com/ggerganov/whisper.cpp.git /whisper.cpp && \
    cd /whisper.cpp && git checkout $WHISPER_COMMIT && \
    cmake -B build -DBUILD_SHARED_LIBS=OFF -DWHISPER_BUILD_TESTS=OFF -DWHISPER_BUILD_EXAMPLES=OFF && \
    cmake --build build --config Release

ENV CGO_CFLAGS="-I/whisper.cpp/include -I/whisper.cpp/ggml/include"
ENV CGO_LDFLAGS="-L/whisper.cpp/build/src -L/whisper.cpp/build/ggml/src -lwhisper -lggml -lggml-base -lggml-cpu -lstdc++ -lm"

# Copy source code and compile statically
COPY . .
RUN CGO_ENABLED=1 GOOS=linux \
    go build -ldflags="-s -w -extldflags '-static'" -o server ./cmd/server

# Stage 2: Final image
FROM alpine:3

RUN apk add --no-cache ca-certificates curl
COPY --from=builder /app/server /server
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/entrypoint.sh"]
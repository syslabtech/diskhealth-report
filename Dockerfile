# Stage 1: Build the Go binary
FROM golang:1.23-bookworm AS build

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies first
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy the rest of the application source code
COPY . .

# Install xz-utils for extracting .tar.xz, curl for downloading UPX, and other dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    xz-utils \
    curl \
    librdkafka-dev \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# Download UPX manually and install it
RUN curl -L https://github.com/upx/upx/releases/download/v4.0.2/upx-4.0.2-amd64_linux.tar.xz -o upx.tar.xz \
    && tar -xf upx.tar.xz \
    && cp upx-4.0.2-amd64_linux/upx /usr/local/bin/ \
    && rm -rf upx.tar.xz upx-4.0.2-amd64_linux

# Build the Go binary for a Linux target
RUN CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -o disktool -a -ldflags="-s -w" -installsuffix cgo

# Compress the binary using UPX
RUN upx --ultra-brute -qq disktool && upx -t disktool 


# Stage 2: Create a lightweight final image using Alpine
FROM debian:bookworm-slim

# Set the working directory inside the container
WORKDIR /app

# Install necessary runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    librdkafka-dev \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# Copy the compressed binary from the build stage
COPY --from=build /app/disktool /app/disktool
COPY --from=build /app/config /app/config
COPY --from=build /app/deviceTBW.json /app
COPY --from=build /app/outputdiskdata.txt /app

# Add labels for metadata
LABEL org.opencontainers.image.title="Disk Health Analyzer Tool" \
      org.opencontainers.image.description="A lightweight Go application for managing disk operations." \
      org.opencontainers.image.authors="syslabtech"

# Expose any necessary ports (if needed, e.g., EXPOSE 8080)
# EXPOSE 8080

# Command to run the application
CMD ["/app/disktool"]

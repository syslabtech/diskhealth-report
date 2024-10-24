# Use a valid Go version and base image
FROM golang:1.23-bookworm AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the rest of the application source code
COPY . .

RUN ls -ltr

# Download dependencies
RUN go mod download && go mod verify

# Install xz-utils for extracting .tar.xz and curl to download the UPX binary
RUN apt-get update && apt-get install -y xz-utils curl librdkafka-dev tzdata

# Download UPX manually and install it
RUN curl -L https://github.com/upx/upx/releases/download/v4.0.2/upx-4.0.2-amd64_linux.tar.xz -o upx.tar.xz \
    && tar -xf upx.tar.xz \
    && cp upx-4.0.2-amd64_linux/upx /usr/local/bin/ \
    && rm -rf upx.tar.xz upx-4.0.2-amd64_linux

# Build the Go binary for a Linux target
RUN CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -o disktool -a -ldflags="-s -w" -installsuffix cgo

# Compress the binary using UPX
RUN upx --ultra-brute -qq disktool && upx -t disktool 

# Stage 2: Use a minimal image for the final production build
# FROM gcr.io/distroless/base           26.3 MB
FROM alpine:latest

RUN apk add --no-cache bash tzdata curl

# Set the working directory
WORKDIR /app

# Copy the compressed binary and configuration files from the build stage
COPY --from=build /app/disktool /app
COPY --from=build /app/config /app/config
COPY --from=build /app/deviceTBW.json /app
COPY --from=build /app/outputdiskdata.txt /app

RUN ls -ltr

# Run the disktool binary directly
CMD ["/app/disktool"]

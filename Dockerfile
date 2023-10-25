FROM golang:latest

WORKDIR /

# Initialize a new Go module.
RUN go mod init virtual-friends-brain

# Copy local code to the container image.
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /virtual-friends-brain

# Update and install dependencies
RUN apt update && \
    apt install -y ffmpeg && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 8080
EXPOSE $PORT

# Run the web service on container startup.
ENTRYPOINT ["/virtual-friends-brain"]

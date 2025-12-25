
# Use the official Go image as a parent image
FROM golang:alpine

# Install necessary tools
RUN apk add --no-cache \
    git \
    curl \
    wget \
    unzip

# Install Delve for debugging
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Install Go tools
RUN go install -v golang.org/x/tools/cmd/goimports@latest

# Install gopls
RUN go install -v golang.org/x/tools/gopls@latest

# Create a non-root user to use the container
RUN useradd -ms /bin/bash vscode

#RUN mkdir -p /go && chown -R vscode:vscode /go
#RUN mkdir -p /workspace && chown -R vscode:vscode /workspace


# Set the user to vscode
USER root 
#USER vscode

# Set the working directory
WORKDIR /workspace

# Expose the port for Delve debugger
EXPOSE 2345

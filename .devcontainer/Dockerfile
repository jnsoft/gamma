
FROM docker.io/library/golang:alpine

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
RUN adduser -D -s /bin/bash vscode
#RUN apk add --no-cache shadow && useradd -ms /bin/bash vscode

#RUN mkdir -p /go && chown -R vscode:vscode /go
#RUN mkdir -p /workspace && chown -R vscode:vscode /workspace


# Set the user to vscode
USER root 
#USER vscode

# Set the working directory
WORKDIR /workspace

# Expose the port for Delve debugger
EXPOSE 2345

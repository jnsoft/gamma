FROM golang
WORKDIR /App
COPY . ./
ENTRYPOINT ["/bin/bash"]
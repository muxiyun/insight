FROM golang:1.11.0

# Create app directory
RUN mkdir -p $GOPATH/src/github.com/muxiyun/insight
COPY . $GOPATH/src/github.com/muxiyun/insight
WORKDIR $GOPATH/src/github.com/muxiyun/insight

# Build static file
RUN go build foo.go

# Expose the application on port 8080
EXPOSE 8080

# Set the entry point of the container to the bee command that runs the
# application and watches for changes
CMD ["./foo"]
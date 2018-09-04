FROM golang:1.11.0

# Create app directory
RUN mkdir -p /usr/src/app

# Build static file
RUN go get github.com/gin-gonic/gin github.com/influxdata/influxdb/client/v2 github.com/jinzhu/gorm github.com/jinzhu/gorm/dialects/mysql
RUN go build foo.go

COPY . /usr/src/app
WORKDIR /usr/src/app

# Expose the application on port 8080
EXPOSE 8080

# Set the entry point of the container to the bee command that runs the
# application and watches for changes
CMD ["./foo"]
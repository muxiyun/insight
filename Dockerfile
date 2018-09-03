FROM golang:1.11.0

RUN go get ithub.com/gin-gonic/gin github.com/influxdata/influxdb/client/v2 github.com/jinzhu/gorm github.com/jinzhu/gorm/dialects/mysql

# Expose the application on port 8080
EXPOSE 8080

RUN go build foo.go

# Set the entry point of the container to the bee command that runs the
# application and watches for changes
CMD ["./foo"]
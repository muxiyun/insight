package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/influxdata/influxdb/client/v2"
)

// Binding from JSON
type ReportData struct {
	DeviceId  string `json:"deviceId" binding:"required"`
	ProductId string `json:"pid" binding:"required"`
	Type      string `json:"type" binding:"required"`
	MainCat   string `json:"mainCat" binding:"required"`
	SubCat    string `json:"subCat" binding:"required"`
	Extra     string `json:"extra" binding:"required"`
	Value     string `json:"value" binding:"required"`
	Timestamp int64  `json:"timestamp" binding:"required"`
}

type ReportDataReq struct {
	Data []ReportData `json:"data" binding:"required"`
}

const (
	DB_NAME  = "mydb" // os.Getenv("INSIGHT_INFLUX_DB_NAME")
	username = "foo"  // os.Getenv("INSIGHT_INFLUX_USER_NAME")
	password = "bar"  // os.Getenv("INSIGHT_INFLUX_PASSWORD")
)

func addBatchPoint(json *ReportData, bp client.BatchPoints) {
	var tags map[string]string
	tags = make(map[string]string)

	tags["pid"] = json.ProductId
	tags["did"] = json.DeviceId
	tags["mianCat"] = json.MainCat
	tags["subCat"] = json.SubCat
	tags["type"] = json.Type

	fields := map[string]interface{}{
		"value": json.Value,
		"extra": json.Extra,
	}

	time := time.Unix(json.Timestamp, 0)

	pt, err := client.NewPoint("data1", tags, fields, time)
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)
}

func main() {
	// Create a new HTTPClient
	influxClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: username,
		Password: password,
	})

	if err != nil {
		log.Fatal(err)
	}
	defer influxClient.Close()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/statistics", func(c *gin.Context) {
		var json ReportDataReq
		if err := c.ShouldBindJSON(&json); err == nil {
			// Create a new point batch
			bp, err := client.NewBatchPoints(client.BatchPointsConfig{
				Database:  DB_NAME,
				Precision: "s",
			})
			if err != nil {
				log.Fatal(err)
			}
			for _, element := range json.Data {
				// index is the index where we are
				// element is the element from someSlice for where we are
				addBatchPoint(&element, bp)
			}

			// Write the batch
			if err := influxClient.Write(bp); err != nil {
				log.Fatal(err)
			}

			// Close client resources
			if err := influxClient.Close(); err != nil {
				log.Fatal(err)
			}
			c.JSON(http.StatusOK, gin.H{"status": time.Now()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}

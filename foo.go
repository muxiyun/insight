package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Device struct {
	gorm.Model
	DeviceId  string `json:"did" binding:"required"`       // DeviceId
	Platform  string `json:"platform" binding:"required"`  //  Native/Web
	OS        string `json:"os" binding:"required"`        // iOS/Android/Windows/macOS
	UserAgent string `json:"userAgent"`                    // UA For Browser
	Sid       string `json:"sid" binding:"required"`       // Student Id
	OsVersion string `json:"osVersion" binding:"required"` // 11.0/12.0
	Pid       string `json:"pid" binding:"required"`       // ProductId
}

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

var (
	DB_NAME        = os.Getenv("INSIGHT_INFLUX_DB_NAME")
	username       = os.Getenv("INSIGHT_INFLUX_USER_NAME")
	password       = os.Getenv("INSIGHT_INFLUX_PASSWORD")
	MYSQL_DB_NAME  = os.Getenv("INSIGHT_MYSQL_DB_NAME")
	MYSQL_DB_URL   = os.Getenv("INSIGHT_MYSQL_DB_URL")
	mysql_username = os.Getenv("INSIGHT_MYSQL_USER_NAME")
	mysql_password = os.Getenv("INSIGHT_MYSQL_PASSWORD")
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
	print(mysql_username + ":" + mysql_username + "@tcp(" + MYSQL_DB_URL + ":3306)/" + MYSQL_DB_NAME + "?charset=utf8&parseTime=True&loc=Local")
	db, err := gorm.Open("mysql", mysql_username+":"+mysql_password+"@tcp("+MYSQL_DB_URL+":3306)/"+MYSQL_DB_NAME+"?charset=utf8&parseTime=True&loc=Local")
	defer db.Close()

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

	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&Device{})

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/device", func(c *gin.Context) {
		var json Device
		if err := c.ShouldBindJSON(&json); err == nil {
			var device Device
			if err := db.Where("device_id = ?", json.DeviceId).First(&device).Error; err != nil {
				db.Create(&json)
			} else {
				db.Delete(&device)
				db.Create(&json)
			}

			c.JSON(http.StatusOK, gin.H{"status": time.Now()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
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

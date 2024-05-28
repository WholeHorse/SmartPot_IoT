package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"github.com/gin-gonic/gin"

	_ "github.com/lib/pq"
)

var db *sql.DB

type Sensor struct {
	ID     string  `json:"id"`
	Type   string  `json:"type"`
	Value  float64 `json:"value"`
	Status string  `json:"status"`
}

func getSensors(c *gin.Context) {
	sensors, err := loadSensorsFromDatabase()
	if err != nil {
		log.Printf("Error loading sensors from database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sensors)
}

func addSensor(c *gin.Context) {
	var newSensor Sensor
	if err := c.BindJSON(&newSensor); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := addSensorToDatabase(newSensor)
	if err != nil {
		log.Printf("Error adding sensor to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, newSensor)
}

func connectToDatabase() error {
	connStr := "user=iot_admin dbname=iot-db sslmode=disable password=iotpass host=localhost port=5432"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Error opening database connection: %v", err)
		return err
	}
	err = db.Ping()
	if err != nil {
		log.Printf("Error pinging database: %v", err)
		return err
	}
	return nil
}

func addSensorToDatabase(sensor Sensor) error {
	_, err := db.Exec("INSERT INTO sensors (id, type, value, status) VALUES ($1, $2, $3, $4)", sensor.ID, sensor.Type, sensor.Value, sensor.Status)
	if err != nil {
		log.Printf("Error executing SQL query: %v", err)
	}
	return err
}

func deleteSensor(c *gin.Context) {
	id := c.Param("id")
	err := deleteSensorFromDatabase(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Sensor deleted"})
}

func deleteSensorFromDatabase(id string) error {
	_, err := db.Exec("DELETE FROM sensors WHERE id = $1", id)
	return err
}

func loadSensorsFromDatabase() ([]Sensor, error) {
	rows, err := db.Query("SELECT id, type, value, status FROM sensors")
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return nil, err
	}
	defer rows.Close()

	var sensors []Sensor
	for rows.Next() {
		var s Sensor
		err = rows.Scan(&s.ID, &s.Type, &s.Value, &s.Status)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		sensors = append(sensors, s)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error with rows: %v", err)
		return nil, err
	}

	return sensors, nil
}

func main() {
	err := connectToDatabase()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	r := gin.Default()
	r.GET("/sensors", getSensors)
	r.POST("/sensors/add", addSensor)
	r.DELETE("/sensors/delete/:id", deleteSensor)

	go func() {
		if err := r.Run(":8080"); err != nil {
			log.Fatal("Server Shutdown:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
}

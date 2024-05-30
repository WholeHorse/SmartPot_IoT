package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
    "strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var db *sql.DB

type Sensor struct {
    ID     string  `json:"id"`
    Type   string  `json:"type"`
    Value  float64 `json:"value"`
    Status string  `json:"status"`
    PotID  int     `json:"pot_id"`
}

type Device struct {
    ID     string `json:"id"`
    Type   string `json:"type"`
    Status string `json:"status"`
    PotID  int    `json:"pot_id"`
}


type Pot struct {
    ID      int       `json:"id"`
    Name    string    `json:"name"`
    Sensors []Sensor  `json:"sensors"`
    Devices []Device  `json:"devices"`
}

func addPot(c *gin.Context) {
    var newPot Pot
    if err := c.BindJSON(&newPot); err != nil {
        log.Printf("Error binding JSON: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    err := addPotToDatabase(&newPot)
    if err != nil {
        log.Printf("Error adding pot to database: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    logAction(strconv.Itoa(newPot.ID), "", "Pot added")
    c.JSON(http.StatusOK, newPot)
}

func addPotToDatabase(pot *Pot) error {
    var potID int
    err := db.QueryRow("INSERT INTO pots (name) VALUES ($1) RETURNING id", pot.Name).Scan(&potID)
    if err != nil {
        log.Printf("Error executing SQL query: %v", err)
        return err
    }
    pot.ID = potID
    return nil
}


func deletePot(c *gin.Context) {
    id := c.Param("id")
    err := deletePotFromDatabase(id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    logAction("", id, "Pot deleted")
    c.JSON(http.StatusOK, gin.H{"status": "Pot deleted"})
}

func deletePotFromDatabase(id string) error {
    _, err := db.Exec("DELETE FROM pots WHERE id = $1", id)
    if err != nil {
        log.Printf("Error executing SQL query: %v", err)
    }
    return err
}

func getPots(c *gin.Context) {
    pots, err := loadPotsFromDatabase()
    if err != nil {
        log.Printf("Error loading pots from database: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, pots)
}

func loadPotsFromDatabase() ([]Pot, error) {
    rows, err := db.Query("SELECT id, name FROM pots")
    if err != nil {
        log.Printf("Error querying database: %v", err)
        return nil, err
    }
    defer rows.Close()

    var pots []Pot
    for rows.Next() {
        var p Pot
        err = rows.Scan(&p.ID, &p.Name)
        if err != nil {
            log.Printf("Error scanning row: %v", err)
            return nil, err
        }

        p.Sensors, err = loadSensorsByPotID(p.ID)
        if err != nil {
            return nil, err
        }

        p.Devices, err = loadDevicesByPotID(p.ID)
        if err != nil {
            return nil, err
        }

        pots = append(pots, p)
    }

    if err = rows.Err(); err != nil {
        log.Printf("Error with rows: %v", err)
        return nil, err
    }

    return pots, nil
}

func loadSensorsByPotID(potID int) ([]Sensor, error) {
    rows, err := db.Query("SELECT id, type, value, status, pot_id FROM sensors WHERE pot_id = $1", potID)
    if err != nil {
        log.Printf("Error querying database: %v", err)
        return nil, err
    }
    defer rows.Close()

    var sensors []Sensor
    for rows.Next() {
        var s Sensor
        err = rows.Scan(&s.ID, &s.Type, &s.Value, &s.Status, &s.PotID)
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

func loadDevicesByPotID(potID int) ([]Device, error) {
    rows, err := db.Query("SELECT id, type, status, pot_id FROM devices WHERE pot_id = $1", potID)
    if err != nil {
        log.Printf("Error querying database: %v", err)
        return nil, err
    }
    defer rows.Close()

    var devices []Device
    for rows.Next() {
        var d Device
        err = rows.Scan(&d.ID, &d.Type, &d.Status, &d.PotID)
        if err != nil {
            log.Printf("Error scanning row: %v", err)
            return nil, err
        }
        devices = append(devices, d)
    }

    if err = rows.Err(); err != nil {
        log.Printf("Error with rows: %v", err)
        return nil, err
    }

    return devices, nil
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
    logAction(newSensor.ID, "", "Sensor added")
    c.JSON(http.StatusOK, newSensor)
}

func addSensorToDatabase(sensor Sensor) error {
    _, err := db.Exec("INSERT INTO sensors (id, type, value, status, pot_id) VALUES ($1, $2, $3, $4, $5)", sensor.ID, sensor.Type, sensor.Value, sensor.Status, sensor.PotID)
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
    // Log the action
    logAction(id, "", "Sensor deleted")
    c.JSON(http.StatusOK, gin.H{"status": "Sensor deleted"})
}

func deleteSensorFromDatabase(id string) error {
	_, err := db.Exec("DELETE FROM sensors WHERE id = $1", id)
	return err
}

func loadSensorsFromDatabase() ([]Sensor, error) {
	rows, err := db.Query("SELECT id, type, value, status, pot_id FROM sensors")
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return nil, err
	}
	defer rows.Close()

	var sensors []Sensor
	for rows.Next() {
		var s Sensor
		err = rows.Scan(&s.ID, &s.Type, &s.Value, &s.Status, &s.PotID)
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

func addDevice(c *gin.Context) {
    var newDevice Device
    if err := c.BindJSON(&newDevice); err != nil {
        log.Printf("Error binding JSON: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    err := addDeviceToDatabase(newDevice)
    if err != nil {
        log.Printf("Error adding device to database: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    logAction("", newDevice.ID, "Device added")
    c.JSON(http.StatusOK, newDevice)
}

func addDeviceToDatabase(device Device) error {
    _, err := db.Exec("INSERT INTO devices (id, type, status, pot_id) VALUES ($1, $2, $3, $4)", device.ID, device.Type, device.Status, device.PotID)
    if err != nil {
        log.Printf("Error executing SQL query: %v", err)
    }
    return err
}

func getDevices(c *gin.Context) {
    devices, err := loadDevicesFromDatabase()
    if err != nil {
        log.Printf("Error loading devices from database: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, devices)
}

func loadDevicesFromDatabase() ([]Device, error) {
    rows, err := db.Query("SELECT id, type, status, pot_id FROM devices")
    if err != nil {
        log.Printf("Error querying database: %v", err)
        return nil, err
    }
    defer rows.Close()

    var devices []Device
    for rows.Next() {
        var d Device
        err = rows.Scan(&d.ID, &d.Type, &d.Status, &d.PotID)
        if err != nil {
            log.Printf("Error scanning row: %v", err)
            return nil, err
        }
        devices = append(devices, d)
    }

    if err = rows.Err(); err != nil {
        log.Printf("Error with rows: %v", err)
        return nil, err
    }

    return devices, nil
}

func updateDeviceStatus(c *gin.Context) {
    id := c.Param("id")
    var status struct {
        Status string `json:"status"`
    }
    if err := c.BindJSON(&status); err != nil {
        log.Printf("Error binding JSON: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    err := updateDeviceStatusInDatabase(id, status.Status)
    if err != nil {
        log.Printf("Error updating device status in database: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    // Log the action
    logAction("", id, "Device status updated to "+status.Status)
    c.JSON(http.StatusOK, gin.H{"status": "Device status updated"})
}

func updateDeviceStatusInDatabase(id, status string) error {
	_, err := db.Exec("UPDATE devices SET status = $1 WHERE id = $2", status, id)
	if err != nil {
		log.Printf("Error executing SQL query: %v", err)
	}
	return err
}

func deleteDevice(c *gin.Context) {
    id := c.Param("id")
    err := deleteDeviceFromDatabase(id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    // Log the action
    logAction("", id, "Device deleted")
    c.JSON(http.StatusOK, gin.H{"status": "Device deleted"})
}

func deleteDeviceFromDatabase(id string) error {
    _, err := db.Exec("DELETE FROM devices WHERE id = $1", id)
    if err != nil {
        log.Printf("Error executing SQL query: %v", err)
    }
    return err
}

func logAction(sensorID, deviceID, action string) error {
    _, err := db.Exec("INSERT INTO logs (sensor_id, device_id, action) VALUES ($1, $2, $3)", sensorID, deviceID, action)
    if err != nil {
        log.Printf("Error logging action: %v", err)
    }
    return err
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

	r.GET("/devices", getDevices)
	r.POST("/devices/add", addDevice)
	r.PUT("/devices/update/:id", updateDeviceStatus)
	r.DELETE("/devices/delete/:id", deleteDevice)

    r.POST("/pots/add", addPot)
    r.DELETE("/pots/delete/:id", deletePot)
    r.GET("/pots", getPots)

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

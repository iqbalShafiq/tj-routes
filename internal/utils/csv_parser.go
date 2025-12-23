package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"tj-routes/internal/models"
)

// RouteCSVRow represents a row in the routes CSV file
type RouteCSVRow struct {
	RouteNumber string
	Name        string
	Description string
	Status      string
}

// StopCSVRow represents a row in the stops CSV file
type StopCSVRow struct {
	Name      string
	Type      string
	Latitude  float64
	Longitude float64
	Address   string
	City      string
	District  string
	Facilities string
	Status    string
}

// VehicleCSVRow represents a row in the vehicles CSV file
type VehicleCSVRow struct {
	VehiclePlate string
	RouteID      uint
	VehicleType  string
	Capacity     int
	Status       string
}

// ParseRoutesCSV parses a CSV file containing route data
func ParseRoutesCSV(filePath string) ([]RouteCSVRow, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least a header and one data row")
	}

	// Skip header row
	rows := make([]RouteCSVRow, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 3 {
			return nil, fmt.Errorf("row %d: insufficient columns (expected at least 3: route_number, name, description)", i+1)
		}

		row := RouteCSVRow{
			RouteNumber: strings.TrimSpace(record[0]),
			Name:        strings.TrimSpace(record[1]),
			Description: strings.TrimSpace(record[2]),
		}

		if len(record) > 3 {
			row.Status = strings.TrimSpace(record[3])
		}
		if row.Status == "" {
			row.Status = string(models.StatusActive)
		}

		// Validate required fields
		if row.RouteNumber == "" {
			return nil, fmt.Errorf("row %d: route_number is required", i+1)
		}
		if row.Name == "" {
			return nil, fmt.Errorf("row %d: name is required", i+1)
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// ParseStopsCSV parses a CSV file containing stop data
func ParseStopsCSV(filePath string) ([]StopCSVRow, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least a header and one data row")
	}

	// Skip header row
	rows := make([]StopCSVRow, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 4 {
			return nil, fmt.Errorf("row %d: insufficient columns (expected at least 4: name, type, latitude, longitude)", i+1)
		}

		row := StopCSVRow{
			Name:     strings.TrimSpace(record[0]),
			Type:     strings.TrimSpace(record[1]),
			Address:  "",
			City:     "",
			District: "",
		}

		// Parse latitude
		lat, err := strconv.ParseFloat(strings.TrimSpace(record[2]), 64)
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid latitude: %w", i+1, err)
		}
		row.Latitude = lat

		// Parse longitude
		lng, err := strconv.ParseFloat(strings.TrimSpace(record[3]), 64)
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid longitude: %w", i+1, err)
		}
		row.Longitude = lng

		// Optional fields
		if len(record) > 4 {
			row.Address = strings.TrimSpace(record[4])
		}
		if len(record) > 5 {
			row.City = strings.TrimSpace(record[5])
		}
		if len(record) > 6 {
			row.District = strings.TrimSpace(record[6])
		}
		if len(record) > 7 {
			row.Facilities = strings.TrimSpace(record[7])
		}
		if len(record) > 8 {
			row.Status = strings.TrimSpace(record[8])
		}
		if row.Status == "" {
			row.Status = string(models.StatusActive)
		}

		// Validate required fields
		if row.Name == "" {
			return nil, fmt.Errorf("row %d: name is required", i+1)
		}
		if row.Type == "" {
			return nil, fmt.Errorf("row %d: type is required", i+1)
		}
		if row.Type != string(models.StopTypeStop) && row.Type != string(models.StopTypeTerminal) {
			return nil, fmt.Errorf("row %d: type must be 'stop' or 'terminal'", i+1)
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// ParseVehiclesCSV parses a CSV file containing vehicle data
func ParseVehiclesCSV(filePath string) ([]VehicleCSVRow, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least a header and one data row")
	}

	// Skip header row
	rows := make([]VehicleCSVRow, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 2 {
			return nil, fmt.Errorf("row %d: insufficient columns (expected at least 2: vehicle_plate, route_id)", i+1)
		}

		row := VehicleCSVRow{
			VehiclePlate: strings.TrimSpace(record[0]),
			VehicleType:  "",
			Capacity:     0,
		}

		// Parse route_id
		routeID, err := strconv.ParseUint(strings.TrimSpace(record[1]), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid route_id: %w", i+1, err)
		}
		row.RouteID = uint(routeID)

		// Optional fields
		if len(record) > 2 {
			row.VehicleType = strings.TrimSpace(record[2])
		}
		if len(record) > 3 {
			capacity, err := strconv.Atoi(strings.TrimSpace(record[3]))
			if err == nil {
				row.Capacity = capacity
			}
		}
		if len(record) > 4 {
			row.Status = strings.TrimSpace(record[4])
		}
		if row.Status == "" {
			row.Status = string(models.StatusActive)
		}

		// Validate required fields
		if row.VehiclePlate == "" {
			return nil, fmt.Errorf("row %d: vehicle_plate is required", i+1)
		}
		if row.RouteID == 0 {
			return nil, fmt.Errorf("row %d: route_id must be greater than 0", i+1)
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// ValidateFacilitiesJSON validates and returns facilities as JSON string
func ValidateFacilitiesJSON(facilitiesStr string) (*string, error) {
	if facilitiesStr == "" {
		return nil, nil
	}

	// Try to parse as JSON to validate
	var facilities interface{}
	if err := json.Unmarshal([]byte(facilitiesStr), &facilities); err != nil {
		return nil, fmt.Errorf("invalid JSON format for facilities: %w", err)
	}

	// Return as JSON string
	jsonBytes, err := json.Marshal(facilities)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal facilities: %w", err)
	}

	jsonStr := string(jsonBytes)
	return &jsonStr, nil
}


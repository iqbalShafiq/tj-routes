# Bulk Upload CSV Format Guide

This directory contains example CSV files for bulk uploading routes, stops, and vehicles.

## Routes CSV Format

**File:** `routes_example.csv`

**Columns:**
- `route_number` (required): Unique route identifier (string)
- `name` (required): Route name (string)
- `description` (optional): Route description (string)
- `status` (optional): Route status - "active" or "inactive" (default: "active")

**Example:**
```csv
route_number,name,description,status
R001,City Center to Airport,Main route connecting city center to airport,active
R002,North Station to South Terminal,Route serving north and south terminals,active
```

**Notes:**
- Duplicate detection: Routes with the same `route_number` (case-insensitive) will be skipped
- The `route_number` must be unique

## Stops CSV Format

**File:** `stops_example.csv`

**Columns:**
- `name` (required): Stop name (string)
- `type` (required): Stop type - "stop" or "terminal" (string)
- `latitude` (required): Latitude coordinate (float)
- `longitude` (required): Longitude coordinate (float)
- `address` (optional): Street address (string)
- `city` (optional): City name (string)
- `district` (optional): District/neighborhood (string)
- `facilities` (optional): JSON string with facilities (e.g., `{"wifi":true,"restroom":true}`)
- `status` (optional): Stop status - "active" or "inactive" (default: "active")

**Example:**
```csv
name,type,latitude,longitude,address,city,district,facilities,status
Central Station,terminal,40.7128,-74.0060,123 Main St,New York,Manhattan,"{""wifi"":true,""restroom"":true}",active
Park Avenue Stop,stop,40.7580,-73.9855,456 Park Ave,New York,Manhattan,"{""shelter"":true}",active
```

**Notes:**
- Duplicate detection: Stops with the same `latitude` and `longitude` (within 0.0001 degrees tolerance ≈ 11 meters) will be skipped
- Facilities must be valid JSON format
- Coordinates should be in decimal degrees format

## Vehicles CSV Format

**File:** `vehicles_example.csv`

**Columns:**
- `vehicle_plate` (required): Vehicle license plate (string)
- `route_id` (required): ID of the route this vehicle is assigned to (integer, must exist)
- `vehicle_type` (optional): Type of vehicle (string, e.g., "Bus", "Minibus")
- `capacity` (optional): Passenger capacity (integer)
- `status` (optional): Vehicle status - "active" or "inactive" (default: "active")

**Example:**
```csv
vehicle_plate,route_id,vehicle_type,capacity,status
ABC-1234,1,Bus,50,active
XYZ-5678,1,Bus,45,active
DEF-9012,2,Minibus,25,active
```

**Notes:**
- Duplicate detection: Vehicles with the same `vehicle_plate` (case-insensitive) will be skipped
- The `route_id` must reference an existing route in the database
- If a route doesn't exist, the vehicle will be skipped and counted as an error

## Upload Process

1. **Upload CSV File**: Use the POST endpoint `/api/v1/bulk-upload/{entityType}` where `entityType` is one of: `route`, `stop`, or `vehicle`
2. **Check Status**: Use GET `/api/v1/bulk-upload/{id}` to check processing status
3. **View Results**: The response includes:
   - `total_rows`: Total number of rows in CSV
   - `success_count`: Number of successfully inserted records
   - `duplicate_count`: Number of duplicate records skipped
   - `error_count`: Number of records that failed to insert
   - `status`: Current processing status (pending, processing, completed, failed)

## Processing Behavior

- **Background Processing**: Files are processed asynchronously in the background
- **Batch Processing**: Records are processed in batches of 100 for performance
- **Resume on Restart**: If the app crashes, processing will resume from the last processed row
- **Duplicate Handling**: Duplicates are logged but not inserted
- **Error Handling**: Individual row errors don't stop the entire process

## Tips

1. Always include the header row in your CSV files
2. Ensure required fields are filled for all rows
3. For stops, use valid JSON format for facilities field
4. For vehicles, ensure route_id references exist before uploading
5. Check the upload status periodically to monitor progress


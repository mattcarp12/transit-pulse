package transit

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
)

// The official BART static GTFS schedule URL
const StaticScheduleURL = "https://www.bart.gov/dev/schedules/google_transit.zip"

// Stop represents a physical station on the transit network.
type Stop struct {
	ID   string
	Name string
	Lat  float64
	Lon  float64
}

// ShapePoint represents a single GPS coordinate along a physical track.
type ShapePoint struct {
	Lat      float64
	Lon      float64
	Sequence int
}

type StaticData struct {
	Stops  map[string]Stop
	Shapes map[string][]ShapePoint
	Trips  map[string]Trip
	Routes map[string]Route
}

// FetchStaticData downloads the GTFS zip, extracts stops.txt in memory,
// and returns a map for O(1) instantaneous lookups by Stop ID.
func (c *Client) FetchStaticData(ctx context.Context) (StaticData, error) {
	// 1. Download the zip file
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, StaticScheduleURL, nil)
	if err != nil {
		return StaticData{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return StaticData{}, fmt.Errorf("failed to download static schedule: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return StaticData{}, fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}

	// 2. Read the entire zip into memory
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return StaticData{}, fmt.Errorf("failed to read zip body: %w", err)
	}

	// 3. Open the zip archive from memory
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return StaticData{}, fmt.Errorf("failed to create zip reader: %w", err)
	}

	stops := make(map[string]Stop)
	routeShapes := make(map[string][]ShapePoint)
	trips := make(map[string]Trip)
	routes := make(map[string]Route)

	// 4. Find and parse CSV files we care about (stops.txt and shapes.txt)
	for _, file := range zipReader.File {
		switch file.Name {
		case "stops.txt":
			stops, err = parseStopsCSV(file)
		case "shapes.txt":
			routeShapes, err = parseShapesCSV(file)
		case "trips.txt":
			trips, err = parseTripsCSV(file)
		case "routes.txt":
			routes, err = parseRoutesCSV(file)
		}
		if err != nil {
			return StaticData{}, fmt.Errorf("failed to parse %s: %w", file.Name, err)
		}
	}

	return StaticData{Stops: stops, Shapes: routeShapes, Trips: trips, Routes: routes}, nil
}

// parseStopsCSV reads the CSV file and maps the columns to our Stop struct.
func parseStopsCSV(file *zip.File) (map[string]Stop, error) {
	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)

	// Read the first row to determine column indices (GTFS columns can change order)
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Map header names to their column index
	headerMap := make(map[string]int)
	for i, name := range headers {
		headerMap[name] = i
	}

	// Extract the indices we care about
	idIdx := headerMap["stop_id"]
	nameIdx := headerMap["stop_name"]
	latIdx := headerMap["stop_lat"]
	lonIdx := headerMap["stop_lon"]

	// Parse the rest of the rows
	stops := make(map[string]Stop)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV records: %w", err)
	}

	for _, record := range records {
		lat, _ := strconv.ParseFloat(record[latIdx], 64)
		lon, _ := strconv.ParseFloat(record[lonIdx], 64)

		stop := Stop{
			ID:   record[idIdx],
			Name: record[nameIdx],
			Lat:  lat,
			Lon:  lon,
		}
		stops[stop.ID] = stop
	}

	return stops, nil
}

// parseShapesCSV reads the shapes.txt file, extracts the coordinates,
// and groups them by shape_id in the correct sequential order.
func parseShapesCSV(file *zip.File) (map[string][]ShapePoint, error) {
	// 1. Open the file directly from the zip archive in memory
	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 2. Initialize the CSV reader
	csvReader := csv.NewReader(f)

	// 3. Read the header row to find our column indexes
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	headerMap := make(map[string]int)
	for i, name := range headers {
		headerMap[name] = i
	}

	idIdx := headerMap["shape_id"]
	latIdx := headerMap["shape_pt_lat"]
	lonIdx := headerMap["shape_pt_lon"]
	seqIdx := headerMap["shape_pt_sequence"]

	// 4. Initialize our map to hold the final grouped shapes
	shapes := make(map[string][]ShapePoint)

	// 5. Read all the remaining rows
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read shapes CSV: %w", err)
	}

	// 6. Iterate over every single row
	for _, record := range records {
		shapeID := record[idIdx]

		// Convert strings to floats/ints.
		// In Go, 64 means 64-bit precision (standard for high-accuracy math/GPS).
		lat, _ := strconv.ParseFloat(record[latIdx], 64)
		lon, _ := strconv.ParseFloat(record[lonIdx], 64)
		seq, _ := strconv.Atoi(record[seqIdx]) // Atoi = ASCII to Integer

		point := ShapePoint{
			Lat:      lat,
			Lon:      lon,
			Sequence: seq,
		}

		// Append this point to the slice for this specific shapeID.
		// If the shapeID doesn't exist in the map yet, Go handles creating it automatically.
		shapes[shapeID] = append(shapes[shapeID], point)
	}

	// 7. Sort the points for every shape to ensure a smooth line
	for id, points := range shapes {
		// sort.Slice is a highly optimized sorting algorithm built into Go.
		// It takes the slice, and a custom "less than" function to compare two items (i and j).
		sort.Slice(points, func(i, j int) bool {
			return points[i].Sequence < points[j].Sequence
		})

		// Update the map with the newly sorted list
		shapes[id] = points
	}

	return shapes, nil
}

// parseTripsCSV reads the trips.txt file and creates a mapping of trip_id to its static route and shape information.
// The key is the trip_id, and the value is a Trip struct containing the route_id and shape_id.
func parseTripsCSV(file *zip.File) (map[string]Trip, error) {
	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)

	// Read the header row to determine column indices
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	headerMap := make(map[string]int)
	for i, name := range headers {
		headerMap[name] = i
	}

	idIdx := headerMap["trip_id"]
	routeIdx := headerMap["route_id"]
	shapeIdx := headerMap["shape_id"]

	trips := make(map[string]Trip)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV records: %w", err)
	}

	for _, record := range records {
		trip := Trip{
			ID:      record[idIdx],
			RouteID: record[routeIdx],
			ShapeID: record[shapeIdx],
		}
		trips[trip.ID] = trip
	}

	return trips, nil
}

func parseRoutesCSV(file *zip.File) (map[string]Route, error) {
	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read routes headers: %w", err)
	}

	headerMap := make(map[string]int)
	for i, name := range headers {
		headerMap[name] = i
	}

	routes := make(map[string]Route)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read routes records: %w", err)
	}

	for _, record := range records {
		id := record[headerMap["route_id"]]

		// GTFS has short names (e.g. "Red") and long names (e.g. "Richmond-Daly City").
		// Sometimes one is blank, so we grab both just in case.
		shortName := ""
		if idx, ok := headerMap["route_short_name"]; ok {
			shortName = record[idx]
		}

		longName := ""
		if idx, ok := headerMap["route_long_name"]; ok {
			longName = record[idx]
		}

		color := "FFFFFF" // Default to white if the agency didn't provide a color
		if idx, ok := headerMap["route_color"]; ok && record[idx] != "" {
			color = record[idx]
		}

		routes[id] = Route{
			ID:        id,
			ShortName: shortName,
			LongName:  longName,
			Color:     color,
		}
	}

	return routes, nil
}

package transit

// SystemState is the final JSON payload sent to the frontend.
type SystemState struct {
	Timestamp int64             `json:"timestamp"`
	Vehicles  []VehiclePosition `json:"trains"`
}

// VehiclePosition represents a single train's real-time position and status.
type VehiclePosition struct {
	TripID       string  `json:"trip_id"`
	RouteID      string  `json:"route_id"`
	RouteName    string  `json:"route_name"`
	RouteColor   string  `json:"route_color"`
	ShapeID      string  `json:"shape_id"`
	CurrentLat   float64 `json:"current_lat"`
	CurrentLon   float64 `json:"current_lon"`
	NextStopName string  `json:"next_stop_name"`
	ETA          int64   `json:"eta"` // Unix timestamp of arrival
	DelaySeconds int32   `json:"delay_seconds"`
}

// Route represents the static marketing information for a line (routes.txt)
type Route struct {
	ID        string `json:"id"`
	ShortName string `json:"short_name"`
	LongName  string `json:"long_name"`
	Color     string `json:"color"`
}

// Trip represents the static information about a train trip, including its route and shape.
// This is the "bridge" between the live train and the physical track
type Trip struct {
	ID      string
	RouteID string
	ShapeID string
}

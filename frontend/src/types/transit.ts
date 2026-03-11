// Shapes
export interface ShapePoint {
	Lat: number;
	Lon: number;
	Sequence: number;
}

export interface Route {
	id: string;
	short_name: string;
	long_name: string;
	color: string;
}

export interface Shape {
	ID: string;
	Points: ShapePoint[];
	PrimaryRoute: Route;
}

export type ShapesPayload = Record<string, Shape>;

// Stops
export interface Stop {
	ID: string;
	Name: string;
	Lat: number;
	Lon: number;
}

export type StopsPayload = Record<string, Stop>;

// Trains
export interface TrainState {
	trip_id: string;
	route_id: string;
	shape_id: string;
	current_lat: number;
	current_lon: number;
	next_stop_name: string;
	delay_seconds: number;
	route_color: string;
	route_name: string;
}

export interface NetworkState {
	timestamp: number;
	trains: TrainState[];
}

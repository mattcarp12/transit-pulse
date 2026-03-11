// data/buildGeoJSON.ts

import type {
	NetworkState,
	ShapesPayload,
	StopsPayload,
} from "../types/transit";

export function buildShapesGeoJSON(
	data: ShapesPayload,
): GeoJSON.FeatureCollection {
	return {
		type: "FeatureCollection",
		features: Object.entries(data).map(([shapeId, shape]) => ({
			type: "Feature",
			geometry: {
				type: "LineString",
				coordinates: shape.Points.map((p) => [p.Lon, p.Lat]),
			},
			properties: {
				shape_id: shapeId,
				route_id: shape.PrimaryRoute.id,
				route_name:
					shape.PrimaryRoute.long_name || shape.PrimaryRoute.short_name,
				route_color: shape.PrimaryRoute.color,
			},
		})),
	};
}

export function buildStopsGeoJSON(
	data: StopsPayload,
): GeoJSON.FeatureCollection {
	return {
		type: "FeatureCollection",
		features: Object.values(data).map((stop) => ({
			type: "Feature",
			geometry: {
				type: "Point",
				coordinates: [stop.Lon, stop.Lat],
			},
			properties: {
				stop_id: stop.ID,
				stop_name: stop.Name,
			},
		})),
	};
}

export function buildTrainsGeoJSON(
	data: NetworkState,
): GeoJSON.FeatureCollection {
	return {
		type: "FeatureCollection",
		features: data.trains.map((train) => ({
			type: "Feature",
			geometry: {
				type: "Point",
				coordinates: [train.current_lon, train.current_lat],
			},
			properties: {
				id: train.trip_id,
				route: train.route_id,
				delay: train.delay_seconds,
				next_stop: train.next_stop_name,
				shape_id: train.shape_id,
				route_color: train.route_color || "#888888",
				route_name: train.route_name,
			},
		})),
	};
}

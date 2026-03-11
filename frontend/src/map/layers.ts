// map/layers.ts

export function addShapeLayers(map: maplibregl.Map) {
	map.addLayer({
		id: "shapes-layer",
		type: "line",
		source: "shapes",
		paint: {
			"line-color": ["get", "route_color"],
			"line-width": 2,
			"line-opacity": 0.6,
		},
	});

	map.addLayer({
		id: "shapes-highlight-layer",
		type: "line",
		source: "shapes",
		paint: {
			"line-color": ["get", "route_color"],
			"line-width": 10,
			"line-opacity": 1.0,
			"line-blur": 2,
		},
		filter: ["==", "shape_id", ""],
	});
}

export function addStopLayers(map: maplibregl.Map) {
	map.addLayer({
		id: "stops-layer",
		type: "circle",
		source: "stops",
		paint: {
			"circle-radius": 4,
			"circle-color": "#ff0000",
			"circle-stroke-width": 1,
			"circle-stroke-color": "#fff",
		},
	});

	map.addLayer({
		id: "stops-labels",
		type: "symbol",
		source: "stops",
		layout: {
			"text-field": ["get", "stop_name"],
			"text-size": 10,
			"text-offset": [0, 1.2],
			"text-anchor": "top",
		},
		paint: {
			"text-color": "#ffffff",
			"text-halo-color": "#000000",
			"text-halo-width": 1,
		},
	});
}

export function addTrainLayer(map: maplibregl.Map) {
	map.addLayer({
		id: "trains-layer",
		type: "circle",
		source: "trains-source",
		paint: {
			"circle-radius": 6,
			"circle-color": ["get", "route_color"],
			"circle-stroke-width": 2,
			"circle-stroke-color": "#000",
		},
	});
}

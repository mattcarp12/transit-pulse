// map/initMap.ts
import maplibregl from "maplibre-gl";

export function initMap(container: HTMLDivElement): maplibregl.Map {
	return new maplibregl.Map({
		container,
		style: "https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json",
		center: [-122.27, 37.8],
		zoom: 9.5,
		pitch: 45,
	});
}

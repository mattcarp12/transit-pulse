import maplibregl from "maplibre-gl";
import { useEffect, useRef } from "react";
import "maplibre-gl/dist/maplibre-gl.css";

import {
  buildShapesGeoJSON,
  buildStopsGeoJSON,
  buildTrainsGeoJSON,
} from "./data/buildGeoJSON";

import {
  addShapeLayers,
  addStopLayers,
  addTrainLayer,
} from "./map/layers";

export default function App() {
  const mapContainer = useRef<HTMLDivElement>(null);
  const mapRef = useRef<maplibregl.Map | null>(null);

  useEffect(() => {
    if (!mapContainer.current) return;

    // 1. Initialize the map
    const map = new maplibregl.Map({
      container: mapContainer.current,
      style: "https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json",
      center: [-122.27, 37.8],
      zoom: 9.5,
      pitch: 45,
    });

    mapRef.current = map;

    // Popup reused for train hover
    const popup = new maplibregl.Popup({
      closeButton: false,
      closeOnClick: false,
      offset: 10,
    });

    // 2. When the map loads, add all static sources + layers
    map.on("load", async () => {
      // --- Load static shapes ---
      const shapes = await fetch(import.meta.env.VITE_SHAPES_URL).then((r) =>
        r.json()
      );
      map.addSource("shapes", {
        type: "geojson",
        data: buildShapesGeoJSON(shapes),
      });

      // --- Load static stops ---
      const stops = await fetch(import.meta.env.VITE_STOPS_URL).then((r) =>
        r.json()
      );
      map.addSource("stops", {
        type: "geojson",
        data: buildStopsGeoJSON(stops),
      });

      // --- Create empty trains source ---
      map.addSource("trains-source", {
        type: "geojson",
        data: { type: "FeatureCollection", features: [] },
      });

      // --- Add layers ---
      addShapeLayers(map);
      addStopLayers(map);
      addTrainLayer(map);

      // --- Train hover interactions ---
      map.on("mouseenter", "trains-layer", (e) => {
        map.getCanvas().style.cursor = "pointer";
        if (!e.features?.length) return;

        const feature = e.features[0];
        const shapeId = feature.properties?.shape_id || "";

        // Highlight the matching route
        map.setFilter("shapes-highlight-layer", ["==", "shape_id", shapeId]);

        const coords = (feature.geometry as any).coordinates.slice();
        const { route, next_stop, delay } = feature.properties;

        let delayText =
          '<span style="color: #33ff55; font-weight: bold;">On Time</span>';
        if (delay > 60) {
          const minutes = Math.floor(delay / 60);
          delayText = `<span style="color: #ff3333; font-weight: bold;">${minutes} min delayed</span>`;
        }

        const html = `
          <div style="font-family: sans-serif; padding: 4px; color: #333;">
            <div style="font-size: 14px; font-weight: bold; margin-bottom: 4px;">Route: ${route}</div>
            <div style="font-size: 12px; margin-bottom: 2px;">Next: ${next_stop}</div>
            <div style="font-size: 12px;">Status: ${delayText}</div>
          </div>
        `;

        popup.setLngLat(coords).setHTML(html).addTo(map);
      });

      map.on("mouseleave", "trains-layer", () => {
        map.getCanvas().style.cursor = "";
        popup.remove();
        map.setFilter("shapes-highlight-layer", ["==", "shape_id", ""]);
      });

      // --- Live train polling (the key part that must stay here) ---
      const fetchTransitData = async () => {
        try {
          const response = await fetch(import.meta.env.VITE_LIVE_DATA_URL);
          const data = await response.json();
          const geojson = buildTrainsGeoJSON(data);

          const source = map.getSource(
            "trains-source"
          ) as maplibregl.GeoJSONSource;

          if (source) {
            source.setData(geojson);
          }
        } catch (err) {
          console.error("Failed to fetch transit data:", err);
        }
      };

      // Initial fetch
      await fetchTransitData();

      // Poll every 5 seconds
      const intervalId = setInterval(fetchTransitData, 5000);

      // Clean up on map removal
      map.once("remove", () => clearInterval(intervalId));
    });

    return () => map.remove();
  }, []);

  return (
    <div
      ref={mapContainer}
      style={{
        width: "100%",
        height: "100%",
        position: "absolute",
        top: 0,
        left: 0,
      }}
    />
  );
}
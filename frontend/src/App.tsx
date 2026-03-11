import maplibregl from "maplibre-gl";
import { useEffect, useRef } from "react";
import "maplibre-gl/dist/maplibre-gl.css";

// Define the shape of the data coming from our Go backend
interface TrainState {
  trip_id: string;
  route_id: string;
  shape_id: string; 
  current_lat: number;
  current_lon: number;
  next_stop_name: string;
  delay_seconds: number;
}

interface NetworkState {
  timestamp: number;
  trains: TrainState[];
}

interface ShapePoint {
  Lat: number;
  Lon: number;
  Sequence: number;
}

type ShapesPayload = Record<string, ShapePoint[]>;

function buildShapesGeoJSON(data: ShapesPayload): GeoJSON.FeatureCollection {
  const features: GeoJSON.Feature[] = [];

  // Loop through every shape_id in the dictionary
  for (const [shapeId, points] of Object.entries(data)) {
    // A LineString requires an array of [longitude, latitude] arrays
    const coordinates = points.map(p => [p.Lon, p.Lat]);

    features.push({
      type: 'Feature',
      geometry: {
        type: 'LineString', // Tells MapLibre to draw a line, not a dot
        coordinates: coordinates,
      },
      properties: {
        shape_id: shapeId, // We attach the ID so we can target it for the glow effect later
      },
    });
  }

  return {
    type: 'FeatureCollection',
    features: features,
  };
}

export default function App() {
  const mapContainer = useRef<HTMLDivElement>(null);
  const mapInstance = useRef<maplibregl.Map | null>(null);

  useEffect(() => {
    // The function that talks to our Go backend
    const fetchTransitData = async () => {
      try {
        const response = await fetch(import.meta.env.VITE_LIVE_DATA_URL);
        const data: NetworkState = await response.json();

        // Convert our flattened JSON into the GeoJSON format MapLibre requires
        const geoJsonFeatures: GeoJSON.Feature[] = data.trains.map((train) => ({
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
            shape_id: train.shape_id, // Pass the shape_id so we can link to the route shapes for highlighting
          },
        }));

        const geoJsonFeatureCollection: GeoJSON.FeatureCollection = {
          type: "FeatureCollection",
          features: geoJsonFeatures,
        };

        // Update the map source with the new coordinates
        const source = mapInstance.current?.getSource(
          "trains-source",
        ) as maplibregl.GeoJSONSource;
        if (source) {
          source.setData(geoJsonFeatureCollection);
        }
      } catch (error) {
        console.error("Failed to fetch transit data:", error);
      }
    };

    if (!mapContainer.current) return;

    mapInstance.current = new maplibregl.Map({
      container: mapContainer.current,
      style: "https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json",
      center: [-122.27, 37.8],
      zoom: 9.5,
      pitch: 45,
    });

    const map = mapInstance.current;

    // Create a reusable popup instance, but don't add it to the map yet
    const popup = new maplibregl.Popup({
      closeButton: false,
      closeOnClick: false,
      offset: 10, // Push it slightly above the dot
    });

    map.on("load", async () => {

      // --- 1. FETCH AND ADD THE STATIC ROUTE SHAPES ---
      try {
        const response = await fetch(import.meta.env.VITE_SHAPES_URL);
        const shapesData: ShapesPayload = await response.json();
        const shapesGeoJSON = buildShapesGeoJSON(shapesData);

        map.addSource("shapes", {
          type: "geojson",
          data: shapesGeoJSON,
        });

        // Layer A: The Dim Base Tracks
        map.addLayer({
          id: "shapes-layer",
          type: "line",
          source: "shapes",
          layout: {
            "line-join": "round",
            "line-cap": "round",
          },
          paint: {
            "line-color": "#f12424",
            "line-width": 2,
            "line-opacity": 0.5,
          },
        });

        // Layer B: The Bright Highlight Tracks (on top)
        map.addLayer({
          id: "shapes-highlight-layer",
          type: "line",
          source: "shapes",
          layout: {
            "line-join": "round",
            "line-cap": "round",
          },
          paint: {
            "line-color": "#00ffff",
            "line-width": 4,
            "line-opacity": 0.8,
            "line-blur": 2,
          },
          // CRITICAL: Hide all lines by default; we'll reveal them dynamically based on the train's route
          filter: ["==", "shape_id", ""],
        });
      } catch (error) {
        console.error("Failed to load route shapes:", error);
      }

      // --- 2. SET UP THE TRAIN SOURCE AND LAYER ---
      map.addSource("trains-source", {
        type: "geojson",
        data: { type: "FeatureCollection", features: [] },
      });

      map.addLayer({
        id: "trains-layer",
        type: "circle",
        source: "trains-source",
        paint: {
          "circle-radius": 6,
          "circle-color": [
            "case",
            [">", ["get", "delay"], 60],
            "#ff3333", // Red if delayed more than 60 seconds
            "#33ff55", // Neon green otherwise
          ],
          "circle-stroke-width": 2,
          "circle-stroke-color": "#000",
        },
      });

      // --- 3. DYNAMIC ROUTE HIGHLIGHTING ---

      // When the mouse enters a train dot, show the popup
      map.on("mouseenter", "trains-layer", (e) => {
        // Change the cursor to a pointer finger
        map.getCanvas().style.cursor = "pointer";

        if (!e.features || e.features.length === 0) return;

        // Extract the properties we packed into the GeoJSON earlier
        const feature = e.features[0];

        // Read the feature's shapeId property to know which route to highlight
        const shapeId = feature.properties?.shape_id || "";

        // THE MAGIC: Update the filter on the highlight layer to only show the line that matches this shapeId
        map.setFilter("shapes-highlight-layer", ["==", "shape_id", shapeId]);


        const coordinates = (
          feature.geometry as GeoJSON.Point
        ).coordinates.slice() as [number, number];
        const { route, next_stop, delay } = feature.properties as { route: string; next_stop: string; delay: number };

        // Format the delay text
        let delayText =
          '<span style="color: #33ff55; font-weight: bold;">On Time</span>';
        if (delay > 60) {
          const minutes = Math.floor(delay / 60);
          delayText = `<span style="color: #ff3333; font-weight: bold;">${minutes} min delayed</span>`;
        }

        // Build the HTML for the tooltip
        const htmlContent = `
            <div style="font-family: sans-serif; padding: 4px; color: #333;">
              <div style="font-size: 14px; font-weight: bold; margin-bottom: 4px;">Route: ${route}</div>
              <div style="font-size: 12px; margin-bottom: 2px;">Next: ${next_stop}</div>
              <div style="font-size: 12px;">Status: ${delayText}</div>
            </div>
          `;

        // Attach the popup to the map at the exact coordinate
        popup.setLngLat(coordinates).setHTML(htmlContent).addTo(map);
      });

      // When the mouse leaves the dot, remove the popup
      map.on("mouseleave", "trains-layer", () => {
        map.getCanvas().style.cursor = "";
        popup.remove();

        // Reset the highlight layer filter to hide all lines again
        map.setFilter("shapes-highlight-layer", ["==", "shape_id", ""]);
      });

      // -----------------------------------

      fetchTransitData();
      const intervalId = setInterval(fetchTransitData, 5000);
      return () => clearInterval(intervalId);
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

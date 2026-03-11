package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mattcarp12/transit-pulse/backend/internal/transit"
)

// App holds our dependencies so our HTTP handlers can access them
type App struct {
	client     *transit.Client
	staticData transit.StaticData
}

func main() {
	client := transit.NewClient()

	// 1. Load the static data ONCE on startup
	fmt.Println("Downloading static GTFS schedule...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	staticData, err := client.FetchStaticData(ctx)
	cancel()

	if err != nil {
		log.Fatalf("Failed to fetch static stops: %v", err)
	}
	fmt.Printf("Loaded %d stops into memory.\n", len(staticData.Stops))

	app := &App{
		client:     client,
		staticData: staticData,
	}

	// 2. Set up the API Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/trains", app.trainsHandler)
	mux.HandleFunc("/shapes", app.shapesHandler)

	// 3. Start the server
	port := ":8080"
	fmt.Printf("Transit Pulse API running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}

// trainsHandler responds to GET /trains with our JSON payload
func (app *App) trainsHandler(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers so our local React app can hit this endpoint
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Fetch the live data (with a 5-second timeout)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	feed, err := app.client.FetchTripUpdates(ctx)
	if err != nil {
		http.Error(w, `{"error": "failed to fetch live data"}`, http.StatusInternalServerError)
		return
	}

	// Build the clean JSON payload
	networkState := transit.BuildNetworkState(feed, app.staticData.Stops, app.staticData.Trips)

	// Send it to the client
	if err := json.NewEncoder(w).Encode(networkState); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

func (app *App) shapesHandler(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers so our local React app can hit this endpoint
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Send the route shapes to the client
	if err := json.NewEncoder(w).Encode(app.staticData.Shapes); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

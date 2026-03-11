package transit

import (
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
)

// BuildNetworkState processes the raw GTFS feed and static stops into our clean JSON model.
func BuildNetworkState(feed *gtfs.FeedMessage, staticData StaticData) SystemState {
	state := SystemState{
		Timestamp: int64(feed.Header.GetTimestamp()),
		Vehicles:  make([]VehiclePosition, 0),
	}

	currentTime := time.Now().Unix()

	for _, entity := range feed.GetEntity() {
		tripUpdate := entity.GetTripUpdate()
		if tripUpdate == nil {
			continue // Skip if it's not a trip update (e.g., it's a service alert)
		}

		stopUpdates := tripUpdate.GetStopTimeUpdate()
		if len(stopUpdates) == 0 {
			continue
		}

		// The first item in the list is usually the immediate next stop
		nextStopUpdate := stopUpdates[0]
		nextStopID := nextStopUpdate.GetStopId()

		// Look up the physical stop in our static dictionary
		physicalStop, exists := staticData.Stops[nextStopID]
		if !exists {
			continue // If we can't find the stop, we can't map the train
		}

		arrival := nextStopUpdate.GetArrival()
		if arrival == nil {
			continue
		}

		tripID := tripUpdate.GetTrip().GetTripId()
		shapeID := "" // Placeholder
		routeName := ""
		routeColor := ""

		if trip, ok := staticData.Trips[tripID]; ok {
			shapeID = trip.ShapeID

			if route, ok := staticData.Routes[trip.RouteID]; ok {
				if route.LongName != "" {
					routeName = route.LongName
				} else {
					routeName = route.ShortName
				}
				routeColor = route.Color
			}
		}

		// --- GEOSPATIAL INTERPOLATION PLACEHOLDER ---
		// For this iteration, we will "snap" the train's location to the next stop
		// if it is very close, or slightly offset it based on time.
		// We will build the full Point A -> Point B math in the next step when we
		// parse the previous stops. For now, we seed it with the destination coordinates.
		calculatedLat := physicalStop.Lat
		calculatedLon := physicalStop.Lon

		// If the train is more than 60 seconds away, simulate it being "en route"
		// by slightly offsetting the coordinate for visual testing on the map.
		if arrival.GetTime()-currentTime > 60 {
			calculatedLat -= 0.005 // Arbitrary offset for testing
			calculatedLon -= 0.005
		}

		train := VehiclePosition{
			TripID:       tripUpdate.GetTrip().GetTripId(),
			RouteID:      tripUpdate.GetTrip().GetRouteId(),
			ShapeID:      shapeID,
			CurrentLat:   calculatedLat,
			CurrentLon:   calculatedLon,
			NextStopName: physicalStop.Name,
			ETA:          arrival.GetTime(),
			DelaySeconds: arrival.GetDelay(),
			RouteName:    routeName,
			RouteColor:   routeColor,
		}

		state.Vehicles = append(state.Vehicles, train)
	}

	return state
}

package cmsv6

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
)

// drivingLastRouteCollection is the MongoDB collection
// WatchDrivingLastRoute listens to.
const drivingLastRouteCollection = "drivinglastroute"

// drivingLastRouteWatchRetryDelay is how long WatchDrivingLastRoute waits
// before reopening the change stream after it errors out.
const drivingLastRouteWatchRetryDelay = 5 * time.Second

// drivingLastRouteStatusEvent is the subset of drivinglastroute documents
// WatchDrivingLastRoute cares about, e.g.:
//
//	{
//	  devIDNO: '91079C3',
//	  date: '2026-09-08',
//	  event: 'status',
//	  onoffline: NumberInt('0')
//	}
//
// devIDNO is matched against model.Vehicle.ExternalID; onoffline is 1 for
// online, anything else (observed: 0) for offline.
type drivingLastRouteStatusEvent struct {
	DevIDNO   string `bson:"devIDNO"`
	Event     string `bson:"event"`
	OnOffline int32  `bson:"onoffline"`
}

// WatchDrivingLastRoute opens a change stream on the drivinglastroute
// collection and, for every "status" event, updates the matching vehicle's
// status via vehicles.SetStatusByExternalID — devIDNO is matched against
// model.Vehicle.ExternalID, and onoffline maps to
// model.VehicleStatusOnline/model.VehicleStatusOffline.
//
// It blocks until ctx is canceled, reopening the change stream (after
// drivingLastRouteWatchRetryDelay) whenever it errors out, so callers
// should run it in its own goroutine. A single failed status update (e.g.
// no vehicle registered with that external ID yet) is logged and does not
// stop the watch.
func (m *MongoClient) WatchDrivingLastRoute(ctx context.Context, vehicles port.VehicleService) error {
	collection := m.database.Collection(drivingLastRouteCollection)

	for {
		err := m.watchDrivingLastRouteOnce(ctx, collection, vehicles)
		if ctx.Err() != nil {
			return ctx.Err()
		}

		m.log.Errorw("drivinglastroute change stream failed, retrying",
			"error", err, "retryIn", drivingLastRouteWatchRetryDelay)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(drivingLastRouteWatchRetryDelay):
		}
	}
}

// drivingLastRoutePipeline restricts the change stream to insert events —
// drivinglastroute never updates a document in place, every new reading is
// its own inserted document — so there's no need to watch for
// updates/replaces/deletes, and the full document is already present on
// every matched event without requesting an update lookup.
var drivingLastRoutePipeline = mongo.Pipeline{
	bson.D{{Key: "$match", Value: bson.D{{Key: "operationType", Value: "insert"}}}},
}

func (m *MongoClient) watchDrivingLastRouteOnce(ctx context.Context, collection *mongo.Collection, vehicles port.VehicleService) error {
	stream, err := collection.Watch(ctx, drivingLastRoutePipeline)
	if err != nil {
		return fmt.Errorf("cmsv6: failed to open drivinglastroute change stream: %w", err)
	}
	defer stream.Close(ctx)

	m.log.Info("listening for drivinglastroute changes")

	for stream.Next(ctx) {
		var change struct {
			FullDocument drivingLastRouteStatusEvent `bson:"fullDocument"`
		}
		if err := stream.Decode(&change); err != nil {
			m.log.Errorw("failed to decode drivinglastroute change event", "error", err)
			continue
		}

		event := change.FullDocument
		if event.Event != "status" || event.DevIDNO == "" {
			continue
		}

		status := model.VehicleStatusOffline
		if event.OnOffline == 1 {
			status = model.VehicleStatusOnline
		}

		if err := vehicles.SetStatusByExternalID(ctx, event.DevIDNO, status); err != nil {
			m.log.Errorw("failed to set vehicle status from drivinglastroute event",
				"externalId", event.DevIDNO, "status", status, "error", err)
		}
	}

	return stream.Err()
}

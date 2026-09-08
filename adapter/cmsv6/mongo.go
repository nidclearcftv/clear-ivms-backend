package cmsv6

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"

	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type MongoOptions struct {
	Logger *zap.SugaredLogger

	URI      string `validate:"required"`
	Database string `validate:"required"`
}

// MongoClient wraps a mongo.Client scoped to a single database, configured
// via MongoOptions.
type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
	opts     MongoOptions
	log      *zap.SugaredLogger
}

func NewMongoClient(ctx context.Context, opts MongoOptions) (*MongoClient, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	log := opts.Logger
	if log == nil {
		log = zap.NewNop().Sugar()
	}
	log = log.With("adapter", "cmsv6", "component", "mongo")

	client, err := mongo.Connect(options.Client().ApplyURI(opts.URI))
	if err != nil {
		return nil, fmt.Errorf("cmsv6: failed to create mongo client: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("cmsv6: failed to ping mongo: %w", err)
	}

	log.Infow("connected to mongo", "database", opts.Database)

	return &MongoClient{
		client:   client,
		database: client.Database(opts.Database),
		opts:     opts,
		log:      log,
	}, nil
}

// Database returns the mongo.Database this client is scoped to (see
// MongoOptions.Database), for callers to run queries against.
func (m *MongoClient) Database() *mongo.Database {
	return m.database
}

// Close disconnects the underlying mongo.Client.
func (m *MongoClient) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

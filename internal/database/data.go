package database

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Data struct {
	Conn     *mongo.Client
	Database *mongo.Database
	Coll     *Collections
}

type Collections struct {
	Config *mongo.Collection
}

// NewDatabase connects to MongoDB and checks it can be reached, giving up
// early if ctx is canceled (the bot is shutting down).
func NewDatabase(ctx context.Context, uri string) (*Data, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	name := strings.TrimPrefix(parsed.Path, "/")
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("database name is missing in the URI")
	}

	conn, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Connect doesn't contact the server, so check it can be reached now
	// rather than on the first query.
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := conn.Ping(pingCtx, nil); err != nil {
		_ = conn.Disconnect(context.Background())
		return nil, fmt.Errorf("can't reach MongoDB: %w", err)
	}

	database := conn.Database(name)

	coll := &Collections{
		Config: database.Collection("config"),
	}

	return &Data{
		Conn:     conn,
		Database: database,
		Coll:     coll,
	}, nil
}

package sampledb

import (
	"context"

	"github.com/gookit/event"
	ev "github.com/gotrackery/gotrackery/internal/event"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

var _ event.Listener = (*DB)(nil)

type DB struct {
	logger *zerolog.Logger
	db     *pgxpool.Pool
}

func NewDB(l *zerolog.Logger, db *pgxpool.Pool) (*DB, error) {
	if db == nil {
		panic("missing *pgxpool.Pool, parameter must not be nil")
	}

	err := db.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return &DB{logger: l, db: db}, nil
}

func (d *DB) Handle(e event.Event) (err error) {
	pos := e.(*ev.GenericEvent).GetPosition()
	tr, err := CreateFromGeneric(pos)
	if err != nil {
		return err
	}
	err = d.insert(context.Background(), tr)
	if err != nil {
		return
	}

	return nil
}

func (d *DB) insert(ctx context.Context, pos *Position) error {
	insCmd := `
		SELECT public.insert_position (@protocol, @deviceID, @serverTime, @deviceTime,
		    @fixTime, @valid, @latitude, @longitude,
		    @altitude, @speed, @course, @address,
		    @attributes, @accuracy, @network)`
	args := pgx.NamedArgs{
		"protocol":   pos.Protocol,
		"deviceID":   pos.DeviceID,
		"serverTime": pos.ServerTime,
		"deviceTime": pos.DeviceTime,
		"fixTime":    pos.FixTime,
		"valid":      pos.Valid,
		"latitude":   pos.Latitude,
		"longitude":  pos.Longitude,
		"altitude":   pos.Altitude,
		"speed":      pos.Speed,
		"course":     pos.Course,
		"address":    pos.Address,
		"attributes": pos.Attributes,
		"accuracy":   pos.Accuracy,
		"network":    pos.Network,
	}
	var tcDevID int64
	err := d.db.QueryRow(ctx, insCmd, args).Scan(&tcDevID)
	if err != nil {
		return err
	}
	return nil
}

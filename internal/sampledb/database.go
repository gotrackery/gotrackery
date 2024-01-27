package sampledb

import (
	"context"
	"fmt"
	"strings"

	"github.com/gookit/event"
	ev "github.com/gotrackery/gotrackery/internal/event"
	
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	_ event.Listener   = (*DB)(nil)
	_ event.Subscriber = (*DB)(nil)
)

type DB struct {
	db *pgxpool.Pool
}

const (
	SELF_NAME = "postgres"
)

func (d *DB) String() string {
	return SELF_NAME
}

func (d *DB) close() {
	d.db.Close()
}

func NewDB(db *pgxpool.Pool) (*DB, error) {
	if db == nil {
		panic("missing *pgxpool.Pool, parameter must not be nil")
	}

	err := db.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

func (d *DB) SubscribedEvents() map[string]any {
	return map[string]any{
		fmt.Sprintf("%s.%s", ev.PositionRecived, SELF_NAME): d,
		fmt.Sprintf("%s.%s", ev.CloseConnection, SELF_NAME): d,
	}
}

func (d *DB) Handle(e event.Event) (err error) {
	eve, ok := e.(*ev.GenericEvent)
	if !ok || eve == nil {
		return fmt.Errorf("GenericEvent not transferred")
	}
	name, ok := strings.CutSuffix(eve.Name(), "."+SELF_NAME)
	if !ok {
		return fmt.Errorf("event not found for listner: %s", SELF_NAME)
	}
	switch name {
	case string(ev.PositionRecived):
		pos := eve.Position()
		if pos == nil {
			return fmt.Errorf("position not specified")
		}
		tr, err := CreateFromCommon(*pos)
		if err != nil {
			return fmt.Errorf("create position from common: %w", err)
		}
		err = d.insert(context.Background(), tr)
		if err != nil {
			return fmt.Errorf("insert position: %w", err)
		}

	case string(ev.CloseConnection):
		d.close()
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

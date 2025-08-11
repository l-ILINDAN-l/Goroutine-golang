package postgresrepo

import (
	"calendar/internal/config"
	"calendar/internal/domain/event"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	// The blank import is required to register the PostgreSQL driver.
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"time"
)

// PostgresRepo implements the EventRepository interface for PostgreSQL
type PostgresRepo struct {
	eventDB *sql.DB
	logger  *logrus.Logger
}

// New creates a new instance of the repository
func New(ctx context.Context, cfg *config.PostgresConfig, logger *logrus.Logger) (*PostgresRepo, error) {
	eventDB, err := connectToDB(ctx, cfg.Events)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to events database: %w", err)
	}

	return &PostgresRepo{
		eventDB: eventDB,
		logger:  logger,
	}, nil
}

func connectToDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}

// Create creates a new event in the database within a transaction
func (r *PostgresRepo) Create(ctx context.Context, event *event.Event) error {
	tx, err := r.eventDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			r.logger.WithFields(logrus.Fields{
				"event_id": event.ID,
			}).Errorf("failed to rollback transaction: %v", err)
		}
	}(tx)

	sqlQuery := `INSERT INTO events(id, user_id, date, event_text) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, sqlQuery, event.ID, event.UserID, event.Date.Format("2006-01-02"), event.EventText)
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"event_id": event.ID,
			"user_id":  event.UserID,
		}).Errorf("CRITICAL: failed to insert event: %v", err)
		return err
	}
	err = tx.Commit()
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"event_id": event.ID,
			"user_id":  event.UserID,
		}).Errorf("CRITICAL: failed to commit transaction: %v", err)
		return err
	}

	return nil
}

// Update updates an existing event
func (r *PostgresRepo) Update(ctx context.Context, event *event.Event) error {
	tx, err := r.eventDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			r.logger.WithFields(logrus.Fields{
				"event_id": event.ID,
			}).Errorf("failed to rollback transaction: %v", err)
		}
	}(tx)

	sqlQuery := `UPDATE events SET date = $2, event_text = $3 WHERE id = $1`
	_, err = tx.ExecContext(ctx, sqlQuery, event.ID, event.Date, event.EventText)
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"event_id": event.ID,
		}).Errorf("CRITICAL: failed to update event: %v", err)
		return err
	}
	err = tx.Commit()
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"event_id": event.ID,
		}).Errorf("CRITICAL: failed to commit transaction: %v", err)
		return err
	}

	return nil
}

// Delete removes an event
func (r *PostgresRepo) Delete(ctx context.Context, eventID uuid.UUID) error {
	tx, err := r.eventDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			r.logger.WithFields(logrus.Fields{
				"event_id": eventID,
			}).Errorf("failed to rollback transaction: %v", err)
		}
	}(tx)

	sqlQuery := `DELETE FROM events WHERE id = $1`
	_, err = tx.ExecContext(ctx, sqlQuery, eventID)
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"event_id": eventID,
		}).Errorf("CRITICAL: failed to delete event: %v", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"event_id": eventID,
		}).Errorf("CRITICAL: failed to commit transaction: %v", err)
		return err
	}

	return nil
}

// FindByID finds a single event by its unique ID
func (r *PostgresRepo) FindByID(ctx context.Context, eventID uuid.UUID) (*event.Event, error) {
	sqlQuery := `SELECT id, user_id, date, event_text FROM events WHERE id = $1`
	row := r.eventDB.QueryRowContext(ctx, sqlQuery, eventID)

	var ev event.Event
	err := row.Scan(&ev.ID, &ev.UserID, &ev.Date, &ev.EventText)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
	}

	return &ev, nil
}

// FindByDateRange finds all events for a user within a given date range
func (r *PostgresRepo) FindByDateRange(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]*event.Event, error) {
	sqlQuery := `SELECT id, user_id, date, event_text FROM events WHERE user_id = $1 AND date >= $2 AND date <= $3 ORDER BY date`
	rows, err := r.eventDB.QueryContext(ctx, sqlQuery, userID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			r.logger.WithFields(logrus.Fields{
				"user_id": userID,
			}).Errorf("CRITICAL failed to close rows: %v", err)
		}
	}(rows)

	var events []*event.Event
	for rows.Next() {
		var ev event.Event
		if err := rows.Scan(&ev.ID, &ev.UserID, &ev.Date, &ev.EventText); err != nil {
			return nil, err
		}
		events = append(events, &ev)
	}

	return events, rows.Err()
}

// GetLatest retrieves the most recent events up to a given limit
func (r *PostgresRepo) GetLatest(ctx context.Context, limit uint) ([]*event.Event, error) {
	sqlQuery := `SELECT id, user_id, date, event_text FROM events ORDER BY date DESC LIMIT $1`
	rows, err := r.eventDB.QueryContext(ctx, sqlQuery, limit)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			r.logger.Errorf("failed to close rows: %v", err)
		}
	}(rows)

	var events []*event.Event
	for rows.Next() {
		var ev event.Event
		if err := rows.Scan(&ev.ID, &ev.UserID, &ev.Date, &ev.EventText); err != nil {
			return nil, err
		}
		events = append(events, &ev)
	}

	return events, rows.Err()
}

// Close closes the database connection.
func (r *PostgresRepo) Close() error {
	return r.eventDB.Close()
}

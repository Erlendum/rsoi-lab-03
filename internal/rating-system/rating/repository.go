package rating

import (
	"context"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

const (
	defaultTimeout = 5 * time.Second
)

type repository struct {
	conn *sqlx.DB
}

func NewRepository(conn *sqlx.DB) *repository {
	return &repository{conn: conn}
}

func (r *repository) CreateRatingRecord(ctx context.Context, record *ratingRecord) (int, error) {
	insertRatingRecordQuery := `
	WITH inserted AS (
		INSERT INTO rating
			(username)
				VALUES ($1)
			ON CONFLICT (username) DO NOTHING
			RETURNING id
	)
	SELECT id FROM inserted;`

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var id int
	err := r.conn.QueryRowContext(ctx, insertRatingRecordQuery, *record.UserName).Scan(&id)
	if err != nil {
		return 0, errors.Wrap(err, "failed to execute query")
	}

	return id, nil
}

func (r *repository) createUpdateBuilderForRecord(username string, record ratingRecord) (sq.UpdateBuilder, bool) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	isEmpty := true
	updateBuilder := psql.Update("rating").Where(sq.Eq{"username": username})
	if record.UserName != nil {
		updateBuilder = updateBuilder.Set("username", record.UserName)
		isEmpty = false
	}
	if record.Stars != nil {
		updateBuilder = updateBuilder.Set("stars", record.Stars)
		isEmpty = false
	}

	return updateBuilder, isEmpty
}

func (r *repository) UpdateRatingRecord(ctx context.Context, userName string, record *ratingRecord) error {
	builder, isEmpty := r.createUpdateBuilderForRecord(userName, *record)
	if isEmpty {
		return nil
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	_, err = r.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	return nil
}

func (r *repository) GetRatingRecord(ctx context.Context, username string) (ratingRecord, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Select("id", "username", "stars").From("rating").Where(sq.Eq{"username": username})

	query, args, err := builder.ToSql()
	if err != nil {
		return ratingRecord{}, errors.Wrap(err, "failed to build query")
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	res := ratingRecord{}

	err = r.conn.GetContext(ctx, &res, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ratingRecord{}, errRecordNotFound
		}
		return ratingRecord{}, errors.Wrap(err, "failed to execute query")
	}

	return res, nil
}

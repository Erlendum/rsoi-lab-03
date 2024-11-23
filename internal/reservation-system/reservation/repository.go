package reservation

import (
	"context"
	"database/sql"
	my_time "github.com/Erlendum/rsoi-lab-03/pkg/time"
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

func (r *repository) CreateReservation(ctx context.Context, res *reservation) (int, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	builder := psql.Insert("reservation").Columns("reservation_uid", "username", "book_uid", "library_uid", "status", "start_date", "till_date").
		Values(*res.ReservationUid, *res.UserName, *res.BookUid, *res.LibraryUid, *res.Status, res.StartDate.String(), res.TillDate.String())
	query, args, err := builder.Suffix("RETURNING id").ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "failed to build query")
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var id int
	err = r.conn.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		return 0, errors.Wrap(err, "failed to execute query")
	}

	return id, nil
}

func (r *repository) DeleteReservation(ctx context.Context, uid string) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Delete("reservation").Where(sq.Eq{"reservation_uid": uid})

	query, args, err := builder.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	res, err := r.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}

func (r *repository) UpdateReservationStatus(ctx context.Context, uid string, username string, status string) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Update("reservation").Set("status", status).Where(sq.And{sq.Eq{"reservation_uid": uid}, sq.Eq{"username": username}})

	query, args, err := builder.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	res, err := r.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}

func (r *repository) GetReservation(ctx context.Context, uid string) (reservation, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Select("reservation_uid", "username", "book_uid", "library_uid", "status", "start_date", "till_date").From("reservation").Where(sq.Eq{"reservation_uid": uid})

	query, args, err := builder.ToSql()
	if err != nil {
		return reservation{}, errors.Wrap(err, "failed to build query")
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	res := reservation{}

	var startDate, tillDate string
	err = r.conn.QueryRowContext(ctx, query, args...).Scan(&res.ReservationUid, &res.UserName, &res.BookUid, &res.LibraryUid, &res.Status, &startDate, &tillDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return reservation{}, errNotFound
		}
		return reservation{}, errors.Wrap(err, "failed to execute query")
	}

	res.StartDate, err = my_time.NewDate(startDate)
	if err != nil {
		return reservation{}, errors.Wrap(err, "failed to parse start date")
	}
	res.TillDate, err = my_time.NewDate(tillDate)
	if err != nil {
		return reservation{}, errors.Wrap(err, "failed to parse till date")
	}

	return res, nil
}

func (r *repository) GetReservations(ctx context.Context, username string, status string) ([]reservation, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Select("reservation_uid", "username", "book_uid", "library_uid", "status", "start_date", "till_date").From("reservation").Where(sq.And{sq.Eq{"username": username}, sq.Eq{"status": status}})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	res := make([]reservation, 0)

	rows, err := r.conn.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to perform query %s", query)
	}
	defer rows.Close()

	for rows.Next() {
		var model reservation
		var startDate, tillDate string
		if err = rows.Scan(&model.ReservationUid, &model.UserName, &model.BookUid, &model.LibraryUid, &model.Status, &startDate, &tillDate); err != nil {
			return nil, errors.Wrap(err, "failed to row scan")
		}
		model.StartDate, err = my_time.NewDate(startDate)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse start date")
		}
		model.TillDate, err = my_time.NewDate(tillDate)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse till date")
		}
		res = append(res, model)
	}

	return res, nil
}

package library

import (
	"context"
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

func (r *repository) GetLibraries(ctx context.Context, city string, offset, limit int) ([]library, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	builder := psql.Select("id", "library_uid", "name", "address", "city").
		From("library").
		Where(sq.Eq{"city": city}).Limit(uint64(limit)).Offset(uint64(offset))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	libraries := make([]library, 0)
	err = r.conn.SelectContext(ctx, &libraries, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}

	if len(libraries) == 0 {
		return nil, errors.Wrap(errLibraryNotFound, "library not found")
	}

	return libraries, nil

}

func (r *repository) GetBooksByLibrary(ctx context.Context, libraryUid string, offset, limit int, showAll bool) ([]book, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	builder := psql.Select("b.id", "b.book_uid", "b.name", "b.author", "b.genre", "b.condition", "lb.available_count").
		From("books b").
		Join("library_books lb ON lb.book_id = b.id").
		Join("library l ON lb.library_id = l.id").
		Where(sq.Eq{"l.library_uid": libraryUid}).Limit(uint64(limit)).Offset(uint64(offset))

	if !showAll {
		builder = builder.Where(sq.Gt{"lb.available_count": 0})
	}
	query, args, err := builder.ToSql()

	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	books := make([]book, 0)
	err = r.conn.SelectContext(ctx, &books, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}

	if len(books) == 0 {
		return nil, errors.Wrap(errBookNotFound, "book not found")
	}

	return books, nil
}

func (r *repository) GetBooksAvailableCount(ctx context.Context, libraryUid, bookUid string) (int, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	builder := psql.Select("lb.available_count").
		From("library_books lb").
		Join("books b ON lb.book_id = b.id").
		Join("library l ON lb.library_id = l.id").
		Where(sq.Eq{"l.library_uid": libraryUid, "b.book_uid": bookUid})

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "failed to build query")
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var count int
	err = r.conn.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, errors.Wrap(err, "failed to execute query")
	}

	return count, nil
}

func (r *repository) GetBooksByUids(ctx context.Context, uids []string) ([]book, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	builder := psql.Select("id", "book_uid", "name", "author", "genre", "condition").
		From("books").
		Where(sq.Eq{"book_uid": uids})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	books := make([]book, 0)
	err = r.conn.SelectContext(ctx, &books, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}

	if len(books) == 0 {
		return nil, errors.Wrap(errBookNotFound, "book not found")
	}

	return books, nil
}

func (r *repository) GetLibrariesByUids(ctx context.Context, uids []string) ([]library, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	builder := psql.Select("id", "library_uid", "name", "address", "city").
		From("library").
		Where(sq.Eq{"library_uid": uids})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	libraries := make([]library, 0)
	err = r.conn.SelectContext(ctx, &libraries, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}

	if len(libraries) == 0 {
		return nil, errors.Wrap(errLibraryNotFound, "library not found")
	}

	return libraries, nil
}

func (r *repository) UpdateBooksAvailableCount(ctx context.Context, libraryUid, bookUid string, count int) error {
	query := `
UPDATE library_books
SET available_count = $1
WHERE book_id = (
    SELECT id FROM books WHERE book_uid = $2
)
AND library_id = (
    SELECT id FROM library WHERE library_uid = $3
);
`
	args := []interface{}{count, bookUid, libraryUid}

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
		return errors.Wrap(errRecordNotFound, "no rows affected")
	}

	return nil
}

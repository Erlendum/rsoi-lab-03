package rating

type ratingRecord struct {
	ID       *int    `db:"id"`
	UserName *string `db:"username"`
	Stars    *int    `db:"stars"`
}

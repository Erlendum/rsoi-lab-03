package reservation

import (
	my_time "github.com/Erlendum/rsoi-lab-03/pkg/time"
)

type reservation struct {
	ID             *int          `db:"id"`
	ReservationUid *string       `db:"reservation_uid"`
	UserName       *string       `db:"username"`
	BookUid        *string       `db:"book_uid"`
	LibraryUid     *string       `db:"library_uid"`
	Status         *string       `db:"status"`
	StartDate      *my_time.Date `db:"start_date"`
	TillDate       *my_time.Date `db:"till_date"`
}

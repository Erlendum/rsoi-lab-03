package library

type library struct {
	ID         int    `db:"id"`
	LibraryUid string `db:"library_uid"`
	Name       string `db:"name"`
	Address    string `db:"address"`
	City       string `db:"city"`
}

type book struct {
	ID             int    `db:"id"`
	BookUid        string `db:"book_uid"`
	Name           string `db:"name"`
	Author         string `db:"author"`
	Genre          string `db:"genre"`
	Condition      string `db:"condition"`
	AvailableCount int    `db:"available_count"`
}

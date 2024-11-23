package time

import (
	"strings"
	"time"
)

type Date time.Time

const dateFormat = "2006-01-02"

func NewDate(date string) (*Date, error) {
	index := strings.Index(date, "T")
	if index == -1 {
		index = len(date)
	}
	parsedTime, err := time.Parse(dateFormat, date[:index])
	d := Date(parsedTime)
	return &d, err
}

func (d *Date) UnmarshalJSON(b []byte) error {
	str := string(b)
	// Удаляем кавычки вокруг строки
	str = str[1 : len(str)-1]
	parsedTime, err := time.Parse(dateFormat, str)
	if err != nil {
		return err
	}
	*d = Date(parsedTime)
	return nil
}

func (d *Date) String() string {
	return time.Time(*d).Format(dateFormat)
}

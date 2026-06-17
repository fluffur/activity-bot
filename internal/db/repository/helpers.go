package repository

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func text(string string) pgtype.Text {
	return pgtype.Text{
		String: string,
		Valid:  string != "",
	}
}

func timestamptz(time time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  time,
		Valid: !time.IsZero(),
	}
}

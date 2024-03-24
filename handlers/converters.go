package handlers

import (
	"math/big"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func PgtypeNumeric(num string) pgtype.Numeric {
	var bigNum = new(big.Int)
	bigNum, ok := bigNum.SetString(num, 10)
	if !ok {
		return pgtype.Numeric{Int: big.NewInt(0), Valid: false}
	}
	return pgtype.Numeric{Int: bigNum, Valid: true}
}

func Int32(num string) int32 {
	n, err := strconv.Atoi(num)
	if err != nil {
		return 0
	}
	return int32(n)
}

func PgTypeDate(date string) pgtype.Date {
	// parse from YYYY-MM-DD into time first
	// then convert to pgtype.Date
	var layout = "2006-01-02"
	t, err := time.Parse(layout, date)
	if err != nil {
		return pgtype.Date{Valid: false, Time: time.Time{}}
	}
	return pgtype.Date{Valid: true, Time: t}
}

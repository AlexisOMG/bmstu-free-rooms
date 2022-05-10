package database

import (
	"database/sql"
	"errors"
	"fmt"
	"ics/service"
)

func mapErrors(err error, wrap string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf(wrap, service.ErrorNotFound)
	}

	return fmt.Errorf(wrap, err)
}

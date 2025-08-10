package utils

import (
	"errors"

	"github.com/mattn/go-sqlite3"
)

func IsUniqueConstraintError(err error) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.Code == sqlite3.ErrConstraint && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
	}
	return false
}

// func IsUniqueConstraintError(err error) bool {
// 	var sqlErr sqlite3.Error
// 	if errors.As(err, &sqlErr) {
// 		return sqlErr.Code == sqlite3.ErrConstraint && sqlErr.ExtendedCode == sqlite3.ErrConstraintUnique
// 	}
// 	return false
// }

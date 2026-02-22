package repository

import (
	"database/sql"
	"time"

	"goAuth/internals/helpers"
	"goAuth/internals/types"
)

type PostgresOTPRepository struct {
	db    *sql.DB
	table string
}

func NewPostgresOTPRepository(db *sql.DB, table string) *PostgresOTPRepository {
	if table == "" {
		table = "otp"
	}
	return &PostgresOTPRepository{db: db, table: table}
}

func (r *PostgresOTPRepository) Create(userID, otpHash string, expiry time.Time) error {
	if err := r.DeleteByUserID(userID); err != nil {
		return err
	}
	id := helpers.GenerateID()
	q := `INSERT INTO ` + quoteIdentifier(r.table) + ` (id, user_id, otp, expiry) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(q, id, userID, otpHash, expiry)
	return err
}

func (r *PostgresOTPRepository) CreateTx(tx *sql.Tx, userID, otpHash string, expiry time.Time) error {
	delQ := `DELETE FROM ` + quoteIdentifier(r.table) + ` WHERE user_id = $1`
	if _, err := tx.Exec(delQ, userID); err != nil {
		return err
	}
	id := helpers.GenerateID()
	q := `INSERT INTO ` + quoteIdentifier(r.table) + ` (id, user_id, otp, expiry) VALUES ($1, $2, $3, $4)`
	_, err := tx.Exec(q, id, userID, otpHash, expiry)
	return err
}

func (r *PostgresOTPRepository) GetByUserID(userID string) (otpHash string, expiry time.Time, err error) {
	q := `SELECT otp, expiry FROM ` + quoteIdentifier(r.table) + ` WHERE user_id = $1`
	err = r.db.QueryRow(q, userID).Scan(&otpHash, &expiry)
	if err == sql.ErrNoRows {
		return "", time.Time{}, types.ErrNotFound
	}
	return otpHash, expiry, err
}

func (r *PostgresOTPRepository) DeleteByUserID(userID string) error {
	q := `DELETE FROM ` + quoteIdentifier(r.table) + ` WHERE user_id = $1`
	_, err := r.db.Exec(q, userID)
	return err
}

func (r *PostgresOTPRepository) DeleteExpired() error {
	q := `DELETE FROM ` + quoteIdentifier(r.table) + ` WHERE expiry < now()`
	_, err := r.db.Exec(q)
	return err
}

var _ types.OTPRepository = (*PostgresOTPRepository)(nil)

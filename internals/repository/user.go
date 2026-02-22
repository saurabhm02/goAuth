package repository

import (
	"database/sql"
	"strings"
	"time"

	"goAuth/internals/helpers"
	"goAuth/internals/model"
	"goAuth/internals/types"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type PostgresUserRepository struct {
	db    *sql.DB
	table string
}

func NewPostgresUserRepository(db *sql.DB, table string) *PostgresUserRepository {
	if table == "" {
		table = "users"
	}
	return &PostgresUserRepository{db: db, table: table}
}

func (r *PostgresUserRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *PostgresUserRepository) Create(user *model.User) error {
	if user.ID == "" {
		user.ID = helpers.GenerateID()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	q := `INSERT INTO ` + quoteIdentifier(r.table) + ` (id, email, phone, password_hash, refresh_token_hash, reset_token_hash, reset_token_expiry, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(q,
		user.ID, user.Email, user.Phone, user.PasswordHash,
		user.RefreshTokenHash, user.ResetTokenHash, user.ResetTokenExpiry,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return types.ErrConflict
		}
		return err
	}
	return nil
}

// CreateTx inserts the user within an existing transaction.
func (r *PostgresUserRepository) CreateTx(tx *sql.Tx, user *model.User) error {
	if user.ID == "" {
		user.ID = helpers.GenerateID()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	q := `INSERT INTO ` + quoteIdentifier(r.table) + ` (id, email, phone, password_hash, refresh_token_hash, reset_token_hash, reset_token_expiry, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := tx.Exec(q,
		user.ID, user.Email, user.Phone, user.PasswordHash,
		user.RefreshTokenHash, user.ResetTokenHash, user.ResetTokenExpiry,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return types.ErrConflict
		}
		return err
	}
	return nil
}

func (r *PostgresUserRepository) DeleteByID(id string) error {
	q := `DELETE FROM ` + quoteIdentifier(r.table) + ` WHERE id = $1`
	_, err := r.db.Exec(q, id)
	return err
}

func (r *PostgresUserRepository) GetByID(id string) (*model.User, error) {
	q := `SELECT id, email, phone, password_hash, refresh_token_hash, reset_token_hash, reset_token_expiry, created_at, updated_at
		FROM ` + quoteIdentifier(r.table) + ` WHERE id = $1`
	row := r.db.QueryRow(q, id)
	return r.scanRow(row)
}

func (r *PostgresUserRepository) GetByEmail(email string) (*model.User, error) {
	q := `SELECT id, email, phone, password_hash, refresh_token_hash, reset_token_hash, reset_token_expiry, created_at, updated_at
		FROM ` + quoteIdentifier(r.table) + ` WHERE email = $1`
	row := r.db.QueryRow(q, email)
	return r.scanRow(row)
}

func (r *PostgresUserRepository) GetByEmailOrPhone(emailOrPhone string) (*model.User, error) {
	q := `SELECT id, email, phone, password_hash, refresh_token_hash, reset_token_hash, reset_token_expiry, created_at, updated_at
		FROM ` + quoteIdentifier(r.table) + ` WHERE email = $1 OR phone = $1`
	row := r.db.QueryRow(q, emailOrPhone)
	return r.scanRow(row)
}

func (r *PostgresUserRepository) scanRow(row *sql.Row) (*model.User, error) {
	var u model.User
	var refreshHash, resetHash sql.NullString
	var resetExpiry pq.NullTime
	err := row.Scan(
		&u.ID, &u.Email, &u.Phone, &u.PasswordHash,
		&refreshHash, &resetHash, &resetExpiry,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, types.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if refreshHash.Valid {
		u.RefreshTokenHash = refreshHash.String
	}
	if resetHash.Valid {
		u.ResetTokenHash = resetHash.String
	}
	if resetExpiry.Valid {
		u.ResetTokenExpiry = &resetExpiry.Time
	}
	return &u, nil
}

func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

var _ types.UserRepository = (*PostgresUserRepository)(nil)
var _ types.Transactor = (*PostgresUserRepository)(nil)

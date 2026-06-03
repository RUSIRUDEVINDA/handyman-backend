package customer

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Upsert(ctx context.Context, customer *Customer) error
	GetByUserID(ctx context.Context, userID string) (*Customer, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Upsert(ctx context.Context, customer *Customer) error {
	query := `
		INSERT INTO customers (id, user_id, full_name, phone, address, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id)
		DO UPDATE SET
			full_name = EXCLUDED.full_name,
			phone = EXCLUDED.phone,
			address = EXCLUDED.address,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(ctx, query, customer.ID, customer.UserID, customer.FullName, customer.Phone, customer.Address, customer.CreatedAt, customer.UpdatedAt)
	return err
}

func (r *repository) GetByUserID(ctx context.Context, userID string) (*Customer, error) {
	query := `
		SELECT id, user_id, full_name, phone, address, created_at, updated_at
		FROM customers
		WHERE user_id = $1
	`

	var customer Customer
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&customer.ID,
		&customer.UserID,
		&customer.FullName,
		&customer.Phone,
		&customer.Address,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

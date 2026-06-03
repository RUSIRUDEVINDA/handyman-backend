package booking

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, booking *Booking) error
	GetByID(ctx context.Context, id string) (*Booking, error)
	ListForCustomer(ctx context.Context, customerID string) ([]Booking, error)
	ListForHandyman(ctx context.Context, handymanID string) ([]Booking, error)
	AssignHandyman(ctx context.Context, id string, handymanID string) error
	UpdateStatus(ctx context.Context, id string, status Status) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, b *Booking) error {
	query := `
		INSERT INTO bookings (id, customer_id, handyman_id, service_name, notes, status, amount_cents, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query, b.ID, b.CustomerID, b.HandymanID, b.ServiceName, b.Notes, b.Status, b.AmountCents, b.CreatedAt, b.UpdatedAt)
	return err
}

func (r *repository) GetByID(ctx context.Context, id string) (*Booking, error) {
	query := `
		SELECT id, customer_id, handyman_id, service_name, notes, status, amount_cents, created_at, updated_at
		FROM bookings
		WHERE id = $1
	`
	var b Booking
	err := r.db.QueryRow(ctx, query, id).Scan(&b.ID, &b.CustomerID, &b.HandymanID, &b.ServiceName, &b.Notes, &b.Status, &b.AmountCents, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repository) ListForCustomer(ctx context.Context, customerID string) ([]Booking, error) {
	return r.list(ctx, `WHERE customer_id = $1`, customerID)
}

func (r *repository) ListForHandyman(ctx context.Context, handymanID string) ([]Booking, error) {
	return r.list(ctx, `WHERE handyman_id = $1`, handymanID)
}

func (r *repository) AssignHandyman(ctx context.Context, id string, handymanID string) error {
	query := `UPDATE bookings SET handyman_id = $1, updated_at = now() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, handymanID, id)
	return err
}

func (r *repository) UpdateStatus(ctx context.Context, id string, status Status) error {
	query := `UPDATE bookings SET status = $1, updated_at = now() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *repository) list(ctx context.Context, where string, arg string) ([]Booking, error) {
	query := `
		SELECT id, customer_id, handyman_id, service_name, notes, status, amount_cents, created_at, updated_at
		FROM bookings
		` + where + `
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := make([]Booking, 0)
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.CustomerID, &b.HandymanID, &b.ServiceName, &b.Notes, &b.Status, &b.AmountCents, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}

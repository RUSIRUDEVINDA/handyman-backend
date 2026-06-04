package payment

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, payment *Payment) error
	UpdateProviderResult(ctx context.Context, providerOrderID string, providerPaymentID *string, status Status) error
	UpdateStatus(ctx context.Context, id string, status Status) error
	GetByBookingID(ctx context.Context, bookingID string) (*Payment, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, payment *Payment) error {
	query := `
		INSERT INTO payments (
			id,
			booking_id,
			customer_id,
			method,
			provider,
			provider_payment_id,
			provider_order_id,
			amount_cents,
			currency,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(
		ctx,
		query,
		payment.ID,
		payment.BookingID,
		payment.CustomerID,
		payment.Method,
		payment.Provider,
		payment.ProviderPaymentID,
		payment.ProviderOrderID,
		payment.AmountCents,
		payment.Currency,
		payment.Status,
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	return err
}

func (r *repository) UpdateProviderResult(ctx context.Context, providerOrderID string, providerPaymentID *string, status Status) error {
	query := `
		UPDATE payments
		SET provider_payment_id = COALESCE($1, provider_payment_id),
			status = $2,
			updated_at = now()
		WHERE provider_order_id = $3
	`
	_, err := r.db.Exec(ctx, query, providerPaymentID, status, providerOrderID)
	return err
}

func (r *repository) UpdateStatus(ctx context.Context, id string, status Status) error {
	query := `
		UPDATE payments
		SET status = $1, updated_at = now()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *repository) GetByBookingID(ctx context.Context, bookingID string) (*Payment, error) {
	query := `
		SELECT
			id,
			booking_id,
			customer_id,
			method,
			provider,
			provider_payment_id,
			provider_order_id,
			amount_cents,
			currency,
			status,
			created_at,
			updated_at
		FROM payments
		WHERE booking_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var payment Payment
	err := r.db.QueryRow(ctx, query, bookingID).Scan(
		&payment.ID,
		&payment.BookingID,
		&payment.CustomerID,
		&payment.Method,
		&payment.Provider,
		&payment.ProviderPaymentID,
		&payment.ProviderOrderID,
		&payment.AmountCents,
		&payment.Currency,
		&payment.Status,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

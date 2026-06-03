package payment

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, payment *Payment) error
	UpdateStatusByIntentID(ctx context.Context, paymentIntentID string, status Status) error
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
			stripe_payment_intent_id,
			amount_cents,
			currency,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(
		ctx,
		query,
		payment.ID,
		payment.BookingID,
		payment.CustomerID,
		payment.StripePaymentIntentID,
		payment.AmountCents,
		payment.Currency,
		payment.Status,
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	return err
}

func (r *repository) UpdateStatusByIntentID(ctx context.Context, paymentIntentID string, status Status) error {
	query := `
		UPDATE payments
		SET status = $1, updated_at = now()
		WHERE stripe_payment_intent_id = $2
	`
	_, err := r.db.Exec(ctx, query, status, paymentIntentID)
	return err
}

func (r *repository) GetByBookingID(ctx context.Context, bookingID string) (*Payment, error) {
	query := `
		SELECT
			id,
			booking_id,
			customer_id,
			stripe_payment_intent_id,
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
		&payment.StripePaymentIntentID,
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

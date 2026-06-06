package dashboard

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Customer(ctx context.Context, customerID string) (*CustomerDashboard, error)
	Handyman(ctx context.Context, handymanID string) (*HandymanDashboard, error)
	Admin(ctx context.Context) (*AdminDashboard, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Customer(ctx context.Context, customerID string) (*CustomerDashboard, error) {
	query := `
		SELECT
			(COUNT(*) FILTER (WHERE status IN ('REQUESTED', 'APPROVED', 'PAYMENT_PENDING', 'CONFIRMED', 'IN_PROGRESS')))::int AS active_bookings,
			(COUNT(*) FILTER (WHERE status = 'COMPLETED'))::int AS completed_jobs,
			(SELECT COUNT(*)::int FROM payments WHERE customer_id = $1 AND status = 'PENDING') AS pending_payments,
			(SELECT COUNT(*)::int FROM reviews WHERE customer_id = $1) AS reviews_given,
			(SELECT COALESCE(SUM(amount_cents), 0)::bigint FROM payments WHERE customer_id = $1 AND status = 'SUCCEEDED') AS total_spent_cents
		FROM bookings
		WHERE customer_id = $1
	`

	dashboard := &CustomerDashboard{}
	err := r.db.QueryRow(ctx, query, customerID).Scan(
		&dashboard.ActiveBookings,
		&dashboard.CompletedJobs,
		&dashboard.PendingPayments,
		&dashboard.ReviewsGiven,
		&dashboard.TotalSpentCents,
	)
	if err != nil {
		return nil, err
	}

	dashboard.RecentBookings, err = r.listBookings(ctx, `WHERE customer_id = $1`, customerID, 5)
	if err != nil {
		return nil, err
	}
	dashboard.PendingPaymentItems, err = r.listBookings(ctx, `WHERE customer_id = $1 AND status IN ('APPROVED', 'PAYMENT_PENDING')`, customerID, 5)
	if err != nil {
		return nil, err
	}
	return dashboard, nil
}

func (r *repository) Handyman(ctx context.Context, handymanID string) (*HandymanDashboard, error) {
	query := `
		SELECT
			(SELECT COUNT(*)::int FROM bookings WHERE status = 'REQUESTED' AND (handyman_id IS NULL OR handyman_id = $1)) AS pending_requests,
			(SELECT COUNT(*)::int FROM bookings WHERE handyman_id = $1 AND status IN ('APPROVED', 'PAYMENT_PENDING', 'CONFIRMED', 'IN_PROGRESS')) AS active_jobs,
			(SELECT COUNT(*)::int FROM bookings WHERE handyman_id = $1 AND status = 'COMPLETED') AS completed_jobs,
			(SELECT COALESCE(SUM(p.amount_cents), 0)::bigint FROM payments p JOIN bookings b ON b.id = p.booking_id WHERE b.handyman_id = $1 AND p.status = 'SUCCEEDED') AS total_earnings_cents,
			(SELECT COALESCE(SUM(p.amount_cents), 0)::bigint FROM payments p JOIN bookings b ON b.id = p.booking_id WHERE b.handyman_id = $1 AND p.method = 'CASH' AND p.status = 'PENDING') AS pending_cash_cents,
			COALESCE((SELECT rating::float8 FROM handymen WHERE user_id = $1), 0) AS rating,
			COALESCE((SELECT review_count FROM handymen WHERE user_id = $1), 0)::int AS review_count
	`

	dashboard := &HandymanDashboard{}
	err := r.db.QueryRow(ctx, query, handymanID).Scan(
		&dashboard.PendingRequests,
		&dashboard.ActiveJobs,
		&dashboard.CompletedJobs,
		&dashboard.TotalEarningsCents,
		&dashboard.PendingCashCents,
		&dashboard.Rating,
		&dashboard.ReviewCount,
	)
	if err != nil {
		return nil, err
	}

	dashboard.RecentJobs, err = r.listBookings(ctx, `WHERE handyman_id = $1 OR (handyman_id IS NULL AND status = 'REQUESTED')`, handymanID, 5)
	if err != nil {
		return nil, err
	}
	return dashboard, nil
}

func (r *repository) Admin(ctx context.Context) (*AdminDashboard, error) {
	query := `
		SELECT
			(SELECT COUNT(*)::int FROM users) AS total_users,
			(SELECT COUNT(*)::int FROM users WHERE role = 'CUSTOMER') AS total_customers,
			(SELECT COUNT(*)::int FROM users WHERE role = 'HANDYMAN') AS total_handymen,
			(SELECT COUNT(*)::int FROM bookings) AS total_bookings,
			(SELECT COUNT(*)::int FROM bookings WHERE status IN ('REQUESTED', 'APPROVED', 'PAYMENT_PENDING', 'CONFIRMED', 'IN_PROGRESS')) AS active_bookings,
			(SELECT COUNT(*)::int FROM bookings WHERE status = 'COMPLETED') AS completed_bookings,
			(SELECT COALESCE(SUM(amount_cents), 0)::bigint FROM payments WHERE status = 'SUCCEEDED') AS total_revenue_cents,
			(SELECT COUNT(*)::int FROM payments WHERE status = 'PENDING') AS pending_payments,
			(SELECT COALESCE(SUM(amount_cents), 0)::bigint FROM payments WHERE method = 'CASH' AND status = 'PENDING') AS pending_cash_cents
	`

	dashboard := &AdminDashboard{}
	err := r.db.QueryRow(ctx, query).Scan(
		&dashboard.TotalUsers,
		&dashboard.TotalCustomers,
		&dashboard.TotalHandymen,
		&dashboard.TotalBookings,
		&dashboard.ActiveBookings,
		&dashboard.CompletedBookings,
		&dashboard.TotalRevenueCents,
		&dashboard.PendingPayments,
		&dashboard.PendingCashCents,
	)
	if err != nil {
		return nil, err
	}

	dashboard.RecentBookings, err = r.listBookings(ctx, `WHERE TRUE`, "", 8)
	if err != nil {
		return nil, err
	}
	return dashboard, nil
}

func (r *repository) listBookings(ctx context.Context, where string, arg string, limit int) ([]BookingSummary, error) {
	limitParam := "$2"
	if arg == "" {
		limitParam = "$1"
	}

	query := `
		SELECT id, service_name, status, amount_cents, created_at
		FROM bookings
		` + where + `
		ORDER BY created_at DESC
		LIMIT ` + limitParam + `
	`

	var rows pgxRows
	var err error
	if arg == "" {
		rows, err = r.db.Query(ctx, query, limit)
	} else {
		rows, err = r.db.Query(ctx, query, arg, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]BookingSummary, 0)
	for rows.Next() {
		var item BookingSummary
		if err := rows.Scan(&item.ID, &item.ServiceName, &item.Status, &item.AmountCents, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type pgxRows interface {
	Close()
	Next() bool
	Scan(dest ...any) error
	Err() error
}

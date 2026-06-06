package dashboard

import "time"

type BookingSummary struct {
	ID          string    `json:"id"`
	ServiceName string    `json:"service_name"`
	Status      string    `json:"status"`
	AmountCents int64     `json:"amount_cents"`
	CreatedAt   time.Time `json:"created_at"`
}

type CustomerDashboard struct {
	ActiveBookings      int              `json:"active_bookings"`
	CompletedJobs       int              `json:"completed_jobs"`
	PendingPayments     int              `json:"pending_payments"`
	ReviewsGiven        int              `json:"reviews_given"`
	TotalSpentCents     int64            `json:"total_spent_cents"`
	RecentBookings      []BookingSummary `json:"recent_bookings"`
	PendingPaymentItems []BookingSummary `json:"pending_payment_items"`
}

type HandymanDashboard struct {
	PendingRequests    int              `json:"pending_requests"`
	ActiveJobs         int              `json:"active_jobs"`
	CompletedJobs      int              `json:"completed_jobs"`
	TotalEarningsCents int64            `json:"total_earnings_cents"`
	PendingCashCents   int64            `json:"pending_cash_cents"`
	Rating             float64          `json:"rating"`
	ReviewCount        int              `json:"review_count"`
	RecentJobs         []BookingSummary `json:"recent_jobs"`
}

type AdminDashboard struct {
	TotalUsers        int              `json:"total_users"`
	TotalCustomers    int              `json:"total_customers"`
	TotalHandymen     int              `json:"total_handymen"`
	TotalBookings     int              `json:"total_bookings"`
	ActiveBookings    int              `json:"active_bookings"`
	CompletedBookings int              `json:"completed_bookings"`
	TotalRevenueCents int64            `json:"total_revenue_cents"`
	PendingPayments   int              `json:"pending_payments"`
	PendingCashCents  int64            `json:"pending_cash_cents"`
	RecentBookings    []BookingSummary `json:"recent_bookings"`
}

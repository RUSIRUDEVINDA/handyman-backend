package review

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, review *Review) error
	ListForHandyman(ctx context.Context, handymanID string) ([]Review, error)
	StatsForHandyman(ctx context.Context, handymanID string) (*RatingStats, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, review *Review) error {
	query := `
		INSERT INTO reviews (id, booking_id, customer_id, handyman_id, rating, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, review.ID, review.BookingID, review.CustomerID, review.HandymanID, review.Rating, review.Comment, review.CreatedAt)
	return err
}

func (r *repository) ListForHandyman(ctx context.Context, handymanID string) ([]Review, error) {
	query := `
		SELECT id, booking_id, customer_id, handyman_id, rating, comment, created_at
		FROM reviews
		WHERE handyman_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, handymanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]Review, 0)
	for rows.Next() {
		var review Review
		if err := rows.Scan(&review.ID, &review.BookingID, &review.CustomerID, &review.HandymanID, &review.Rating, &review.Comment, &review.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, rows.Err()
}

func (r *repository) StatsForHandyman(ctx context.Context, handymanID string) (*RatingStats, error) {
	query := `SELECT COALESCE(AVG(rating), 0), COUNT(*) FROM reviews WHERE handyman_id = $1`

	var stats RatingStats
	err := r.db.QueryRow(ctx, query, handymanID).Scan(&stats.Average, &stats.Count)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

package handyman

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Upsert(ctx context.Context, handyman *Handyman) error
	GetByUserID(ctx context.Context, userID string) (*Handyman, error)
	GetByID(ctx context.Context, id string) (*Handyman, error)
	Search(ctx context.Context, skill string) ([]Handyman, error)
	UpdateRating(ctx context.Context, handymanID string, rating float64, reviewCount int) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Upsert(ctx context.Context, h *Handyman) error {
	query := `
		INSERT INTO handymen (id, user_id, full_name, phone, bio, skills, hourly_rate, status, rating, review_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (user_id)
		DO UPDATE SET
			full_name = EXCLUDED.full_name,
			phone = EXCLUDED.phone,
			bio = EXCLUDED.bio,
			skills = EXCLUDED.skills,
			hourly_rate = EXCLUDED.hourly_rate,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(ctx, query, h.ID, h.UserID, h.FullName, h.Phone, h.Bio, h.Skills, h.HourlyRate, h.Status, h.Rating, h.ReviewCount, h.CreatedAt, h.UpdatedAt)
	return err
}

func (r *repository) GetByUserID(ctx context.Context, userID string) (*Handyman, error) {
	return r.getOne(ctx, `WHERE user_id = $1`, userID)
}

func (r *repository) GetByID(ctx context.Context, id string) (*Handyman, error) {
	return r.getOne(ctx, `WHERE id = $1`, id)
}

func (r *repository) Search(ctx context.Context, skill string) ([]Handyman, error) {
	query := `
		SELECT id, user_id, full_name, phone, bio, skills, hourly_rate, status, rating, review_count, created_at, updated_at
		FROM handymen
		WHERE ($1 = '' OR $1 = ANY(skills))
		ORDER BY rating DESC, review_count DESC, full_name ASC
	`
	rows, err := r.db.Query(ctx, query, skill)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	handymen := make([]Handyman, 0)
	for rows.Next() {
		var h Handyman
		if err := rows.Scan(&h.ID, &h.UserID, &h.FullName, &h.Phone, &h.Bio, &h.Skills, &h.HourlyRate, &h.Status, &h.Rating, &h.ReviewCount, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		handymen = append(handymen, h)
	}
	return handymen, rows.Err()
}

func (r *repository) UpdateRating(ctx context.Context, handymanID string, rating float64, reviewCount int) error {
	query := `UPDATE handymen SET rating = $1, review_count = $2, updated_at = now() WHERE id = $3`
	_, err := r.db.Exec(ctx, query, rating, reviewCount, handymanID)
	return err
}

func (r *repository) getOne(ctx context.Context, where string, arg string) (*Handyman, error) {
	query := `
		SELECT id, user_id, full_name, phone, bio, skills, hourly_rate, status, rating, review_count, created_at, updated_at
		FROM handymen
		` + where

	var h Handyman
	err := r.db.QueryRow(ctx, query, arg).Scan(&h.ID, &h.UserID, &h.FullName, &h.Phone, &h.Bio, &h.Skills, &h.HourlyRate, &h.Status, &h.Rating, &h.ReviewCount, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

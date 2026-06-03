package location

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	UpdateLocation(ctx context.Context, location *HandymanLocation) error
	FindNearby(ctx context.Context, lat float64, lng float64, radiusKm float64) ([]NearbyHandyman, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) UpdateLocation(ctx context.Context, location *HandymanLocation) error {
	query := `
		INSERT INTO handyman_locations (handyman_id, geom, updated_at)
		VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), $4)
		ON CONFLICT (handyman_id)
		DO UPDATE SET geom = EXCLUDED.geom, updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(ctx, query, location.HandymanID, location.Lng, location.Lat, location.UpdatedAt)
	return err
}

func (r *repository) FindNearby(ctx context.Context, lat float64, lng float64, radiusKm float64) ([]NearbyHandyman, error) {
	query := `
		SELECT
			handyman_id,
			ST_Distance(
				geom::geography,
				ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
			) / 1000 AS distance_km
		FROM handyman_locations
		WHERE ST_DWithin(
			geom::geography,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
			$3 * 1000
		)
		ORDER BY distance_km ASC
	`
	rows, err := r.db.Query(ctx, query, lng, lat, radiusKm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nearby := make([]NearbyHandyman, 0)
	for rows.Next() {
		var item NearbyHandyman
		if err := rows.Scan(&item.HandymanID, &item.DistanceKm); err != nil {
			return nil, err
		}
		nearby = append(nearby, item)
	}
	return nearby, rows.Err()
}

package location

import (
	"context"
	"errors"
	"time"

	"github.com/RUSIRUDEVINDA/handyman-backend/internal/events"
)

type Service interface {
	UpdateLocation(ctx context.Context, handymanID string, req UpdateLocationRequest) (*HandymanLocation, error)
	FindNearby(ctx context.Context, lat float64, lng float64, radiusKm float64) ([]NearbyHandyman, error)
}

type service struct {
	repo Repository
	bus  events.Bus
}

func NewService(repo Repository, bus events.Bus) Service {
	return &service{repo: repo, bus: bus}
}

func (s *service) UpdateLocation(ctx context.Context, handymanID string, req UpdateLocationRequest) (*HandymanLocation, error) {
	if req.Lat < -90 || req.Lat > 90 || req.Lng < -180 || req.Lng > 180 {
		return nil, errors.New("invalid latitude or longitude")
	}

	location := &HandymanLocation{
		HandymanID: handymanID,
		Lat:        req.Lat,
		Lng:        req.Lng,
		UpdatedAt:  time.Now().UTC(),
	}
	if err := s.repo.UpdateLocation(ctx, location); err != nil {
		return nil, err
	}
	s.bus.Publish(events.LocationUpdated, location)
	return location, nil
}

func (s *service) FindNearby(ctx context.Context, lat float64, lng float64, radiusKm float64) ([]NearbyHandyman, error) {
	if radiusKm <= 0 {
		radiusKm = 10
	}
	return s.repo.FindNearby(ctx, lat, lng, radiusKm)
}

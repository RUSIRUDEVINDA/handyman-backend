package dashboard

import "context"

type Service interface {
	Customer(ctx context.Context, customerID string) (*CustomerDashboard, error)
	Handyman(ctx context.Context, handymanID string) (*HandymanDashboard, error)
	Admin(ctx context.Context) (*AdminDashboard, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Customer(ctx context.Context, customerID string) (*CustomerDashboard, error) {
	return s.repo.Customer(ctx, customerID)
}

func (s *service) Handyman(ctx context.Context, handymanID string) (*HandymanDashboard, error) {
	return s.repo.Handyman(ctx, handymanID)
}

func (s *service) Admin(ctx context.Context) (*AdminDashboard, error) {
	return s.repo.Admin(ctx)
}

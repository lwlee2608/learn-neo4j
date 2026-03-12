package service

import (
	"context"

	"github.com/lwlee2608/learn-neo4j/internal/domain"
	"github.com/lwlee2608/learn-neo4j/internal/repository"
)

type SupplyChainService struct {
	repo *repository.SupplyChainRepository
}

func NewSupplyChainService(repo *repository.SupplyChainRepository) *SupplyChainService {
	return &SupplyChainService{repo: repo}
}

func (s *SupplyChainService) CreateCompany(ctx context.Context, company domain.Company) error {
	return s.repo.CreateCompany(ctx, company)
}

func (s *SupplyChainService) ListCompanies(ctx context.Context) ([]domain.Company, error) {
	return s.repo.ListCompanies(ctx)
}

func (s *SupplyChainService) GetCompany(ctx context.Context, name string) (*domain.CompanyWithRelations, error) {
	return s.repo.GetCompany(ctx, name)
}

func (s *SupplyChainService) CreateSuppliesEquipmentTo(ctx context.Context, rel domain.SuppliesEquipmentTo) error {
	return s.repo.CreateSuppliesEquipmentTo(ctx, rel)
}

func (s *SupplyChainService) CreateManufacturesFor(ctx context.Context, rel domain.ManufacturesFor) error {
	return s.repo.CreateManufacturesFor(ctx, rel)
}

func (s *SupplyChainService) CreateSuppliesChipsTo(ctx context.Context, rel domain.SuppliesChipsTo) error {
	return s.repo.CreateSuppliesChipsTo(ctx, rel)
}

func (s *SupplyChainService) CreateCompetesWith(ctx context.Context, rel domain.CompetesWith) error {
	return s.repo.CreateCompetesWith(ctx, rel)
}

func (s *SupplyChainService) CreateProvidesCloudFor(ctx context.Context, rel domain.ProvidesCloudFor) error {
	return s.repo.CreateProvidesCloudFor(ctx, rel)
}

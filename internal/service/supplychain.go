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

func (s *SupplyChainService) CreateChip(ctx context.Context, chip domain.Chip) error {
	return s.repo.CreateChip(ctx, chip)
}

func (s *SupplyChainService) ListChips(ctx context.Context) ([]domain.Chip, error) {
	return s.repo.ListChips(ctx)
}

func (s *SupplyChainService) GetChip(ctx context.Context, name string) (*domain.ChipWithRelations, error) {
	return s.repo.GetChip(ctx, name)
}

func (s *SupplyChainService) CreateDesigned(ctx context.Context, rel domain.Designed) error {
	return s.repo.CreateDesigned(ctx, rel)
}

func (s *SupplyChainService) CreateManufactures(ctx context.Context, rel domain.Manufactures) error {
	return s.repo.CreateManufactures(ctx, rel)
}

func (s *SupplyChainService) CreateSuppliesEquipmentTo(ctx context.Context, rel domain.SuppliesEquipmentTo) error {
	return s.repo.CreateSuppliesEquipmentTo(ctx, rel)
}

func (s *SupplyChainService) CreateProvidesCloudFor(ctx context.Context, rel domain.ProvidesCloudFor) error {
	return s.repo.CreateProvidesCloudFor(ctx, rel)
}

func (s *SupplyChainService) CreateUses(ctx context.Context, rel domain.Uses) error {
	return s.repo.CreateUses(ctx, rel)
}

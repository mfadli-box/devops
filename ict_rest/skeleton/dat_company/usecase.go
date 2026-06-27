package dat_company

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

type useCase struct {
	repo Repository
}

func NUseCase(r Repository) UseCase {
	return &useCase{repo: r}
}

func (u *useCase) PLCompany(ctx context.Context) ([]CompanyItem, error) {
	return u.repo.PLCompany(ctx)
}

func (u *useCase) PLCompanyUser(ctx context.Context, userID string) ([]CompanyItem, error) {
	return u.repo.PLCompanyUser(ctx, strings.TrimSpace(userID))
}

func (u *useCase) ALCompany(ctx context.Context) ([]CompanyItem, error) {
	return u.repo.ALCompany(ctx)
}

func (u *useCase) ACCompany(ctx context.Context, req CompanyEdit) error {
	req.Slug = strings.TrimSpace(req.Slug)
	req.Name = strings.TrimSpace(req.Name)
	req.Valuta = strings.TrimSpace(req.Valuta)
	if req.Slug == "" || req.Name == "" {
		return errors.New("Kode dan nama wajib diisi")
	}
	if req.Valuta == "" {
		req.Valuta = "IDR"
	}
	return u.repo.ACCompany(ctx, CompanyItem{
		ID:       uuid.New().String(),
		Slug:     req.Slug,
		Name:     req.Name,
		VatID:    strings.TrimSpace(req.VatID),
		RegNo:    strings.TrimSpace(req.RegNo),
		Address:  strings.TrimSpace(req.Address),
		Valuta:   req.Valuta,
		HrisLink: strings.TrimSpace(req.HrisLink),
		IsActive: req.IsActive,
	})
}

func (u *useCase) AUCompany(ctx context.Context, id string, req CompanyEdit) error {
	id = strings.TrimSpace(id)
	req.Slug = strings.TrimSpace(req.Slug)
	req.Name = strings.TrimSpace(req.Name)
	req.Valuta = strings.TrimSpace(req.Valuta)
	if id == "" || req.Slug == "" || req.Name == "" {
		return errors.New("ID, Kode dan nama wajib diisi")
	}
	if req.Valuta == "" {
		req.Valuta = "IDR"
	}
	return u.repo.AUCompany(ctx, CompanyItem{
		ID:       id,
		Slug:     req.Slug,
		Name:     req.Name,
		VatID:    strings.TrimSpace(req.VatID),
		RegNo:    strings.TrimSpace(req.RegNo),
		Address:  strings.TrimSpace(req.Address),
		Valuta:   req.Valuta,
		HrisLink: strings.TrimSpace(req.HrisLink),
		IsActive: req.IsActive,
	})
}

func (u *useCase) ALCompanyModule(ctx context.Context, companyID string) ([]CompanyModuleItem, error) {
	return u.repo.ALCompanyModule(ctx, strings.TrimSpace(companyID))
}

func (u *useCase) ACCompanyModule(ctx context.Context, req CompanyModuleEdit) error {
	req.CompanyID = strings.TrimSpace(req.CompanyID)
	req.ModuleID = strings.TrimSpace(req.ModuleID)
	if req.CompanyID == "" || req.ModuleID == "" {
		return errors.New("Perusahaan dan Modul wajib diisi")
	}
	return u.repo.ACCompanyModule(ctx, CompanyModuleItem{
		ID:        uuid.New().String(),
		CompanyID: req.CompanyID,
		ModuleID:  req.ModuleID,
		IsActive:  req.IsActive,
	})
}

func (u *useCase) AUCompanyModule(ctx context.Context, id string, req CompanyModuleEdit) error {
	id = strings.TrimSpace(id)
	req.CompanyID = strings.TrimSpace(req.CompanyID)
	req.ModuleID = strings.TrimSpace(req.ModuleID)
	if id == "" || req.CompanyID == "" || req.ModuleID == "" {
		return errors.New("ID, Perusahaan dan Modul wajib diisi")
	}
	return u.repo.AUCompanyModule(ctx, CompanyModuleItem{
		ID:        id,
		CompanyID: req.CompanyID,
		ModuleID:  req.ModuleID,
		IsActive:  req.IsActive,
	})
}

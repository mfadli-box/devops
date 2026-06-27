package dat_company

import (
	"context"
	"database/sql"
)

type repository struct {
	db *sql.DB
}

func NRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) PLCompany(ctx context.Context) ([]CompanyItem, error) {
	query := `
		SELECT id, name, slug
		FROM   "dat_company"
		WHERE  is_active = true
		ORDER BY name
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CompanyItem
	for rows.Next() {
		var c CompanyItem
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Slug); err != nil {
			continue
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *repository) PLCompanyUser(ctx context.Context, userID string) ([]CompanyItem, error) {
	query := `
		SELECT c.id, c.name, c.slug
		FROM   "dat_user_company" uc
		JOIN   "dat_company" c ON c.id = uc.company_id
		WHERE  uc.user_id = $1
		  AND  uc.is_active = true
		  AND  c.is_active = true
		ORDER BY c.name
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CompanyItem
	for rows.Next() {
		var c CompanyItem
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Slug); err != nil {
			continue
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *repository) ALCompany(ctx context.Context) ([]CompanyItem, error) {
	query := `
		SELECT  id, slug, name,
				COALESCE(vat_id, ''),
				COALESCE(reg_no, ''),
				COALESCE(address, ''),
				valuta,
				COALESCE(hris_link, ''),
				is_active
		FROM   "dat_company"
		ORDER BY name
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CompanyItem, 0)
	for rows.Next() {
		var c CompanyItem
		if err := rows.Scan(
			&c.ID,
			&c.Slug,
			&c.Name,
			&c.VatID,
			&c.RegNo,
			&c.Address,
			&c.Valuta,
			&c.HrisLink,
			&c.IsActive); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *repository) ACCompany(ctx context.Context, company CompanyItem) error {
	query := `
		INSERT INTO "dat_company" (
			id, slug, name, vat_id, reg_no, address, valuta, hris_link, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), $7, NULLIF($8, ''), $9, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		company.ID,
		company.Slug,
		company.Name,
		company.VatID,
		company.RegNo,
		company.Address,
		company.Valuta,
		company.HrisLink,
		company.IsActive)
	return err
}

func (r *repository) AUCompany(ctx context.Context, company CompanyItem) error {
	query := `
		UPDATE "dat_company"
		SET 	slug = $1,
				name = $2,
				vat_id = NULLIF($3, ''),
				reg_no = NULLIF($4, ''),
				address = NULLIF($5, ''),
				valuta = $6,
				hris_link = NULLIF($7, ''),
				is_active = $8,
				updated_at = NOW()
		WHERE   id = $9
	`
	_, err := r.db.ExecContext(ctx, query,
		company.Slug,
		company.Name,
		company.VatID,
		company.RegNo,
		company.Address,
		company.Valuta,
		company.HrisLink,
		company.IsActive,
		company.ID)
	return err
}

func (r *repository) ALCompanyModule(ctx context.Context, companyID string) ([]CompanyModuleItem, error) {
	query := `
		SELECT cm.id,
				cm.company_id,
				cm.module_id,
				COALESCE(m.code, ''),
				COALESCE(m.name, ''),
				cm.is_active
		FROM "dat_company_module" cm
		LEFT JOIN "dat_module" m ON m.id = cm.module_id
		WHERE ($1 = '' OR cm.company_id = $1)
		ORDER BY m.code, m.name
	`
	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CompanyModuleItem, 0)
	for rows.Next() {
		var item CompanyModuleItem
		if err := rows.Scan(
			&item.ID,
			&item.CompanyID,
			&item.ModuleID,
			&item.ModuleCode,
			&item.ModuleName,
			&item.IsActive); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *repository) ACCompanyModule(ctx context.Context, item CompanyModuleItem) error {
	query := `
		INSERT INTO "dat_company_module" (
			id, company_id, module_id, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		item.ID,
		item.CompanyID,
		item.ModuleID,
		item.IsActive)
	return err
}

func (r *repository) AUCompanyModule(ctx context.Context, item CompanyModuleItem) error {
	query := `
		UPDATE "dat_company_module"
		SET 	company_id = $1,
				module_id = $2,
				is_active = $3,
				updated_at = NOW()
		WHERE   id = $4
	`
	_, err := r.db.ExecContext(ctx, query,
		item.CompanyID,
		item.ModuleID,
		item.IsActive,
		item.ID)
	return err
}

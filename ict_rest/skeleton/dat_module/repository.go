package dat_module

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

func (r *repository) PLModule(ctx context.Context, userID, companyID string) ([]ModuleItem, error) {
	query := `
		WITH RECURSIVE allowed AS (
			SELECT m.id, m.parent_id, m.code, m.name, m.path, m.is_page, m.is_active, up.level::text AS level
			FROM dat_user_company uc
			JOIN dat_company_module cm ON cm.company_id = uc.company_id
			JOIN dat_user_privilege up ON up.user_company_id = uc.id AND up.module_id = cm.module_id
			JOIN dat_module m ON m.id = cm.module_id
			WHERE uc.user_id = $1
				AND uc.company_id = $2
				AND uc.is_active = true
				AND cm.is_active = true
				AND m.is_active = true
				AND up.level::text <> 'hide'

			UNION

			SELECT p.id, p.parent_id, p.code, p.name, p.path, p.is_page, p.is_active, NULL::text AS level
			FROM dat_module p
			JOIN allowed a ON a.parent_id = p.id
			WHERE p.is_active = true
		)
		SELECT
			id, parent_id, code, name, path, is_page, is_active,
			CASE MAX(CASE level WHEN 'post' THEN 2 WHEN 'view' THEN 1 ELSE 0 END)
				WHEN 2 THEN 'post'
				WHEN 1 THEN 'view'
				ELSE 'hide'
			END AS level
		FROM allowed
		GROUP BY id, parent_id, code, name, path, is_page, is_active
		ORDER BY name
	`
	rows, err := r.db.QueryContext(ctx, query,
		userID,
		companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ModuleItem
	for rows.Next() {
		var res ModuleItem
		if err := rows.Scan(
			&res.ID,
			&res.ParentID,
			&res.Code,
			&res.Name,
			&res.Path,
			&res.IsPage,
			&res.IsActive,
			&res.Level); err != nil {
			continue
		}
		out = append(out, res)
	}
	return out, nil
}

func (r *repository) ALModule(ctx context.Context) ([]ModuleList, error) {
	query := `
		SELECT id, COALESCE(parent_id, ''), code, name, path, is_page, is_active
		FROM   "dat_module"
		ORDER BY name
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ModuleList, 0)
	for rows.Next() {
		var res ModuleList
		if err := rows.Scan(
			&res.ID,
			&res.ParentID,
			&res.Code,
			&res.Name,
			&res.Path,
			&res.IsPage,
			&res.IsActive); err != nil {
			return nil, err
		}
		out = append(out, res)
	}
	return out, rows.Err()
}

func (r *repository) ACModule(ctx context.Context, req ModuleList) error {
	query := `
		INSERT INTO "dat_module" (
			id, parent_id, code, name, path, is_page, is_active, created_at, updated_at
		) VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		req.ID,
		req.ParentID,
		req.Code,
		req.Name,
		req.Path,
		req.IsPage,
		req.IsActive)
	return err
}

func (r *repository) AUModule(ctx context.Context, req ModuleList) error {
	query := `
		UPDATE "dat_module"
		SET 	parent_id = NULLIF($1, ''),
				code = $2,
				name = $3,
				path = $4,
				is_page = $5,
				is_active = $6,
				updated_at = NOW()
		WHERE   id = $7
	`
	_, err := r.db.ExecContext(ctx, query,
		req.ParentID,
		req.Code,
		req.Name,
		req.Path,
		req.IsPage,
		req.IsActive,
		req.ID)
	return err
}

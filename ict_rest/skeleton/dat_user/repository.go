package dat_user

import (
	"context"
	"database/sql"
	"time"
)

type repository struct {
	db *sql.DB
}

func NRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) PGUserName(ctx context.Context, companyID string, username string) (*UserProfile, error) {
	query := `
		SELECT	id, username, email, password, fullname,
				COALESCE(phone, ''),
				COALESCE(company_id, ''),
				COALESCE(employee_id, ''),
				COALESCE(regional_id, ''),
				COALESCE(office_id, ''),
				COALESCE(department_id, ''),
				COALESCE(division_id, ''),
				role, is_admin, is_hris, is_active
		FROM	"dat_user"
		WHERE	company_id = $1
		AND	username = $2
	`
	var u UserProfile
	err := r.db.QueryRowContext(ctx, query, companyID, username).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.Password,
		&u.FullName,
		&u.Phone,
		&u.CompanyId,
		&u.EmployeeId,
		&u.RegionalId,
		&u.OfficeId,
		&u.DepartmentId,
		&u.DivisionId,
		&u.Role,
		&u.IsAdmin,
		&u.IsHris,
		&u.IsActive)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) PGUserID(ctx context.Context, userID string) (*UserProfile, error) {
	query := `
		SELECT	id, username, email, password, fullname,
				COALESCE(phone, ''),
				COALESCE(company_id, ''),
				COALESCE(employee_id, ''),
				COALESCE(regional_id, ''),
				COALESCE(office_id, ''),
				COALESCE(department_id, ''),
				COALESCE(division_id, ''),
				role, is_admin, is_hris, is_active
		FROM	"dat_user"
		WHERE	id = $1
	`
	var u UserProfile
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.Password,
		&u.FullName,
		&u.Phone,
		&u.CompanyId,
		&u.EmployeeId,
		&u.RegionalId,
		&u.OfficeId,
		&u.DepartmentId,
		&u.DivisionId,
		&u.Role,
		&u.IsAdmin,
		&u.IsHris,
		&u.IsActive)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) PELogin(ctx context.Context, p UserSession) error {
	query := `
		INSERT INTO "dat_user_session" (
			id, user_id, token, ip_address, user_agent, expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		p.ID,
		p.UserID,
		p.Token,
		p.IPAddress,
		p.UserAgent,
		p.ExpiresAt)
	return err
}

func (r *repository) PELogout(ctx context.Context, token string) error {
	query := `DELETE FROM "dat_user_session" WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *repository) PUProfile(ctx context.Context, userID string, req UserProfileEdit) error {
	query := `
		UPDATE "dat_user"
		SET		fullname = $1,
				email = $2,
				phone = NULLIF($3, ''),
				updated_at = NOW()
		WHERE   id = $4
	`
	_, err := r.db.ExecContext(ctx, query,
		req.FullName,
		req.Email,
		req.Phone,
		userID)
	return err
}

func (r *repository) PGPassword(ctx context.Context, userID string) (string, bool, error) {
	query := `SELECT password, is_hris FROM "dat_user" WHERE id = $1`
	var passwordHash string
	var isHris bool
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&passwordHash,
		&isHris)
	if err != nil {
		return "", false, err
	}
	return passwordHash, isHris, nil
}

func (r *repository) PUPassword(ctx context.Context, userID string, passwordHash string) error {
	query := `
		UPDATE "dat_user"
		SET 	password = $1,
				updated_at = NOW()
		WHERE   id = $2
	`
	_, err := r.db.ExecContext(ctx, query,
		passwordHash,
		userID)
	return err
}

func (r *repository) PLHistory(ctx context.Context, userID string, limit int) ([]UserAction, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT	id,
				user_id,
				COALESCE(company_id, ''),
				COALESCE(module_code, ''),
				action,
				path,
				COALESCE(ip_address, ''),
				COALESCE(user_agent, ''),
				created_at
		FROM	"dat_user_action"
		WHERE	user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		if err == sql.ErrNoRows {
			return []UserAction{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	actions := make([]UserAction, 0)
	for rows.Next() {
		item := UserAction{}
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.CompanyID,
			&item.ModuleCode,
			&item.Action,
			&item.Path,
			&item.IPAddress,
			&item.UserAgent,
			&createdAt); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt
		actions = append(actions, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return actions, nil
}

func (r *repository) ALUser(ctx context.Context) ([]UserItem, error) {
	query := `
		SELECT	id, COALESCE(company_id, ''), username, email, fullname, COALESCE(phone, ''), role, is_admin, is_hris, is_active
		FROM	"dat_user"
		ORDER BY username
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]UserItem, 0)
	for rows.Next() {
		var item UserItem
		if err := rows.Scan(
			&item.ID,
			&item.CompanyID,
			&item.Username,
			&item.Email,
			&item.FullName,
			&item.Phone,
			&item.Role,
			&item.IsAdmin,
			&item.IsHris,
			&item.IsActive); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *repository) ACUser(ctx context.Context, user UserItem, passwordHash string) error {
	query := `
		INSERT INTO "dat_user" (
			id, company_id, username, email, password, fullname, phone, role, is_admin, is_hris, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, $9, $10, $11, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.CompanyID,
		user.Username,
		user.Email,
		passwordHash,
		user.FullName,
		user.Phone,
		user.Role,
		user.IsAdmin,
		user.IsHris,
		user.IsActive)
	return err
}

func (r *repository) AUUser(ctx context.Context, user UserItem, passwordHash string) error {
	query := `
		UPDATE "dat_user"
		SET		company_id = $1,
				username = $2,
				email = $3,
				fullname = $4,
				phone = NULLIF($5, ''),
				role = $6,
				is_admin = $7,
				is_hris = $8,
				is_active = $9,
				password = CASE WHEN $10 <> '' THEN $10 ELSE password END,
				updated_at = NOW()
		WHERE   id = $11
	`
	_, err := r.db.ExecContext(ctx, query,
		user.CompanyID,
		user.Username,
		user.Email,
		user.FullName,
		user.Phone,
		user.Role,
		user.IsAdmin,
		user.IsHris,
		user.IsActive,
		passwordHash,
		user.ID)
	return err
}

func (r *repository) ALUserCompany(ctx context.Context, userID string) ([]UserCompanyItem, error) {
	query := `
		SELECT	uc.id,
				uc.user_id,
				uc.company_id,
				COALESCE(c.name, ''),
				uc.is_active
		FROM "dat_user_company" uc
		LEFT JOIN "dat_company" c ON c.id = uc.company_id
		WHERE ($1 = '' OR uc.user_id = $1)
		ORDER BY c.name
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]UserCompanyItem, 0)
	for rows.Next() {
		var item UserCompanyItem
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.CompanyID,
			&item.CompanyName,
			&item.IsActive); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *repository) ACUserCompany(ctx context.Context, item UserCompanyItem) error {
	query := `
		INSERT INTO "dat_user_company" (
			id, user_id, company_id, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		item.ID,
		item.UserID,
		item.CompanyID,
		item.IsActive)
	return err
}

func (r *repository) AUUserCompany(ctx context.Context, item UserCompanyItem) error {
	query := `
		UPDATE "dat_user_company"
		SET 	user_id = $1,
				company_id = $2,
				is_active = $3,
				updated_at = NOW()
		WHERE   id = $4
	`
	_, err := r.db.ExecContext(ctx, query,
		item.UserID,
		item.CompanyID,
		item.IsActive,
		item.ID)
	return err
}

func (r *repository) ALUserPrivilege(ctx context.Context, userCompanyID string) ([]UserPrivilegeItem, error) {
	query := `
		SELECT	up.id,
				up.user_company_id,
				up.module_id,
				COALESCE(m.code, ''),
				COALESCE(m.name, ''),
				up.level::text
		FROM "dat_user_privilege" up
		LEFT JOIN "dat_module" m ON m.id = up.module_id
		WHERE ($1 = '' OR up.user_company_id = $1)
		ORDER BY m.code, m.name
	`
	rows, err := r.db.QueryContext(ctx, query, userCompanyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]UserPrivilegeItem, 0)
	for rows.Next() {
		var item UserPrivilegeItem
		if err := rows.Scan(
			&item.ID,
			&item.UserCompanyID,
			&item.ModuleID,
			&item.ModuleCode,
			&item.ModuleName,
			&item.Level); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *repository) ACUserPrivilege(ctx context.Context, item UserPrivilegeItem) error {
	query := `
		INSERT INTO "dat_user_privilege" (
			id, user_company_id, module_id, level, created_at, updated_at
		) VALUES ($1, $2, $3, $4::dat_action_type, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		item.ID,
		item.UserCompanyID,
		item.ModuleID,
		item.Level)
	return err
}

func (r *repository) AUUserPrivilege(ctx context.Context, item UserPrivilegeItem) error {
	query := `
		UPDATE "dat_user_privilege"
		SET 	user_company_id = $1,
				module_id = $2,
				level = $3::dat_action_type,
				updated_at = NOW()
		WHERE   id = $4
	`
	_, err := r.db.ExecContext(ctx, query,
		item.UserCompanyID,
		item.ModuleID,
		item.Level,
		item.ID)
	return err
}

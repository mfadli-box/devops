package dat_user

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type useCase struct {
	repo Repository
}

func NUseCase(repo Repository) UseCase {
	return &useCase{repo: repo}
}

func (u *useCase) PELogin(ctx context.Context, req UserLoginItem, ip, ua string) (*UserLoginInfo, error) {
	req.Company = strings.TrimSpace(req.Company)
	req.Username = strings.TrimSpace(req.Username)

	user, err := u.repo.PGUserName(ctx, req.Company, req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("Nama Pengguna atau kata sandi salah")
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, errors.New("Akun anda telah dinonaktifkan")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("Nama Pengguna atau kata sandi salah")
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	sessionToken := hex.EncodeToString(b)
	expiresAt := time.Now().Add(24 * time.Hour)

	sessionParams := UserSession{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     sessionToken,
		IPAddress: ip,
		UserAgent: ua,
		ExpiresAt: expiresAt,
	}
	if err := u.repo.PELogin(ctx, sessionParams); err != nil {
		return nil, err
	}

	return &UserLoginInfo{
		Token:     sessionToken,
		ExpiresAt: expiresAt,
		UserProfile: UserProfile{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			Phone:     user.Phone,
			CompanyId: user.CompanyId,
			Role:      user.Role,
			IsAdmin:   user.IsAdmin,
			IsHris:    user.IsHris,
			IsActive:  user.IsActive,
			Companies: user.Companies,
		},
	}, nil
}

func (u *useCase) PELogout(ctx context.Context, token string) error {
	return u.repo.PELogout(ctx, token)
}

func (u *useCase) PLProfile(ctx context.Context, userID string) (*UserProfile, error) {
	return u.repo.PGUserID(ctx, userID)
}

func (u *useCase) PUProfile(ctx context.Context, userID string, req UserProfileEdit) (*UserProfile, error) {
	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.TrimSpace(req.Email)
	req.Phone = strings.TrimSpace(req.Phone)

	if req.FullName == "" || req.Email == "" {
		return nil, errors.New("Nama lengkap dan email wajib diisi")
	}

	if err := u.repo.PUProfile(ctx, userID, req); err != nil {
		return nil, err
	}

	return u.repo.PGUserID(ctx, userID)
}

func (u *useCase) PUPassword(ctx context.Context, userID string, req UserPasswordEdit) error {
	req.CurrentPassword = strings.TrimSpace(req.CurrentPassword)
	req.NewPassword = strings.TrimSpace(req.NewPassword)
	req.ConfirmPassword = strings.TrimSpace(req.ConfirmPassword)

	if req.CurrentPassword == "" || req.NewPassword == "" || req.ConfirmPassword == "" {
		return errors.New("Kata sandi saat ini, kata sandi baru, dan konfirmasi kata sandi wajib diisi")
	}

	if req.NewPassword != req.ConfirmPassword {
		return errors.New("Konfirmasi kata sandi tidak sama")
	}

	if len(req.NewPassword) < 8 {
		return errors.New("Kata sandi baru minimal 8 karakter")
	}

	passwordHash, isHris, err := u.repo.PGPassword(ctx, userID)
	if err != nil {
		return err
	}

	if isHris {
		return errors.New("Fitur ubah kata sandi tidak tersedia untuk pengguna HRIS")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.CurrentPassword)); err != nil {
		return errors.New("Kata sandi saat ini tidak valid")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return u.repo.PUPassword(ctx, userID, string(newHash))
}

func (u *useCase) PLHistory(ctx context.Context, userID string, limit int) ([]UserAction, error) {
	list, err := u.repo.PLHistory(ctx, userID, limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []UserAction{}, nil
		}
		return nil, err
	}
	if list == nil {
		return []UserAction{}, nil
	}
	return list, nil
}

func (u *useCase) ALUser(ctx context.Context) ([]UserItem, error) {
	return u.repo.ALUser(ctx)
}

func (u *useCase) ACUser(ctx context.Context, req UserEdit) error {
	req.CompanyID = strings.TrimSpace(req.CompanyID)
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	req.FullName = strings.TrimSpace(req.FullName)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Role = strings.TrimSpace(req.Role)
	if req.Username == "" || req.Email == "" || req.FullName == "" {
		return errors.New("Nama Pengguna, email, dan nama lengkap wajib diisi")
	}
	if req.Password == "" {
		return errors.New("Kata sandi wajib diisi untuk pengguna baru")
	}
	if req.Role == "" {
		req.Role = "staff"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return u.repo.ACUser(ctx, UserItem{
		ID:        uuid.New().String(),
		CompanyID: req.CompanyID,
		Username:  req.Username,
		Email:     req.Email,
		FullName:  req.FullName,
		Phone:     req.Phone,
		Role:      req.Role,
		IsAdmin:   req.IsAdmin,
		IsHris:    req.IsHris,
		IsActive:  req.IsActive,
	}, string(hash))
}

func (u *useCase) AUUser(ctx context.Context, id string, req UserEdit) error {
	id = strings.TrimSpace(id)
	req.CompanyID = strings.TrimSpace(req.CompanyID)
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	req.FullName = strings.TrimSpace(req.FullName)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Role = strings.TrimSpace(req.Role)
	if id == "" || req.Username == "" || req.Email == "" || req.FullName == "" {
		return errors.New("Nama Pengguna, email, dan nama lengkap wajib diisi")
	}
	if req.Role == "" {
		req.Role = "staff"
	}
	passwordHash := ""
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		passwordHash = string(hash)
	}
	return u.repo.AUUser(ctx, UserItem{
		ID:        id,
		CompanyID: req.CompanyID,
		Username:  req.Username,
		Email:     req.Email,
		FullName:  req.FullName,
		Phone:     req.Phone,
		Role:      req.Role,
		IsAdmin:   req.IsAdmin,
		IsHris:    req.IsHris,
		IsActive:  req.IsActive,
	}, passwordHash)
}

func (u *useCase) ALUserCompany(ctx context.Context, userID string) ([]UserCompanyItem, error) {
	return u.repo.ALUserCompany(ctx, strings.TrimSpace(userID))
}

func (u *useCase) ACUserCompany(ctx context.Context, req UserCompanyEdit) error {
	req.UserID = strings.TrimSpace(req.UserID)
	req.CompanyID = strings.TrimSpace(req.CompanyID)
	if req.UserID == "" || req.CompanyID == "" {
		return errors.New("ID, ID Pengguna, dan ID Perusahaan wajib diisi")
	}
	return u.repo.ACUserCompany(ctx, UserCompanyItem{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		CompanyID: req.CompanyID,
		IsActive:  req.IsActive,
	})
}

func (u *useCase) AUUserCompany(ctx context.Context, id string, req UserCompanyEdit) error {
	id = strings.TrimSpace(id)
	req.UserID = strings.TrimSpace(req.UserID)
	req.CompanyID = strings.TrimSpace(req.CompanyID)
	if id == "" || req.UserID == "" || req.CompanyID == "" {
		return errors.New("ID, ID Pengguna, dan ID Perusahaan wajib diisi")
	}
	return u.repo.AUUserCompany(ctx, UserCompanyItem{
		ID:        id,
		UserID:    req.UserID,
		CompanyID: req.CompanyID,
		IsActive:  req.IsActive,
	})
}

func (u *useCase) ALUserPrivilege(ctx context.Context, userCompanyID string) ([]UserPrivilegeItem, error) {
	return u.repo.ALUserPrivilege(ctx, strings.TrimSpace(userCompanyID))
}

func (u *useCase) ACUserPrivilege(ctx context.Context, req UserPrivilegeEdit) error {
	req.UserCompanyID = strings.TrimSpace(req.UserCompanyID)
	req.ModuleID = strings.TrimSpace(req.ModuleID)
	req.Level = strings.TrimSpace(strings.ToLower(req.Level))
	if req.UserCompanyID == "" || req.ModuleID == "" {
		return errors.New("ID Perusahaan dan ID Modul wajib diisi")
	}
	if req.Level == "" {
		req.Level = "hide"
	}
	if req.Level != "hide" && req.Level != "view" && req.Level != "book" && req.Level != "post" {
		return errors.New("Tingkatan hak akses tidak valid")
	}
	return u.repo.ACUserPrivilege(ctx, UserPrivilegeItem{
		ID:            uuid.New().String(),
		UserCompanyID: req.UserCompanyID,
		ModuleID:      req.ModuleID,
		Level:         req.Level,
	})
}

func (u *useCase) AUUserPrivilege(ctx context.Context, id string, req UserPrivilegeEdit) error {
	id = strings.TrimSpace(id)
	req.UserCompanyID = strings.TrimSpace(req.UserCompanyID)
	req.ModuleID = strings.TrimSpace(req.ModuleID)
	req.Level = strings.TrimSpace(strings.ToLower(req.Level))
	if id == "" || req.UserCompanyID == "" || req.ModuleID == "" {
		return errors.New("ID, ID Perusahaan, dan ID Modul wajib diisi")
	}
	if req.Level == "" {
		req.Level = "hide"
	}
	if req.Level != "hide" && req.Level != "view" && req.Level != "book" && req.Level != "post" {
		return errors.New("Tingkatan hak akses tidak valid")
	}
	return u.repo.AUUserPrivilege(ctx, UserPrivilegeItem{
		ID:            id,
		UserCompanyID: req.UserCompanyID,
		ModuleID:      req.ModuleID,
		Level:         req.Level,
	})
}

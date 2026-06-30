package dat_user

import (
	"context"
	"time"
)

/* ======================= dat_action_type = hide | view | book | post
========================== dat_user_privilege
  id                       String            @id @default(uuid())
  user_company_id          String
  module_id                String
  level                    dat_action_type   @default(hide)
  created_at               DateTime          @default(now())
  updated_at               DateTime          @default(now())
  user_company             dat_user_company  @relation(fields: [user_company_id], references: [id])
  module                   dat_module        @relation(fields: [module_id], references: [id])

  @@unique([user_company_id, module_id])
========================== */

type UserPrivilegeInfo struct {
	ID            string `json:"id"`
	UserCompanyID string `json:"user_company_id"`
	ModuleID      string `json:"module_id"`
	Level         string `json:"level"`
}

type UserPrivilegeItem struct {
	ID            string `json:"id"`
	UserCompanyID string `json:"user_company_id"`
	ModuleID      string `json:"module_id"`
	ModuleCode    string `json:"module_code"`
	ModuleName    string `json:"module_name"`
	Level         string `json:"level"`
}

type UserPrivilegeEdit struct {
	UserCompanyID string `json:"user_company_id"`
	ModuleID      string `json:"module_id"`
	Level         string `json:"level"`
}

/* ======================= dat_user_company
  id                       String            @id @default(uuid())
  user_id                  String
  company_id               String
  is_active                Boolean           @default(false)
  created_at               DateTime          @default(now())
  updated_at               DateTime          @default(now())
  user                     dat_user          @relation(fields: [user_id], references: [id])
  company                  dat_company       @relation(fields: [company_id], references: [id])
  privileges               dat_user_privilege[]

  @@unique([user_id, company_id])
========================== */

type UserCompany struct {
	ID          string              `json:"id"`
	UserID      string              `json:"user_id"`
	CompanyID   string              `json:"company_id"`
	CompanyName string              `json:"company_name"`
	Privileges  []UserPrivilegeInfo `json:"privileges"`
}

type UserCompanyItem struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	CompanyID   string `json:"company_id"`
	CompanyName string `json:"company_name"`
	IsActive    bool   `json:"is_active"`
}

type UserCompanyEdit struct {
	UserID    string `json:"user_id"`
	CompanyID string `json:"company_id"`
	IsActive  bool   `json:"is_active"`
}

/* ======================= dat_user
  id                       String            @id @default(uuid())
  username                 String
  email                    String
  password                 String
  fullname                 String
  phone                    String?
  company_id               String            @default("")
  employee_id              String?
  regional_id              String?
  office_id                String?
  department_id            String?
  division_id              String?
  role                     String            @default("staff")
  is_admin                 Boolean           @default(false)
  is_hris                  Boolean           @default(false)
  is_active                Boolean           @default(false)
  created_at               DateTime          @default(now())
  updated_at               DateTime          @default(now())
  companies                dat_user_company[]
  sessions                 dat_user_session[]

  @@unique([company_id, username])
========================== */

type UserProfile struct {
	ID           string        `json:"id"`
	Username     string        `json:"username"`
	Email        string        `json:"email"`
	Password     string        `json:"password"`
	FullName     string        `json:"fullname"`
	Phone        string        `json:"phone"`
	CompanyId    string        `json:"company_id"`
	EmployeeId   string        `json:"employee_id"`
	RegionalId   string        `json:"regional_id"`
	OfficeId     string        `json:"office_id"`
	DepartmentId string        `json:"department_id"`
	DivisionId   string        `json:"division_id"`
	Role         string        `json:"role"`
	IsAdmin      bool          `json:"is_admin"`
	IsHris       bool          `json:"is_hris"`
	IsActive     bool          `json:"is_active"`
	Companies    []UserCompany `json:"companies"`
}

type UserLoginItem struct {
	Company  string `json:"company"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserLoginInfo struct {
	Token       string      `json:"token"`
	ExpiresAt   time.Time   `json:"expires_at"`
	UserProfile UserProfile `json:"user_profile"`
}

type UserProfileEdit struct {
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

type UserPasswordEdit struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type UserItem struct {
	ID        string `json:"id"`
	CompanyID string `json:"company_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FullName  string `json:"fullname"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	IsAdmin   bool   `json:"is_admin"`
	IsHris    bool   `json:"is_hris"`
	IsActive  bool   `json:"is_active"`
}

type UserEdit struct {
	CompanyID string `json:"company_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FullName  string `json:"fullname"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	IsAdmin   bool   `json:"is_admin"`
	IsHris    bool   `json:"is_hris"`
	IsActive  bool   `json:"is_active"`
}

/* ======================= dat_user_session
  id                       String            @id @default(uuid())
  user_id                  String
  token                    String            @unique
  ip_address               String?
  user_agent               String?
  created_at               DateTime          @default(now())
  expires_at               DateTime
  user                     dat_user          @relation(fields: [user_id], references: [id])
========================== */

type UserSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

/* ======================= dat_user_action
  id                       String            @id @default(uuid())
  user_id                  String
  company_id               String?
  module_code              String?
  action                   String
  path                     String
  ip_address               String?
  user_agent               String?
  created_at               DateTime          @default(now())

  @@index([user_id])
  @@index([company_id])
========================== */

type UserAction struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	CompanyID  string    `json:"company_id"`
	ModuleCode string    `json:"module_code"`
	Action     string    `json:"action"`
	Path       string    `json:"path"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
}

type Repository interface {
	PGUserName(ctx context.Context, companyID string, username string) (*UserProfile, error)
	PGUserID(ctx context.Context, userID string) (*UserProfile, error)
	PELogin(ctx context.Context, req UserSession) error
	PELogout(ctx context.Context, token string) error
	PUProfile(ctx context.Context, userID string, req UserProfileEdit) error
	PGPassword(ctx context.Context, userID string) (string, bool, error)
	PUPassword(ctx context.Context, userID string, passwordHash string) error
	PLHistory(ctx context.Context, userID string, limit int) ([]UserAction, error)
	ALUser(ctx context.Context) ([]UserItem, error)
	ACUser(ctx context.Context, req UserItem, passwordHash string) error
	AUUser(ctx context.Context, req UserItem, passwordHash string) error
	ALUserCompany(ctx context.Context, userID string) ([]UserCompanyItem, error)
	ACUserCompany(ctx context.Context, req UserCompanyItem) error
	AUUserCompany(ctx context.Context, req UserCompanyItem) error
	ALUserPrivilege(ctx context.Context, userCompanyID string) ([]UserPrivilegeItem, error)
	ACUserPrivilege(ctx context.Context, req UserPrivilegeItem) error
	AUUserPrivilege(ctx context.Context, req UserPrivilegeItem) error
}

type UseCase interface {
	PELogin(ctx context.Context, req UserLoginItem, ip, ua string) (*UserLoginInfo, error)
	PELogout(ctx context.Context, token string) error
	PLProfile(ctx context.Context, userID string) (*UserProfile, error)
	PUProfile(ctx context.Context, userID string, req UserProfileEdit) (*UserProfile, error)
	PUPassword(ctx context.Context, userID string, req UserPasswordEdit) error
	PLHistory(ctx context.Context, userID string, limit int) ([]UserAction, error)
	ALUser(ctx context.Context) ([]UserItem, error)
	ACUser(ctx context.Context, req UserEdit) error
	AUUser(ctx context.Context, id string, req UserEdit) error
	ALUserCompany(ctx context.Context, userID string) ([]UserCompanyItem, error)
	ACUserCompany(ctx context.Context, req UserCompanyEdit) error
	AUUserCompany(ctx context.Context, id string, req UserCompanyEdit) error
	ALUserPrivilege(ctx context.Context, userCompanyID string) ([]UserPrivilegeItem, error)
	ACUserPrivilege(ctx context.Context, req UserPrivilegeEdit) error
	AUUserPrivilege(ctx context.Context, id string, req UserPrivilegeEdit) error
}

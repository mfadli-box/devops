package dat_company

import "context"

/* ======================= dat_company
  id                       String            @id @default(uuid())
  slug                     String            @unique
  name                     String
  vat_id                   String?
  reg_no                   String?
  address                  String?
  valuta                   String            @default("IDR")
  hris_link                String?
  is_active                Boolean           @default(false)
  created_at               DateTime          @default(now())
  updated_at               DateTime          @default(now())
  users                    dat_user_company[]
  modules                  dat_company_module[]
========================== */

type CompanyItem struct {
	ID       string `json:"id"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	VatID    string `json:"vat_id"`
	RegNo    string `json:"reg_no"`
	Address  string `json:"address"`
	Valuta   string `json:"valuta"`
	HrisLink string `json:"hris_link"`
	IsActive bool   `json:"is_active"`
}

type CompanyEdit struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	VatID    string `json:"vat_id"`
	RegNo    string `json:"reg_no"`
	Address  string `json:"address"`
	Valuta   string `json:"valuta"`
	HrisLink string `json:"hris_link"`
	IsActive bool   `json:"is_active"`
}

/* ======================= dat_company_module
  id                       String            @id @default(uuid())
  company_id               String
  module_id                String
  is_active                Boolean           @default(false)
  created_at               DateTime          @default(now())
  updated_at               DateTime          @default(now())
  company                  dat_company       @relation(fields: [company_id], references: [id])
  module                   dat_module        @relation(fields: [module_id], references: [id])

  @@unique([company_id, module_id])
========================== */

type CompanyModuleItem struct {
	ID         string `json:"id"`
	CompanyID  string `json:"company_id"`
	ModuleID   string `json:"module_id"`
	ModuleCode string `json:"module_code"`
	ModuleName string `json:"module_name"`
	IsActive   bool   `json:"is_active"`
}

type CompanyModuleEdit struct {
	CompanyID string `json:"company_id"`
	ModuleID  string `json:"module_id"`
	IsActive  bool   `json:"is_active"`
}

type Repository interface {
	PLCompany(ctx context.Context) ([]CompanyItem, error)
	PLCompanyUser(ctx context.Context, userID string) ([]CompanyItem, error)
	ALCompany(ctx context.Context) ([]CompanyItem, error)
	ACCompany(ctx context.Context, company CompanyItem) error
	AUCompany(ctx context.Context, company CompanyItem) error
	ALCompanyModule(ctx context.Context, companyID string) ([]CompanyModuleItem, error)
	ACCompanyModule(ctx context.Context, item CompanyModuleItem) error
	AUCompanyModule(ctx context.Context, item CompanyModuleItem) error
}

type UseCase interface {
	PLCompany(ctx context.Context) ([]CompanyItem, error)
	PLCompanyUser(ctx context.Context, userID string) ([]CompanyItem, error)
	ALCompany(ctx context.Context) ([]CompanyItem, error)
	ACCompany(ctx context.Context, req CompanyEdit) error
	AUCompany(ctx context.Context, id string, req CompanyEdit) error
	ALCompanyModule(ctx context.Context, companyID string) ([]CompanyModuleItem, error)
	ACCompanyModule(ctx context.Context, req CompanyModuleEdit) error
	AUCompanyModule(ctx context.Context, id string, req CompanyModuleEdit) error
}

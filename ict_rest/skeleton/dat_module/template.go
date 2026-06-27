package dat_module

import (
	"context"
	"database/sql"
)

/* ======================= dat_module
  id                       String            @id @default(uuid())
  parent_id                String?
  code                     String            @unique
  name                     String
  path                     String
  is_page                  Boolean           @default(true)
  is_active                Boolean           @default(true)
  created_at               DateTime          @default(now())
  updated_at               DateTime          @default(now())
  parent                   dat_module?       @relation("modular", fields: [parent_id], references: [id])
  children                 dat_module[]      @relation("modular")
  companies                dat_company_module[]
  privileges               dat_user_privilege[]
========================== */

type ModuleItem struct {
	ID       string
	ParentID sql.NullString
	Code     string
	Name     string
	Path     string
	IsPage   bool
	IsActive bool
	Level    string
}

type ModuleNode struct {
	ID       string       `json:"id"`
	Code     string       `json:"code"`
	Title    string       `json:"title"`
	Path     string       `json:"path"`
	IsPage   bool         `json:"is_page"`
	Level    string       `json:"level"`
	Children []ModuleNode `json:"children,omitempty"`
}

type ModuleList struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsPage   bool   `json:"is_page"`
	IsActive bool   `json:"is_active"`
}

type ModuleEdit struct {
	ParentID string `json:"parent_id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsPage   bool   `json:"is_page"`
	IsActive bool   `json:"is_active"`
}

type Repository interface {
	PLModule(ctx context.Context, userID, companyID string) ([]ModuleItem, error)
	ALModule(ctx context.Context) ([]ModuleList, error)
	ACModule(ctx context.Context, item ModuleList) error
	AUModule(ctx context.Context, item ModuleList) error
}

type UseCase interface {
	PLModule(ctx context.Context, userID, companyID string) ([]ModuleNode, error)
	ALModule(ctx context.Context) ([]ModuleList, error)
	ACModule(ctx context.Context, req ModuleEdit) error
	AUModule(ctx context.Context, id string, req ModuleEdit) error
}

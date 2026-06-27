package dat_module

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

func (u *useCase) PLModule(ctx context.Context, userID, companyID string) ([]ModuleNode, error) {
	userID = strings.TrimSpace(userID)
	companyID = strings.TrimSpace(companyID)
	if userID == "" || companyID == "" {
		return []ModuleNode{}, nil
	}

	rows, err := u.repo.PLModule(ctx, userID, companyID)
	if err != nil {
		return nil, err
	}

	nodes := map[string]ModuleNode{}
	children := map[string][]ModuleNode{}
	var roots []ModuleNode

	for _, r := range rows {
		node := ModuleNode{
			ID:     r.ID,
			Code:   r.Code,
			Title:  r.Name,
			Path:   r.Path,
			IsPage: r.IsPage,
			Level:  r.Level,
		}
		nodes[r.ID] = node
		if r.ParentID.Valid {
			children[r.ParentID.String] = append(children[r.ParentID.String], node)
		} else {
			roots = append(roots, node)
		}
	}

	var attach func(n *ModuleNode)
	attach = func(n *ModuleNode) {
		chs := children[n.ID]
		for i := range chs {
			attach(&chs[i])
		}
		if len(chs) > 0 {
			n.Children = chs
		}
	}

	for i := range roots {
		attach(&roots[i])
	}

	var prune func(in []ModuleNode) []ModuleNode
	prune = func(in []ModuleNode) []ModuleNode {
		out := make([]ModuleNode, 0, len(in))
		for _, node := range in {
			node.Children = prune(node.Children)
			if node.IsPage || len(node.Children) > 0 {
				out = append(out, node)
			}
		}
		return out
	}

	return prune(roots), nil
}

func (u *useCase) ALModule(ctx context.Context) ([]ModuleList, error) {
	return u.repo.ALModule(ctx)
}

func (u *useCase) ACModule(ctx context.Context, req ModuleEdit) error {
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	req.Path = strings.TrimSpace(req.Path)
	req.ParentID = strings.TrimSpace(req.ParentID)
	if req.Code == "" || req.Name == "" || req.Path == "" {
		return errors.New("Kode, nama, dan lokasi wajib diisi")
	}
	return u.repo.ACModule(ctx, ModuleList{
		ID:       uuid.New().String(),
		ParentID: req.ParentID,
		Code:     req.Code,
		Name:     req.Name,
		Path:     req.Path,
		IsPage:   req.IsPage,
		IsActive: req.IsActive,
	})
}

func (u *useCase) AUModule(ctx context.Context, id string, req ModuleEdit) error {
	id = strings.TrimSpace(id)
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	req.Path = strings.TrimSpace(req.Path)
	req.ParentID = strings.TrimSpace(req.ParentID)
	if id == "" || req.Code == "" || req.Name == "" || req.Path == "" {
		return errors.New("ID, kode, nama, dan lokasi wajib diisi")
	}
	return u.repo.AUModule(ctx, ModuleList{
		ID:       id,
		ParentID: req.ParentID,
		Code:     req.Code,
		Name:     req.Name,
		Path:     req.Path,
		IsPage:   req.IsPage,
		IsActive: req.IsActive,
	})
}

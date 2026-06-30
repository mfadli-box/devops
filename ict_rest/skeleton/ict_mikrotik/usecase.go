package ict_mikrotik

type useCase struct {
	repo Repository
}

func NUseCase(repo Repository) UseCase {
	return &useCase{repo: repo}
}

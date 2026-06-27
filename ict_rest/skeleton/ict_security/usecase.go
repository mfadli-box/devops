package ict_security

import (
	"context"
	"database/sql"
	"log"

	"github.com/google/uuid"
)

type useCase struct {
	repo Repository
	db   *sql.DB
}

func NUseCase(r Repository) UseCase {
	return &useCase{repo: r}
}

func (u *useCase) NLSLA(ctx context.Context, limit, offset int) ([]SLAInfo, error) {
	return u.repo.NLSLA(ctx, limit, offset)
}

func (u *useCase) NLIPW(ctx context.Context, search string, limit, offset int) ([]IPWInfo, error) {
	return u.repo.NLIPW(ctx, search, limit, offset)
}

func (u *useCase) NDIPW(ctx context.Context, id string) error {
	return u.repo.NDIPW(ctx, id)
}

func (u *useCase) NLIPB(ctx context.Context, search string, limit, offset int) ([]IPBInfo, error) {
	return u.repo.NLIPB(ctx, search, limit, offset)
}

func (u *useCase) NCIPM(ctx context.Context, ip, desc string) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := u.repo.NDIPB(ctx, tx, ip); err != nil {
		log.Printf("Gagal hapus blacklist IP %s: %v", ip, err)
		return err
	}

	wUUID := uuid.NewString()
	if err := u.repo.NCIPW(ctx, tx, wUUID, ip, desc); err != nil {
		log.Printf("Gagal insert whitelist IP %s: %v", ip, err)
		return err
	}

	modifiers, err := u.repo.XGATCDate(ctx, tx, ip)
	if err != nil {
		log.Printf("Gagal kalkulasi riwayat tanggal IP %s: %v", ip, err)
		return err
	}

	if _, err := u.repo.XMLOGSIp(ctx, tx, ip); err != nil {
		log.Printf("Gagal pemindahan log IP %s ke ict_nginx_app: %v", ip, err)
		return err
	}

	if err := u.repo.XDLOGSAtcIp(ctx, tx, ip); err != nil {
		log.Printf("Gagal hapus log serangan lama IP %s: %v", ip, err)
		return err
	}

	if err := u.repo.XDLOGSSum(ctx, tx, ip); err != nil {
		log.Printf("Gagal hapus summary IP %s: %v", ip, err)
		return err
	}

	for _, mod := range modifiers {
		if err := u.repo.XUSLASum(ctx, tx, mod.LogCount, mod.Date); err != nil {
			log.Printf("Gagal update SLA harian %s: %v", mod.Date, err)
			continue
		}
		_ = u.repo.XUSLAPct(ctx, tx, mod.Date)
	}

	return tx.Commit()
}

func (u *useCase) NLWAF(ctx context.Context, search string, limit, offset int) ([]WAFList, error) {
	return u.repo.NLWAF(ctx, search, limit, offset)
}

func (u *useCase) NCWAF(ctx context.Context, req WAFItem) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ruleID := uuid.NewString()
	if err := u.repo.NCWAF(ctx, tx, ruleID, req.Domain, req.URLPath, req.ArgsPattern, req.Description); err != nil {
		log.Printf("Gagal mendaftarkan aturan WAF: %v", err)
		return err
	}

	modifiers, err := u.repo.XGWAFDate(ctx, tx, req.Domain, req.URLPath, req.ArgsPattern)
	if err != nil {
		log.Printf("Gagal kalkulasi riwayat tanggal untuk pembersihan bypass historis: %v", err)
		return err
	}

	migratedRows, err := u.repo.XMLOGSArg(ctx, tx, req.Domain, req.URLPath, req.ArgsPattern)
	if err != nil {
		log.Printf("Gagal sinkronisasi log untuk migrasi: %v", err)
		return err
	}
	log.Printf("Berhasil memigrasikan %d log yang sesuai dengan kriteria aturan ke ict_nginx_app.", migratedRows)

	if err := u.repo.XDLOGSAtcArg(ctx, tx, req.Domain, req.URLPath, req.ArgsPattern); err != nil {
		log.Printf("Gagal menghapus alert false-positive lama dari ict_nginx_atc: %v", err)
		return err
	}

	for _, mod := range modifiers {
		if err := u.repo.XUSLASum(ctx, tx, mod.LogCount, mod.Date); err != nil {
			log.Printf("Gagal penyesuaian SLA dinamis untuk tanggal %s: %v", mod.Date, err)
			continue
		}
		_ = u.repo.XUSLAPct(ctx, tx, mod.Date)
	}

	return tx.Commit()
}

func (u *useCase) NDWAF(ctx context.Context, id string) error {
	log.Printf("Menghapus aturan WAF : %s", id)
	return u.repo.NDWAF(ctx, id)
}

func (u *useCase) NLATC(ctx context.Context, date string) ([]ATCList, error) {
	return u.repo.NLATC(ctx, date)
}

func (u *useCase) NLLOG(ctx context.Context, ip, date string) (*LOGList, error) {
	tables := []string{"ict_nginx_log", "ict_nginx_app", "ict_nginx_atc"}

	combinedResponse := &LOGList{
		ClientIP: ip,
		Date:     date,
		Logs:     []LOGItem{},
	}
	for _, table := range tables {
		res, err := u.repo.NLLOG(ctx, ip, date, table)
		if err == nil && res.TotalHits > 0 {
			combinedResponse.Logs = append(combinedResponse.Logs, res.Logs...)
		}
	}
	combinedResponse.TotalHits = len(combinedResponse.Logs)
	return combinedResponse, nil
}

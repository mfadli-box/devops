package ict_monitor

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type useCase struct {
	repo Repository
}

func NUseCase(r Repository) UseCase {
	return &useCase{repo: r}
}

func (u *useCase) URHook(ctx context.Context, req UptimeAlertItem) error {
	var resp UptimeAlertInfo

	resp.ID = uuid.New().String()
	resp.MonitorID = req.MonitorID
	resp.MonitorURL = req.MonitorURL
	resp.MonitorFriendlyName = req.MonitorFriendlyName
	resp.AlertType = req.AlertType
	resp.AlertTypeFriendlyName = req.AlertTypeFriendlyName
	resp.MonitorAlertContacts = req.MonitorAlertContacts
	resp.DashboardURL = req.DashboardURL
	resp.MonitorType = req.MonitorType

	if req.AlertDetails != "" {
		resp.AlertDetails = &req.AlertDetails
	}
	if req.AlertDuration != "" && req.AlertDuration != "0" {
		if val, err := strconv.Atoi(req.AlertDuration); err == nil {
			resp.AlertDuration = &val
		}
	}
	if req.ResponseTime != "" {
		if val, err := strconv.Atoi(req.ResponseTime); err == nil {
			resp.ResponseTime = &val
		}
	}
	if req.HTTPStatusCode != "" && req.HTTPStatusCode != "0" {
		if val, err := strconv.Atoi(req.HTTPStatusCode); err == nil {
			resp.HTTPStatusCode = &val
		}
	}
	if req.SSLExpiryDaysLeft != "" {
		if val, err := strconv.Atoi(req.SSLExpiryDaysLeft); err == nil {
			resp.SSLExpiryDaysLeft = &val
		}
	}
	if req.AlertDateTime > 0 {
		resp.AlertDateTime = time.Unix(req.AlertDateTime, 0).UTC()
	} else {
		resp.AlertDateTime = time.Now().UTC()
	}
	if req.IncidentStartTime > 0 {
		resp.IncidentStartTime = time.Unix(req.IncidentStartTime, 0).UTC()
	} else {
		resp.IncidentStartTime = time.Now().UTC()
	}
	if req.IncidentEndTime != "" && req.IncidentEndTime != "0" {
		if tInt, err := strconv.ParseInt(req.IncidentEndTime, 10, 64); err == nil && tInt > 0 {
			tTime := time.Unix(tInt, 0).UTC()
			resp.IncidentEndTime = &tTime
		}
	}
	if req.SSLExpiryDate > 0 {
		tTime := time.Unix(req.SSLExpiryDate, 0).UTC()
		resp.SSLExpiryDate = &tTime
	}
	if req.MonitoringRegions != "" {
		resp.MonitoringRegions = []byte(req.MonitoringRegions)
	} else {
		resp.MonitoringRegions, _ = json.Marshal([]string{})
	}
	if req.MonitorTags != "" {
		resp.MonitorTags = []byte(req.MonitorTags)
	} else {
		resp.MonitorTags, _ = json.Marshal([]string{})
	}
	if req.MonitorGroup != "" {
		resp.MonitorGroup = &req.MonitorGroup
	}

	err := u.repo.URHook(ctx, resp)
	if err != nil {
		return err
	}
	go func(monitorID int, monitorURL string, name string) {
		calcCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		targetDate := time.Now().UTC()
		_ = u.repo.URHookSla(calcCtx, targetDate, monitorID, monitorURL, name)
		if resp.IncidentStartTime.Before(time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)) {
			_ = u.repo.URHookSla(calcCtx, resp.IncidentStartTime, monitorID, monitorURL, name)
		}
	}(resp.MonitorID, resp.MonitorURL, resp.MonitorFriendlyName)

	return nil
}

func extractDomain(rawURL string) string {
	cleanURL := strings.TrimSpace(rawURL)
	if !strings.HasPrefix(strings.ToLower(cleanURL), "http://") && !strings.HasPrefix(strings.ToLower(cleanURL), "https://") {
		cleanURL = "http://" + cleanURL
	}
	parsed, err := url.Parse(cleanURL)
	if err != nil {
		return rawURL
	}
	return parsed.Hostname()
}

func (u *useCase) formatMetaResponse(items interface{}, total int, page int, limit int) map[string]interface{} {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	return map[string]interface{}{
		"data": items,
		"meta": map[string]interface{}{
			"current_page": page,
			"per_page":     limit,
			"total_data":   total,
			"total_pages":  (total + limit - 1) / limit,
		},
	}
}

func (u *useCase) URLog(ctx context.Context, f FilterParams) (map[string]interface{}, error) {
	items, total, err := u.repo.URLog(ctx, f)
	if err != nil {
		return nil, err
	}
	return u.formatMetaResponse(items, total, f.Page, f.Limit), nil
}

func (u *useCase) URSla(ctx context.Context, f FilterParams) (map[string]interface{}, error) {
	items, total, err := u.repo.URSla(ctx, f)
	if err != nil {
		return nil, err
	}
	return u.formatMetaResponse(items, total, f.Page, f.Limit), nil
}

func (u *useCase) URSum(ctx context.Context, f FilterParams) (map[string]interface{}, error) {
	items, total, err := u.repo.URSum(ctx, f)
	if err != nil {
		return nil, err
	}
	return u.formatMetaResponse(items, total, f.Page, f.Limit), nil
}

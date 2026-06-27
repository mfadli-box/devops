package ict_monitor

import (
	"context"
	"strconv"
	"time"
)

type useCase struct {
	repo Repository
}

func NUseCase(r Repository) UseCase {
	return &useCase{repo: r}
}

func (u *useCase) URHook(ctx context.Context, req UptimeAlertItem) error {
	resp := &UptimeAlertInfo{
		MonitorID:             req.MonitorID,
		MonitorURL:            req.MonitorURL,
		MonitorFriendlyName:   req.MonitorFriendlyName,
		AlertType:             req.AlertType,
		AlertTypeFriendlyName: req.AlertTypeFriendlyName,
		MonitorAlertContacts:  req.MonitorAlertContacts,
		DashboardURL:          req.DashboardURL,
		MonitorType:           req.MonitorType,
		AlertDateTime:         time.Unix(req.AlertDateTime, 0),
		IncidentStartTime:     time.Unix(req.IncidentStartTime, 0),
		MonitoringRegions:     []byte(req.MonitoringRegions),
		MonitorTags:           []byte(req.MonitorTags),
	}

	if req.AlertDetails != "" {
		resp.AlertDetails = &req.AlertDetails
	}
	if req.MonitorGroup != "" {
		resp.MonitorGroup = &req.MonitorGroup
	}
	if d, err := strconv.Atoi(req.AlertDuration); err == nil {
		resp.AlertDuration = &d
	}
	if h, err := strconv.Atoi(req.HTTPStatusCode); err == nil {
		resp.HTTPStatusCode = &h
	}
	if r, err := strconv.Atoi(req.ResponseTime); err == nil {
		resp.ResponseTime = &r
	}
	if s, err := strconv.Atoi(req.SSLExpiryDaysLeft); err == nil {
		resp.SSLExpiryDaysLeft = &s
	}

	if req.SSLExpiryDate > 0 {
		t := time.Unix(req.SSLExpiryDate, 0)
		resp.SSLExpiryDate = &t
	}
	if req.IncidentEndTime != "" {
		if et, err := strconv.ParseInt(req.IncidentEndTime, 10, 64); err == nil && et > 0 {
			t := time.Unix(et, 0)
			resp.IncidentEndTime = &t
		}
	}

	return u.repo.URHook(ctx, *resp)
}

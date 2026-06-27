package ict_monitor

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type repository struct {
	db *sql.DB
}

func NRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) URHook(ctx context.Context, item UptimeAlertInfo) error {
	query := `
		INSERT INTO "UptimeAlertInfo" (
			id, monitorid, monitorurl, monitorfriendlyname, alerttype, alerttypefriendlyname,
			alertdetails, alertduration, alertdatetime, monitoralertcontacts, sslexpirydate,
			sslexpirydaysleft, dashboardurl, monitortype, httpstatuscode, monitoringregions,
			monitortags, monitorgroup, incidentstarttime, incidentendtime, responsetime, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, NOW())
	`

	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	_, err := r.db.ExecContext(ctx, query,
		item.ID,
		item.MonitorID,
		item.MonitorURL,
		item.MonitorFriendlyName,
		item.AlertType,
		item.AlertTypeFriendlyName,
		item.AlertDetails,
		item.AlertDuration,
		item.AlertDateTime,
		item.MonitorAlertContacts,
		item.SSLExpiryDate,
		item.SSLExpiryDaysLeft,
		item.DashboardURL,
		item.MonitorType,
		item.HTTPStatusCode,
		item.MonitoringRegions,
		item.MonitorTags,
		item.MonitorGroup,
		item.IncidentStartTime,
		item.IncidentEndTime,
		item.ResponseTime,
	)
	return err
}

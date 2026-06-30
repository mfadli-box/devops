package ict_monitor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type repository struct {
	db *sql.DB
}

func NRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) URHook(ctx context.Context, req UptimeAlertInfo) error {
	query := `
		INSERT INTO "ict_uptimerobot_log" (
			id, monitorid, monitorurl, monitorfriendlyname, alerttype, alerttypefriendlyname,
			alertdetails, alertduration, alertdatetime, monitoralertcontacts, sslexpirydate,
			sslexpirydaysleft, dashboardurl, monitortype, httpstatuscode, monitoringregions,
			monitortags, monitorgroup, incidentstarttime, incidentendtime, responsetime, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, NOW())
		ON CONFLICT (monitorid, incidentstarttime) 
		DO UPDATE SET 
			alerttype = EXCLUDED.alerttype,
			alerttypefriendlyname = EXCLUDED.alerttypefriendlyname,
			alertdetails = EXCLUDED.alertdetails,
			alertduration = EXCLUDED.alertduration,
			alertdatetime = EXCLUDED.alertdatetime,
			incidentendtime = EXCLUDED.incidentendtime,
			httpstatuscode = EXCLUDED.httpstatuscode,
			responsetime = EXCLUDED.responsetime
	`

	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	_, err := r.db.ExecContext(ctx, query,
		req.ID,
		req.MonitorID,
		req.MonitorURL,
		req.MonitorFriendlyName,
		req.AlertType,
		req.AlertTypeFriendlyName,
		req.AlertDetails,
		req.AlertDuration,
		req.AlertDateTime,
		req.MonitorAlertContacts,
		req.SSLExpiryDate,
		req.SSLExpiryDaysLeft,
		req.DashboardURL,
		req.MonitorType,
		req.HTTPStatusCode,
		req.MonitoringRegions,
		req.MonitorTags,
		req.MonitorGroup,
		req.IncidentStartTime,
		req.IncidentEndTime,
		req.ResponseTime,
	)
	return err
}

func (r *repository) URHookSla(ctx context.Context, qdate time.Time, qid int, qurl string, qname string) error {
	startOfDay := time.Date(qdate.Year(), qdate.Month(), qdate.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)
	dateStr := startOfDay.Format("2006-01-02")
	domain := extractDomain(qurl)

	queryLogs := `
		SELECT	incidentstarttime, incidentendtime 
		FROM	"ict_uptimerobot_log"
		WHERE	monitorid = $1 
		  AND	alerttype = 1
		  AND	incidentstarttime < $2
		  AND	(incidentendtime >= $3 OR incidentendtime IS NULL)
	`
	rows, err := r.db.QueryContext(ctx, queryLogs, qid, endOfDay, startOfDay)
	if err != nil {
		return err
	}
	defer rows.Close()

	var totalDowntimeSec int = 0

	for rows.Next() {
		var incStart time.Time
		var incEndNull *time.Time
		if err := rows.Scan(&incStart, &incEndNull); err != nil {
			return err
		}

		actualStart := incStart
		if incStart.Before(startOfDay) {
			actualStart = startOfDay
		}

		actualEnd := endOfDay
		if incEndNull != nil && incEndNull.Before(endOfDay) {
			actualEnd = *incEndNull
		} else if incEndNull == nil {
			now := time.Now().UTC()
			if now.Before(endOfDay) {
				actualEnd = now
			}
		}

		diff := actualEnd.Sub(actualStart).Seconds()
		if diff > 0 {
			totalDowntimeSec += int(diff)
		}
	}

	if totalDowntimeSec > 86400 {
		totalDowntimeSec = 86400
	}
	totalUptimeSec := 86400 - totalDowntimeSec
	slaPercentage := (float64(totalUptimeSec) / 86400.0) * 100.0

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	upsertDailyQuery := `
		INSERT INTO "ict_uptimerobot_sla" (
			id, monitor_id, monitor_friendly_name, monitor_url, date,
			total_downtime_sec, total_uptime_sec, sla_percentage, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (monitor_id, date) 
		DO UPDATE SET 
			total_downtime_sec = EXCLUDED.total_downtime_sec,
			total_uptime_sec = EXCLUDED.total_uptime_sec,
			sla_percentage = EXCLUDED.sla_percentage,
			updated_at = NOW();
	`
	dailyID := uuid.New().String()
	_, err = tx.ExecContext(ctx, upsertDailyQuery,
		dailyID,
		qid,
		qname,
		qurl,
		dateStr,
		totalDowntimeSec,
		totalUptimeSec,
		slaPercentage)
	if err != nil {
		return err
	}

	upsertSummaryQuery := `
		INSERT INTO "ict_uptimerobot_sum" (
				id, date, domain, total_monitors, total_downtime_sec, total_uptime_sec, average_sla, updated_at)
		SELECT	$1::UUID, $2::DATE, $3::VARCHAR, COUNT(DISTINCT monitor_id), SUM(total_downtime_sec),
				SUM(total_uptime_sec), AVG(sla_percentage), NOW()
		FROM	"ict_uptimerobot_sla"
		WHERE	date = $2::DATE AND monitor_url LIKE '%' || $3 || '%'
		ON CONFLICT (date, domain) 
		DO UPDATE SET 
			total_monitors = EXCLUDED.total_monitors,
			total_downtime_sec = EXCLUDED.total_downtime_sec,
			total_uptime_sec = EXCLUDED.total_uptime_sec,
			average_sla = EXCLUDED.average_sla,
			updated_at = NOW();
	`
	summaryID := uuid.New().String()
	_, err = tx.ExecContext(ctx, upsertSummaryQuery, summaryID, dateStr, domain)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repository) URLog(ctx context.Context, f FilterParams) ([]UptimeAlertInfo, int, error) {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 10
	}
	offset := (f.Page - 1) * f.Limit

	var items []UptimeAlertInfo
	var total int

	countQuery := `SELECT COUNT(*) FROM "ict_uptimerobot_log" WHERE 1=1`
	var args []interface{}
	idx := 1

	if f.MonitorID > 0 {
		countQuery += ` AND monitor_id = $` + strconv.Itoa(idx)
		args = append(args, f.MonitorID)
		idx++
	}
	if f.Domain != "" {
		countQuery += ` AND monitor_url LIKE '%' || $` + strconv.Itoa(idx) + ` || '%'`
		args = append(args, f.Domain)
		idx++
	}
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT	id, monitor_id, monitor_url, monitor_friendly_name, alert_type, alert_type_friendly_name,
		    	alert_details, alert_duration, alert_datetime, monitor_alert_contacts, ssl_expiry_date,
		    	ssl_expiry_days_left, dashboard_url, monitor_type, http_status_code, incident_start_time, 
				incident_end_time, response_time, created_at 
		FROM 	"ict_uptimerobot_log" WHERE 1=1
	`
	idxData := 1
	var argsData []interface{}
	if f.MonitorID > 0 {
		dataQuery += ` AND monitor_id = $` + strconv.Itoa(idxData)
		argsData = append(argsData, f.MonitorID)
		idxData++
	}
	if f.Domain != "" {
		dataQuery += ` AND monitor_url LIKE '%' || $` + strconv.Itoa(idxData) + ` || '%'`
		argsData = append(argsData, f.Domain)
		idxData++
	}

	dataQuery += ` ORDER BY incident_start_time DESC LIMIT $` + strconv.Itoa(idxData) + ` OFFSET $` + strconv.Itoa(idxData+1)
	argsData = append(argsData, f.Limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, argsData...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var i UptimeAlertInfo
		err := rows.Scan(
			&i.ID, &i.MonitorID, &i.MonitorURL, &i.MonitorFriendlyName, &i.AlertType, &i.AlertTypeFriendlyName,
			&i.AlertDetails, &i.AlertDuration, &i.AlertDateTime, &i.MonitorAlertContacts, &i.SSLExpiryDate,
			&i.SSLExpiryDaysLeft, &i.DashboardURL, &i.MonitorType, &i.HTTPStatusCode, &i.IncidentStartTime,
			&i.IncidentEndTime, &i.ResponseTime, &i.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, i)
	}
	return items, total, nil
}

func (r *repository) URSla(ctx context.Context, f FilterParams) ([]UptimeSla, int, error) {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 10
	}
	offset := (f.Page - 1) * f.Limit

	var items []UptimeSla
	var total int
	var args []interface{}
	idx := 1

	baseWhere := " WHERE 1=1"
	if f.MonitorID > 0 {
		baseWhere += ` AND monitor_id = $` + strconv.Itoa(idx)
		args = append(args, f.MonitorID)
		idx++
	}
	if f.StartDate != "" {
		baseWhere += ` AND date >= $` + strconv.Itoa(idx)
		args = append(args, f.StartDate)
		idx++
	}
	if f.EndDate != "" {
		baseWhere += ` AND date <= $` + strconv.Itoa(idx)
		args = append(args, f.EndDate)
		idx++
	}

	err := r.db.QueryRowContext(ctx, `
		SELECT	COUNT(*)
		FROM	"ict_uptimerobot_sla"`+baseWhere, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT	id, monitor_id, monitor_friendly_name, monitor_url, date, total_downtime_sec,
				total_uptime_sec, sla_percentage
		FROM	"ict_uptimerobot_sla"` + baseWhere + `
		ORDER BY date DESC LIMIT $` + strconv.Itoa(idx) + `
		OFFSET $` + strconv.Itoa(idx+1)
	args = append(args, f.Limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var i UptimeSla
		if err := rows.Scan(
			&i.ID,
			&i.MonitorID,
			&i.MonitorFriendlyName,
			&i.MonitorURL,
			&i.Date,
			&i.TotalDowntimeSec,
			&i.TotalUptimeSec,
			&i.SlaPercentage,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, i)
	}
	return items, total, nil
}

func (r *repository) URSum(ctx context.Context, f FilterParams) ([]UptimeSum, int, error) {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 10
	}
	offset := (f.Page - 1) * f.Limit

	var items []UptimeSum
	var total int
	var args []interface{}
	idx := 1

	baseWhere := " WHERE 1=1"
	if f.Domain != "" {
		baseWhere += ` AND domain = $` + strconv.Itoa(idx)
		args = append(args, f.Domain)
		idx++
	}
	if f.StartDate != "" {
		baseWhere += ` AND date >= $` + strconv.Itoa(idx)
		args = append(args, f.StartDate)
		idx++
	}
	if f.EndDate != "" {
		baseWhere += ` AND date <= $` + strconv.Itoa(idx)
		args = append(args, f.EndDate)
		idx++
	}

	err := r.db.QueryRowContext(ctx, `
		SELECT	COUNT(*)
		FROM	"ict_uptimerobot_sum"`+baseWhere, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT	id, date, domain, total_monitors, total_downtime_sec, total_uptime_sec, average_sla
		FROM	"ict_uptimerobot_sum"` + baseWhere + `
		ORDER BY date DESC LIMIT $` + strconv.Itoa(idx) + `
		OFFSET $` + strconv.Itoa(idx+1)
	args = append(args, f.Limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var i UptimeSum
		if err := rows.Scan(
			&i.ID,
			&i.Date,
			&i.Domain,
			&i.TotalMonitors,
			&i.TotalDowntimeSec,
			&i.TotalUptimeSec,
			&i.AverageSla,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, i)
	}
	return items, total, nil
}

func (r *repository) DURLog(ctx context.Context, logID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var monitorID int
	var rawURL string
	var friendlyName string
	var incidentStart time.Time

	findQuery := `
		SELECT	monitorid, monitorurl, monitorfriendlyname, incidentstarttime 
		FROM	"ict_uptimerobot_log" 
		WHERE	id = $1
	`
	err = tx.QueryRowContext(ctx, findQuery, logID).Scan(
		&monitorID,
		&rawURL,
		&friendlyName,
		&incidentStart)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("Log tidak ditemukan")
		}
		return err
	}

	deleteQuery := `
		DELETE	FROM "ict_uptimerobot_log"
		WHERE	id = $1
	`
	_, err = tx.ExecContext(ctx, deleteQuery, logID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	err = r.URHookSla(ctx, incidentStart, monitorID, rawURL, friendlyName)
	if err != nil {
		return fmt.Errorf("Log berhasil dihapus - SLA Gagal Kalkulasi Ulang %v", err)
	}

	return nil
}

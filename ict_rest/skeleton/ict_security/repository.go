package ict_security

import (
	"context"
	"database/sql"
	"fmt"
)

type repository struct {
	db *sql.DB
}

func NRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) NLSLA(ctx context.Context, limit, offset int) ([]SLAInfo, error) {
	query := `
		SELECT	id, date, total_requests, successful_requests, client_errors,
				server_errors, attack_requests, avg_response_time, sla_percentage 
	    FROM	"ict_nginx_sla"
		ORDER BY date DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []SLAInfo
	for rows.Next() {
		var s SLAInfo
		if err := rows.Scan(
			&s.ID,
			&s.Date,
			&s.TotalRequests,
			&s.SuccessfulRequests,
			&s.ClientErrors,
			&s.ServerErrors,
			&s.AttackRequests,
			&s.AvgResponseTime,
			&s.SLAPercentage); err == nil {
			list = append(list, s)
		}
	}
	if list == nil {
		list = []SLAInfo{}
	}
	return list, nil
}

func (r *repository) NLIPW(ctx context.Context, search string, limit, offset int) ([]IPWInfo, error) {
	query := `
		SELECT	id, ip_or_cidr, description, created_at
	    FROM	"ict_ip_whitelist"
		WHERE	ip_or_cidr LIKE $1 OR description LIKE $1 
	    ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`
	searchT := fmt.Sprintf("%%%s%%", search)
	rows, err := r.db.QueryContext(ctx, query, searchT, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []IPWInfo
	for rows.Next() {
		var w IPWInfo
		if err := rows.Scan(
			&w.ID,
			&w.IPOrCIDR,
			&w.Description,
			&w.CreatedAt); err == nil {
			list = append(list, w)
		}
	}
	if list == nil {
		list = []IPWInfo{}
	}
	return list, nil
}

func (r *repository) NCIPW(ctx context.Context, tx TxOrDB, id, ip, desc string) error {
	query := `
		INSERT INTO "ict_ip_whitelist" (id, ip_or_cidr, description, created_at) 
		VALUES ($1, $2, $3, NOW()) ON CONFLICT (ip_or_cidr) DO NOTHING
	`
	_, err := tx.ExecContext(ctx, query, id, ip, desc)
	return err
}

func (r *repository) NDIPW(ctx context.Context, ip string) error {
	query := `
		DELETE FROM "ict_ip_whitelist" WHERE ip_or_cidr = $1
	`
	_, err := r.db.ExecContext(ctx, query, ip)
	return err
}

func (r *repository) NLIPB(ctx context.Context, search string, limit, offset int) ([]IPBInfo, error) {
	query := `
		SELECT	id, ip, threat_score, reason, banned_at, expires_at 
	    FROM	"ict_ip_blacklist"
		WHERE	ip LIKE $1 OR reason LIKE $1 
	    ORDER BY banned_at DESC LIMIT $2 OFFSET $3
	`
	searchT := fmt.Sprintf("%%%s%%", search)
	rows, err := r.db.QueryContext(ctx, query, searchT, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []IPBInfo
	for rows.Next() {
		var b IPBInfo
		if err := rows.Scan(
			&b.ID,
			&b.IP,
			&b.ThreatScore,
			&b.Reason,
			&b.BannedAt,
			&b.ExpiresAt); err == nil {
			list = append(list, b)
		}
	}
	if list == nil {
		list = []IPBInfo{}
	}
	return list, nil
}

func (r *repository) NDIPB(ctx context.Context, tx TxOrDB, ip string) error {
	query := `
		DELETE FROM "ict_ip_blacklist" WHERE ip = $1
	`
	_, err := tx.ExecContext(ctx, query, ip)
	return err
}

func (r *repository) NLWAF(ctx context.Context, search string, limit, offset int) ([]WAFList, error) {
	query := `
		SELECT	id, domain, url_path, args_pattern, description, created_at 
		FROM	"ict_waf_bypass_rule"
		WHERE	domain LIKE $1 OR url_path LIKE $1 OR description LIKE $1
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`
	searchTerm := fmt.Sprintf("%%%s%%", search)
	rows, err := r.db.QueryContext(ctx, query, searchTerm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []WAFList
	for rows.Next() {
		var b WAFList
		if err := rows.Scan(
			&b.ID,
			&b.Domain,
			&b.URLPath,
			&b.ArgsPattern,
			&b.Description,
			&b.CreatedAt); err == nil {
			list = append(list, b)
		}
	}

	if list == nil {
		list = []WAFList{}
	}
	return list, nil
}

func (r *repository) NCWAF(ctx context.Context, tx TxOrDB, id, domain, path string, args *string, desc string) error {
	query := `
		INSERT INTO "ict_waf_bypass_rule" (
			id, domain, url_path, args_pattern, description, created_at
		) VALUES ($1, $2, $3, $4, $5, NOW())
	`
	_, err := tx.ExecContext(ctx, query, id, domain, path, args, desc)
	return err
}

func (r *repository) NDWAF(ctx context.Context, id string) error {
	query := `
		DELETE FROM "ict_waf_bypass_rule"
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *repository) NLATC(ctx context.Context, date string) ([]ATCList, error) {
	query := `
		SELECT	id, date, client_ip, traffic_type, target_domain, total_hits, last_seen 
	    FROM	"ict_nginx_atc_sum"
	    WHERE	date = $1 
	    ORDER BY total_hits DESC
	`
	rows, err := r.db.QueryContext(ctx, query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ATCList
	for rows.Next() {
		var s ATCList
		if err := rows.Scan(
			&s.ID,
			&s.Date,
			&s.ClientIP,
			&s.TrafficType,
			&s.TargetDomain,
			&s.TotalHits,
			&s.LastSeen); err == nil {
			list = append(list, s)
		}
	}
	if list == nil {
		list = []ATCList{}
	}
	return list, nil
}

func (r *repository) NLLOG(ctx context.Context, ip, date, table string) (*LOGList, error) {
	query := fmt.Sprintf(`
		SELECT	id, timestamp, url, status, traffic_type, country_iso, responsetime 
		FROM	%s
		WHERE	client_ip = $1 AND TO_CHAR(timestamp, 'YYYY-MM-DD') = $2 
		ORDER BY timestamp DESC
	`, table)

	rows, err := r.db.QueryContext(ctx, query, ip, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resp := &LOGList{
		ClientIP: ip,
		Date:     date,
		Logs:     []LOGItem{},
	}
	for rows.Next() {
		var ld LOGItem
		err := rows.Scan(
			&ld.ID,
			&ld.Timestamp,
			&ld.URL,
			&ld.Status,
			&ld.TrafficType,
			&ld.CountryISO,
			&ld.ResponseTime)
		if err == nil {
			resp.Logs = append(resp.Logs, ld)
		}
	}
	resp.TotalHits = len(resp.Logs)
	return resp, nil
}

func (r *repository) XGATCDate(ctx context.Context, tx TxOrDB, ip string) ([]SLAMode, error) {
	query := `
		SELECT	TO_CHAR(timestamp, 'YYYY-MM-DD') as date_str, COUNT(*) 
		FROM	"ict_nginx_atc"
		WHERE	client_ip = $1
		GROUP BY TO_CHAR(timestamp, 'YYYY-MM-DD')
	`
	rows, err := tx.QueryContext(ctx, query, ip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modifiers []SLAMode
	for rows.Next() {
		var m SLAMode
		if err := rows.Scan(
			&m.Date,
			&m.LogCount); err == nil {
			modifiers = append(modifiers, m)
		}
	}
	return modifiers, nil
}

func (r *repository) XGWAFDate(ctx context.Context, tx TxOrDB, domain, path string, args *string) ([]SLAMode, error) {
	query := `
		SELECT	TO_CHAR(timestamp, 'YYYY-MM-DD') as date_str, COUNT(*) 
		FROM	"ict_nginx_atc"
		WHERE	(domain = $1 OR $1 = '*') 
		  AND	url = $2 
		  AND	($3 IS NULL OR args LIKE '%' || $3 || '%')
		GROUP BY TO_CHAR(timestamp, 'YYYY-MM-DD')
	`
	rows, err := tx.QueryContext(ctx, query, domain, path, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modifiers []SLAMode
	for rows.Next() {
		var m SLAMode
		if err := rows.Scan(
			&m.Date,
			&m.LogCount); err == nil {
			modifiers = append(modifiers, m)
		}
	}
	return modifiers, nil
}

func (r *repository) XMLOGSIp(ctx context.Context, tx TxOrDB, ip string) (int64, error) {
	query := `
		INSERT INTO "ict_nginx_app" (
				id, timestamp, host, server_ip, client_ip, country_iso, xff, domain, url,
				referer, args, upstreamtime, responsetime, request_method, status, size,
				request_body, request_length, protocol, upstreamhost, file_dir, http_user_agent,
				traffic_type)
		SELECT	gen_random_uuid(),
				timestamp, host, server_ip, client_ip, country_iso, xff, domain, url,
				referer, args, upstreamtime, responsetime, request_method, status, size,
				request_body, request_length, protocol, upstreamhost, file_dir, http_user_agent,
				CONCAT('WH_', traffic_type)
		FROM	"ict_nginx_atc"
		WHERE	client_ip = $1
	`
	res, err := tx.ExecContext(ctx, query, ip)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *repository) XMLOGSArg(ctx context.Context, tx TxOrDB, domain, path string, args *string) (int64, error) {
	query := `
		INSERT INTO "ict_nginx_app" (
				id, timestamp, host, server_ip, client_ip, country_iso, xff, domain, url,
				referer, args, upstreamtime, responsetime, request_method, status, size,
				request_body, request_length, protocol, upstreamhost, file_dir, http_user_agent,
				traffic_type)
		SELECT	gen_random_uuid(),
				timestamp, host, server_ip, client_ip, country_iso, xff, domain, url,
				referer, args, upstreamtime, responsetime, request_method, status, size,
				request_body, request_length, protocol, upstreamhost, file_dir, http_user_agent,
				'BYPASSED_RULE_HISTORICAL'
		FROM	"ict_nginx_atc"
		WHERE	(domain = $1 OR $1 = '*') 
		  AND	url = $2 
		  AND	($3 IS NULL OR args LIKE '%' || $3 || '%')
	`
	res, err := tx.ExecContext(ctx, query, domain, path, args)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *repository) XDLOGSAtcIp(ctx context.Context, tx TxOrDB, ip string) error {
	query := `
		DELETE FROM "ict_nginx_atc"
		WHERE client_ip = $1
	`
	_, err := tx.ExecContext(ctx, query, ip)
	return err
}

func (r *repository) XDLOGSAtcArg(ctx context.Context, tx TxOrDB, domain, path string, args *string) error {
	query := `
		DELETE FROM "ict_nginx_atc"
		WHERE (domain = $1 OR $1 = '*') 
		  AND url = $2 
		  AND ($3 IS NULL OR args LIKE '%' || $3 || '%')
	`
	_, err := tx.ExecContext(ctx, query, domain, path, args)
	return err
}

func (r *repository) XDLOGSSum(ctx context.Context, tx TxOrDB, ip string) error {
	query := `
		DELETE FROM "ict_nginx_atc_sum"
		WHERE client_ip = $1
	`
	_, err := tx.ExecContext(ctx, query, ip)
	return err
}

func (r *repository) XUSLASum(ctx context.Context, tx TxOrDB, count int64, date string) error {
	query := `
		UPDATE	"ict_nginx_sla" 
		SET		attack_requests = CASE WHEN attack_requests >= $1 THEN attack_requests - $1 ELSE 0 END,
				successful_requests = successful_requests + $1
		WHERE	date = $2
	`
	_, err := tx.ExecContext(ctx, query, count, date)
	return err
}

func (r *repository) XUSLAPct(ctx context.Context, tx TxOrDB, date string) error {
	query := `
		UPDATE	"ict_nginx_sla" 
		SET		sla_percentage = CASE WHEN total_requests > 0 THEN (successful_requests::numeric / total_requests::numeric) * 100 ELSE 0.00 END 
		WHERE	date = $1
	`
	_, err := tx.ExecContext(ctx, query, date)
	return err
}

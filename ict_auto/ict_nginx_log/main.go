package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Config struct {
	PostgresConn       string
	ElasticsearchURL   string
	HttpFloodThreshold int
}

type NginxLog struct {
	Timestamp     string      `json:"@timestamp"`
	Host          string      `json:"host"`
	ServerIP      string      `json:"server_ip"`
	ClientIP      string      `json:"client_ip"`
	Xff           string      `json:"xff"`
	Domain        string      `json:"domain"`
	URL           string      `json:"url"`
	Referer       string      `json:"referer"`
	Args          string      `json:"args"`
	UpstreamTime  json.Number `json:"upstreamtime"`
	ResponseTime  json.Number `json:"responsetime"`
	RequestMethod string      `json:"request_method"`
	Status        json.Number `json:"status"`
	Size          json.Number `json:"size"`
	RequestBody   string      `json:"request_body"`
	RequestLength string      `json:"request_length"`
	Protocol      string      `json:"protocol"`
	UpstreamHost  string      `json:"upstreamhost"`
	FileDir       string      `json:"file_dir"`
	UserAgent     string      `json:"http_user_agent"`
	GeoIP         *struct {
		Geo *struct {
			CountryISOCode string `json:"country_iso_code"`
		} `json:"geo"`
	} `json:"geoip"`
}

type EsHit struct {
	ID     string   `json:"_id"`
	Source NginxLog `json:"_source"`
}

type EsSearchResponse struct {
	Hits struct {
		Hits []EsHit `json:"hits"`
	} `json:"hits"`
}

type EsCountResponse struct {
	Count int `json:"count"`
}

type AttackKey struct {
	ClientIP    string
	TrafficType string
	Domain      string
}

type IPTracker struct {
	Mu    sync.RWMutex
	Store map[string]int
}

func LoadConfig() *Config {
	godotenv.Load()
	thresh := 100
	if envThresh := os.Getenv("FT_HTTP"); envThresh != "" {
		if val, err := strconv.Atoi(envThresh); err == nil {
			thresh = val
		}
	}
	PG_Host := os.Getenv("PG_HOST")
	PG_Port := os.Getenv("PG_PORT")
	PG_User := os.Getenv("PG_USER")
	PG_Pass := os.Getenv("PG_PASS")
	PG_Data := os.Getenv("PG_DATA")
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		PG_User, PG_Pass, PG_Host, PG_Port, PG_Data)
	return &Config{
		PostgresConn:       dsn,
		ElasticsearchURL:   os.Getenv("ES_LINK"),
		HttpFloodThreshold: thresh,
	}
}

var PgSQL *sql.DB

func InitDB(connStr string) {
	var err error
	PgSQL, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Gagal koneksi ke Basis data: %v", err)
	}

	if err = PgSQL.Ping(); err != nil {
		log.Fatalf("Basis data tidak merespon Ping: %v", err)
	}
}

func initElasticClient() (*elasticsearch.Client, error) {
	elasticAddr := os.Getenv("ES_LINK")
	if elasticAddr == "" {
		elasticAddr = "http://elasticsearch:9200"
	}
	customTransport := &http.Transport{
		DisableKeepAlives:   false,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,

		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	cfg := elasticsearch.Config{
		Addresses: []string{elasticAddr},
		Transport: customTransport,
	}

	return elasticsearch.NewClient(cfg)
}

var (
	RegexSQLI           = regexp.MustCompile(`(?i)(UNION|SELECT|INSERT|UPDATE|DELETE|DROP|ALTER|WHERE|OR\s+\d=\d|--|#|\/\*|AND\s+\d=\d|ORDER\s+BY|GROUP\s+BY)`)
	RegexXSS            = regexp.MustCompile(`(?i)(<script|javascript:|onerror=|onload=|alert\(|%3Cscript|<svg|<iframe|confirm\(|prompt\()`)
	RegexTraversal      = regexp.MustCompile(`(?i)(\.\.\/|\.\.\\|%2e%2e%2f|etc\/passwd|boot\.ini|win\.ini|%5c\.\.)`)
	RegexRCE            = regexp.MustCompile(`(?i)(bin\/sh|bin\/bash|cmd\.exe|powershell|wget\s|curl\s|eval\(|passthru|shell_exec|system\(|popen\()`)
	RegexSensitiveFiles = regexp.MustCompile(`(?i)(\.env|\.git|\.docker|config\.php|wp-config\.php|db\.php|\.bak|\.sql|\.yaml)`)
	RegexBotScanner     = regexp.MustCompile(`(?i)(nikto|sqlmap|dirbuster|w3af|nmap|acunetix|masscan|python-requests|curl|hydra|gobuster|wfuzz|amass|zgrab)`)

	TrackerInstance = &IPTracker{Store: make(map[string]int)}
)

func IsWhitelistedDB(db *sql.DB, ipStr string) bool {
	clientIP := net.ParseIP(strings.TrimSpace(ipStr))
	if clientIP == nil {
		return false
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM ict_ip_whitelist WHERE ip_or_cidr = $1)", ipStr).Scan(&exists)
	if err == nil && exists {
		return true
	}

	rows, err := db.Query("SELECT ip_or_cidr FROM ict_ip_whitelist WHERE ip_or_cidr LIKE '%/%'")
	if err != nil {
		log.Printf("Gagal membaca ict_ip_whitelist : %v", err)
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cidr string
		if err := rows.Scan(&cidr); err == nil {
			_, ipNet, err := net.ParseCIDR(cidr)
			if err == nil && ipNet.Contains(clientIP) {
				return true
			}
		}
	}
	return false
}

func IsBannedDB(db *sql.DB, ipStr string) bool {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM "ict_ip_blacklist"
			WHERE ip = $1 AND (expires_at IS NULL OR expires_at > NOW())
		)`
	err := db.QueryRow(query, ipStr).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func ClassifyTraffic(url, args, body, ua string) string {
	payload := strings.ToLower(url + " " + args + " " + body)
	uaLower := strings.ToLower(ua)

	if RegexBotScanner.MatchString(uaLower) {
		return "BOT_SCANNER"
	}
	if RegexSQLI.MatchString(payload) {
		return "SQL_INJECTION"
	}
	if RegexXSS.MatchString(payload) {
		return "XSS"
	}
	if RegexTraversal.MatchString(payload) {
		return "PATH_TRAVERSAL"
	}
	if RegexRCE.MatchString(payload) {
		return "RCE_COMMAND_INJECTION"
	}
	if RegexSensitiveFiles.MatchString(payload) {
		return "SENSITIVE_FILE_PROBING"
	}
	return "NORMAL"
}

func GetThreatWeight(trafficType string) int {
	switch trafficType {
	case "RCE_COMMAND_INJECTION":
		return 50
	case "SQL_INJECTION":
		return 40
	case "PATH_TRAVERSAL":
		return 35
	case "XSS":
		return 20
	case "SENSITIVE_FILE_PROBING":
		return 15
	case "BOT_SCANNER":
		return 10
	default:
		return 0
	}
}

func UpdateThreatScoreDB(db *sql.DB, ip string, trafficType string) (int, bool) {
	if IsWhitelistedDB(db, ip) {
		return 0, false
	}
	if IsBannedDB(db, ip) {
		return 150, true
	}

	weight := GetThreatWeight(trafficType)
	if weight == 0 {
		TrackerInstance.Mu.RLock()
		currentScore := TrackerInstance.Store[ip]
		TrackerInstance.Mu.RUnlock()
		return currentScore, currentScore >= 100
	}

	TrackerInstance.Mu.Lock()
	TrackerInstance.Store[ip] += weight
	finalScore := TrackerInstance.Store[ip]
	if finalScore > 150 {
		TrackerInstance.Store[ip] = 150
		finalScore = 150
	}
	TrackerInstance.Mu.Unlock()

	if finalScore >= 100 {
		expiryTime := time.Now().Add(24 * time.Hour)

		query := `
			INSERT INTO "ict_ip_blacklist" (
				id, ip, threat_score, reason, expires_at
			) VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (ip) DO UPDATE 
			 SET threat_score = EXCLUDED.threat_score, expires_at = EXCLUDED.expires_at;
		`
		newUUID := uuid.NewString()
		_, err := db.Exec(query, newUUID, ip, finalScore, trafficType, expiryTime)
		if err != nil {
			log.Printf("Gagal menyimpan blokir IP %s ke database: %v", ip, err)
		}
		return finalScore, true
	}

	return finalScore, false
}

func StartSyncWorker(es *elasticsearch.Client, threshold int) {
	runSyncTask(es, threshold)

	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		runSyncTask(es, threshold)
	}
}

func runSyncTask(es *elasticsearch.Client, threshold int) {
	ctx := context.Background()
	now := time.Now()
	currentIndex := fmt.Sprintf("logstash_%s", now.Format("2006.01.02"))

	log.Printf("[%s] Sinkronisasi ...", now.Format("15:04:05"))
	syncAndAnalyze(ctx, es, currentIndex, now, threshold)
	cleanupOldIndices(ctx, es, now)
	runtime.GC()
	debug.FreeOSMemory()
}

func IsBypassedRuleDB(db *sql.DB, domain, urlPath, args string) bool {
	query := `
		SELECT	COUNT(*)
		FROM	"ict_waf_bypass_rule"
		WHERE	(domain = $1 OR domain = '*')
		  AND	url_path = $2
		  AND	($3 = '' OR args_pattern IS NULL OR args LIKE '%' || args_pattern || '%')
	`
	var count int
	err := db.QueryRow(query, domain, urlPath, args).Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}

func syncAndAnalyze(ctx context.Context, es *elasticsearch.Client, indexName string, now time.Time, threshold int) {
	cleanJsonQuery := `{"query": {"bool": {"must": {"match": {"error.message": "Error decoding JSON: invalid character 'x' in string escape code"}}}}}`
	resCleanJson, errCleanJson := es.DeleteByQuery(
		[]string{indexName},
		strings.NewReader(cleanJsonQuery),
		es.DeleteByQuery.WithContext(ctx),
		es.DeleteByQuery.WithConflicts("proceed"),
		es.DeleteByQuery.WithScrollSize(2500),
	)
	if errCleanJson == nil && resCleanJson != nil {
		resCleanJson.Body.Close()
		_, _ = io.Copy(io.Discard, resCleanJson.Body)
		resCleanJson.Body.Close()
	}

	cleaStatus0 := `{"query": {"match": {"status": "0"}}}`
	resClean0, errClean0 := es.DeleteByQuery(
		[]string{indexName},
		strings.NewReader(cleaStatus0),
		es.DeleteByQuery.WithContext(ctx),
		es.DeleteByQuery.WithConflicts("proceed"),
		es.DeleteByQuery.WithScrollSize(2500),
	)
	if errClean0 == nil && resClean0 != nil {
		resClean0.Body.Close()
		_, _ = io.Copy(io.Discard, resClean0.Body)
		resClean0.Body.Close()
	}

	cleanIpQuery := `{"query": {"bool": {"should": [{"term": {"client_ip": ""}}, {"bool": {"must_not": {"exists": {"field": "client_ip"}}}}], "minimum_should_match": 1}}}`
	resCleanIp, errCleanIp := es.DeleteByQuery(
		[]string{indexName},
		strings.NewReader(cleanIpQuery),
		es.DeleteByQuery.WithContext(ctx),
		es.DeleteByQuery.WithConflicts("proceed"),
		es.DeleteByQuery.WithScrollSize(2500),
	)
	if errCleanIp == nil && resCleanIp != nil {
		_, _ = io.Copy(io.Discard, resCleanIp.Body)
		resCleanIp.Body.Close()
	}

	queryFindDates := `
		SELECT DISTINCT date_str FROM (
			SELECT DISTINCT TO_CHAR(timestamp, 'YYYY-MM-DD') AS date_str FROM ict_nginx_log WHERE client_ip = ''
			UNION
			SELECT DISTINCT TO_CHAR(timestamp, 'YYYY-MM-DD') AS date_str FROM ict_nginx_app WHERE client_ip = ''
			UNION
			SELECT DISTINCT TO_CHAR(timestamp, 'YYYY-MM-DD') AS date_str FROM ict_nginx_atc WHERE client_ip = ''
		) t WHERE date_str IS NOT NULL`

	rowsDates, errDates := PgSQL.QueryContext(ctx, queryFindDates)
	if errDates == nil {
		var affectedDates []string
		for rowsDates.Next() {
			var d string
			if err := rowsDates.Scan(&d); err == nil {
				affectedDates = append(affectedDates, d)
			}
		}
		rowsDates.Close()

		if len(affectedDates) > 0 {
			log.Printf("[%d] Memulai pembersihan data client_ip kosong ...", len(affectedDates))

			txClean, errTxClean := PgSQL.BeginTx(ctx, nil)
			if errTxClean == nil {
				cleanSuccess := true
				for _, dateStr := range affectedDates {
					_, _ = txClean.ExecContext(ctx, "SAVEPOINT clean_date_save")

					_, err1 := txClean.ExecContext(ctx, "DELETE FROM ict_nginx_log WHERE client_ip = '' AND TO_CHAR(timestamp, 'YYYY-MM-DD') = $1", dateStr)
					_, err2 := txClean.ExecContext(ctx, "DELETE FROM ict_nginx_app WHERE client_ip = '' AND TO_CHAR(timestamp, 'YYYY-MM-DD') = $1", dateStr)
					_, err3 := txClean.ExecContext(ctx, "DELETE FROM ict_nginx_atc WHERE client_ip = '' AND TO_CHAR(timestamp, 'YYYY-MM-DD') = $1", dateStr)
					_, err4 := txClean.ExecContext(ctx, "DELETE FROM ict_nginx_atc_sum WHERE client_ip = '' AND date = $1", dateStr)

					if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
						_, _ = txClean.ExecContext(ctx, "ROLLBACK TO SAVEPOINT clean_date_save")
						continue
					}

					queryUpdateSLA := `
						WITH metrics AS (
							SELECT 
								COUNT(*) AS total,
								COUNT(CASE WHEN (status::int >= 200 AND status::int < 300) AND (traffic_type = 'NORMAL' OR traffic_type = 'WHITELISTED_TRAFFIC') THEN 1 END) AS success,
								COUNT(CASE WHEN status::int >= 400 AND status::int < 500 THEN 1 END) AS client_err,
								COUNT(CASE WHEN status::int >= 500 THEN 1 END) AS server_err,
								COALESCE(AVG(NULLIF(responsetime, '')::numeric), 0.0000) AS avg_time
							FROM (
								SELECT timestamp, status, traffic_type, responsetime FROM ict_nginx_log WHERE TO_CHAR(timestamp, 'YYYY-MM-DD') = $1
								UNION ALL
								SELECT timestamp, status, traffic_type, responsetime FROM ict_nginx_app WHERE TO_CHAR(timestamp, 'YYYY-MM-DD') = $1
								UNION ALL
								SELECT timestamp, status, traffic_type, responsetime FROM ict_nginx_atc WHERE TO_CHAR(timestamp, 'YYYY-MM-DD') = $1
							) combined
						),
						attack_metrics AS (
							SELECT COUNT(*) AS total_attacks FROM ict_nginx_atc WHERE TO_CHAR(timestamp, 'YYYY-MM-DD') = $1
						)
						UPDATE ict_nginx_sla s
						SET 
							total_requests = m.total,
							successful_requests = m.success,
							client_errors = m.client_err,
							server_errors = m.server_err,
							attack_requests = a.total_attacks,
							avg_response_time = m.avg_time,
							sla_percentage = CASE WHEN m.total > 0 THEN (m.success::numeric / m.total::numeric) * 100 ELSE 0.00 END
						FROM metrics m, attack_metrics a
						WHERE s.date = $1
					`
					if _, errSLA := txClean.ExecContext(ctx, queryUpdateSLA, dateStr); errSLA != nil {
						_, _ = txClean.ExecContext(ctx, "ROLLBACK TO SAVEPOINT clean_date_save")
						cleanSuccess = false
						break
					}
					_, _ = txClean.ExecContext(ctx, "RELEASE SAVEPOINT clean_date_save")
				}

				if cleanSuccess {
					_ = txClean.Commit()
					log.Println("SLA diperbarui.")
				} else {
					txClean.Rollback()
				}
			}
		}
	}

	query := `{"size": 1000, "query": {"match_all": {}}}`
	res, err := es.Search(es.Search.WithContext(ctx), es.Search.WithIndex(indexName), es.Search.WithBody(strings.NewReader(query)))
	if err != nil {
		log.Print("Log Sudah habis")
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("Balikan Gagal Dari Elasticsearch : %s", res.Status())
		return
	}

	var searchResult EsSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		log.Printf("Gagal decode hasil pencarian : %v", err)
		return
	}
	if len(searchResult.Hits.Hits) == 0 {
		return
	}

	TrackerInstance.Mu.Lock()
	TrackerInstance.Store = make(map[string]int)
	for _, hit := range searchResult.Hits.Hits {
		TrackerInstance.Store[hit.Source.ClientIP]++
	}
	TrackerInstance.Mu.Unlock()

	var batchTotal, batchSuccess, batchClientErr, batchServerErr, batchAttack int64
	var totalResponseTime float64

	batchAttackSummary := make(map[AttackKey]int64)
	var deletedDocIDs []string

	tx, err := PgSQL.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Gagal memulai transaksi: %v", err)
		return
	}

	txOpened := true
	defer func() {
		if txOpened {
			tx.Rollback()
		}
	}()

	queryInsertNormal := `
		INSERT INTO "ict_nginx_log" (
			id, timestamp, host, server_ip, client_ip, country_iso, xff, domain, url, referer, args, upstreamtime, responsetime,
			request_method, status, size, request_body, request_length, protocol, upstreamhost, file_dir, http_user_agent, traffic_type
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`
	stmtNormal, err := tx.PrepareContext(ctx, queryInsertNormal)
	if err != nil {
		log.Printf("Gagal Mencatat Riwayat Normal : %v", err)
		return
	}
	defer stmtNormal.Close()

	queryInsertAttack := `
		INSERT INTO "ict_nginx_atc" (
			id, timestamp, host, server_ip, client_ip, country_iso, xff, domain, url, referer, args, upstreamtime, responsetime,
			request_method, status, size, request_body, request_length, protocol, upstreamhost, file_dir, http_user_agent, traffic_type
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`
	stmtAttack, err := tx.PrepareContext(ctx, queryInsertAttack)
	if err != nil {
		log.Printf("Gagal Mencatat Riwayat Serangan : %v", err)
		return
	}
	defer stmtAttack.Close()

	queryInsertApp := `
		INSERT INTO "ict_nginx_app" (
			id, timestamp, host, server_ip, client_ip, country_iso, xff, domain, url, referer, args, upstreamtime, responsetime,
			request_method, status, size, request_body, request_length, protocol, upstreamhost, file_dir, http_user_agent, traffic_type
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`
	stmtApp, err := tx.PrepareContext(ctx, queryInsertApp)
	if err != nil {
		log.Printf("Gagal Mencatat Riwayat Whitelist : %v", err)
		return
	}
	defer stmtApp.Close()

	for _, hit := range searchResult.Hits.Hits {
		logData := hit.Source

		statusInt64, _ := logData.Status.Int64()
		statusInt := int(statusInt64)

		statusStr := logData.Status.String()
		responseTimeStr := logData.ResponseTime.String()
		upstreamTimeStr := logData.UpstreamTime.String()
		sizeStr := logData.Size.String()

		countryIso := "-"
		if logData.GeoIP != nil && logData.GeoIP.Geo != nil && logData.GeoIP.Geo.CountryISOCode != "" {
			countryIso = logData.GeoIP.Geo.CountryISOCode
		}

		reqLenInt, err_int := strconv.ParseInt(logData.RequestLength, 10, 64)
		var requestLengthParam interface{}
		if err_int != nil || reqLenInt <= 0 {
			requestLengthParam = sql.NullInt64{Valid: false}
		} else {
			requestLengthParam = reqLenInt
		}

		respTimeFloat, _ := logData.ResponseTime.Float64()
		totalResponseTime += respTimeFloat

		TrackerInstance.Mu.RLock()
		totalHitByIP := TrackerInstance.Store[logData.ClientIP]
		TrackerInstance.Mu.RUnlock()

		var trafficType string
		if totalHitByIP > threshold {
			trafficType = "HTTP_FLOOD"
		} else {
			trafficType = ClassifyTraffic(logData.URL, logData.Args, logData.RequestBody, logData.UserAgent)
		}

		batchTotal++

		var targetStmt *sql.Stmt

		isClientWhitelisted := IsWhitelistedDB(PgSQL, logData.ClientIP)

		isRuleBypassed := IsBypassedRuleDB(PgSQL, logData.Domain, logData.URL, logData.Args)

		if isClientWhitelisted {

			targetStmt = stmtApp
			trafficType = "WH_" + trafficType

			if statusInt >= 200 && statusInt < 300 {
				batchSuccess++
			} else if statusInt >= 400 && statusInt < 500 {
				batchClientErr++
			} else if statusInt >= 500 {
				batchServerErr++
			}

		} else if isRuleBypassed || trafficType == "NORMAL" {

			targetStmt = stmtNormal

			if isRuleBypassed {
				trafficType = "NORMAL_RULE"
			}

			if statusInt >= 200 && statusInt < 300 {
				batchSuccess++
			} else if statusInt >= 400 && statusInt < 500 {
				batchClientErr++
			} else if statusInt >= 500 {
				batchServerErr++
			}

		} else {

			targetStmt = stmtAttack
			batchAttack++

			if statusInt >= 400 && statusInt < 500 {
				batchClientErr++
			} else if statusInt >= 500 {
				batchServerErr++
			}

			key := AttackKey{
				ClientIP:    logData.ClientIP,
				TrafficType: trafficType,
				Domain:      logData.Domain,
			}
			batchAttackSummary[key]++

			UpdateThreatScoreDB(PgSQL, logData.ClientIP, trafficType)
		}

		rowUUID := uuid.NewString()
		_, _ = tx.ExecContext(ctx, "SAVEPOINT log_insert_save")
		_, err := targetStmt.ExecContext(ctx,
			rowUUID,
			logData.Timestamp, logData.Host, logData.ServerIP, logData.ClientIP, countryIso,
			logData.Xff, logData.Domain, logData.URL, logData.Referer, logData.Args,
			upstreamTimeStr, responseTimeStr, logData.RequestMethod, statusStr, sizeStr,
			logData.RequestBody, requestLengthParam, logData.Protocol, logData.UpstreamHost,
			logData.FileDir, logData.UserAgent, trafficType,
		)
		if err != nil {
			if (logData.ClientIP != "" && logData.ClientIP != "-") || (logData.Domain != "" && logData.Domain != "-") {
				log.Printf("Gagal Mencatat data IP %s : %v", logData.ClientIP, err)
			}
			_, _ = tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT log_insert_save")
			continue
		}

		_, _ = tx.ExecContext(ctx, "RELEASE SAVEPOINT log_insert_save")
		deletedDocIDs = append(deletedDocIDs, hit.ID)
	}

	todayStr := now.Format("2006-01-02")

	if len(batchAttackSummary) > 0 {
		stmtSummary, err := tx.PrepareContext(ctx, `
		INSERT INTO ict_nginx_atc_sum (
			id, date, client_ip, traffic_type, target_domain, total_hits, last_seen
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
		 ON CONFLICT (date, client_ip, traffic_type, target_domain)
		 DO UPDATE SET
		 total_hits = ict_nginx_atc_sum.total_hits + EXCLUDED.total_hits, last_seen = NOW();
		`)
		if err == nil {
			for k, hits := range batchAttackSummary {
				_, _ = tx.ExecContext(ctx, "SAVEPOINT atc_sum_save")

				sumUUID := uuid.NewString()
				_, errExec := stmtSummary.ExecContext(ctx, sumUUID, todayStr, k.ClientIP, k.TrafficType, k.Domain, hits)
				if errExec != nil {
					log.Printf("Gagal Memperbarui ringkasan untuk IP %s (Dilewati): %v", k.ClientIP, errExec)
					_, _ = tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT atc_sum_save")
					continue
				}
				_, _ = tx.ExecContext(ctx, "RELEASE SAVEPOINT atc_sum_save")
			}
			stmtSummary.Close()
		}
	}
	if batchTotal > 0 {
		avgTime := totalResponseTime / float64(batchTotal)

		slaUUID := uuid.NewString()
		_, err = tx.ExecContext(ctx, `
		INSERT INTO ict_nginx_sla (
			id, date, total_requests, successful_requests, client_errors, server_errors, attack_requests,
			avg_response_time, sla_percentage
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0.00)
		ON CONFLICT (date) DO UPDATE SET
			total_requests = ict_nginx_sla.total_requests + EXCLUDED.total_requests,
			successful_requests = ict_nginx_sla.successful_requests + EXCLUDED.successful_requests,
			client_errors = ict_nginx_sla.client_errors + EXCLUDED.client_errors,
			server_errors = ict_nginx_sla.server_errors + EXCLUDED.server_errors,
			attack_requests = ict_nginx_sla.attack_requests + EXCLUDED.attack_requests,
			avg_response_time = ((ict_nginx_sla.avg_response_time * ict_nginx_sla.total_requests) + $9) / (ict_nginx_sla.total_requests + EXCLUDED.total_requests);
		`, slaUUID, todayStr, batchTotal, batchSuccess, batchClientErr, batchServerErr, batchAttack, avgTime, totalResponseTime) // Geser indeks parameter totalResponseTime menjadi $9

		if err != nil {
			log.Printf("Gagal mengolah SLA: %v", err)
			return
		}
		_, _ = tx.ExecContext(ctx, `
		UPDATE ict_nginx_sla SET sla_percentage = CASE WHEN total_requests > 0 THEN (successful_requests::numeric / total_requests::numeric) * 100 ELSE 0 END WHERE date = $1
		`, todayStr)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Gagal Memproses Transaksi : %v", err)
		return
	}
	txOpened = false

	for _, id := range deletedDocIDs {
		_, _ = es.Delete(indexName, id, es.Delete.WithContext(ctx))
	}
	if len(deletedDocIDs) > 0 {
		log.Printf("[%s] Berhasil sinkronisasi %d log dari ES ke PostgreSQL.", indexName, len(deletedDocIDs))
	}
}

func cleanupOldIndices(ctx context.Context, es *elasticsearch.Client, now time.Time) {
	for i := 1; i <= 3; i++ {
		oldDate := now.AddDate(0, 0, -i)
		oldIndexName := fmt.Sprintf("logstash_%s", oldDate.Format("2006.01.02"))
		resCount, err := es.Count(es.Count.WithContext(ctx), es.Count.WithIndex(oldIndexName))
		if err != nil {
			continue
		}
		defer resCount.Body.Close()
		if resCount.StatusCode == http.StatusNotFound {
			continue
		}
		var countResult EsCountResponse
		if err := json.NewDecoder(resCount.Body).Decode(&countResult); err == nil {
			if countResult.Count > 0 {
				query := `{"query": {"match_all": {}}}`
				_, _ = es.DeleteByQuery([]string{oldIndexName}, strings.NewReader(query), es.DeleteByQuery.WithContext(ctx))
			}
			_, _ = es.Indices.Delete([]string{oldIndexName}, es.Indices.Delete.WithContext(ctx))
		}
	}
}

func main() {
	cfg := LoadConfig()

	InitDB(cfg.PostgresConn)
	defer PgSQL.Close()

	esClient, err := initElasticClient()
	if err != nil {
		log.Fatalf("Gagal Memetakan data Elasticsearch : %v", err)
	}

	res, err := esClient.Info()
	if err != nil {
		log.Fatalf("Gagal berkomunikasi dengan cluster Elasticsearch: %v", err)
	}
	res.Body.Close()

	log.Println("Sistem Sinkronisasi Log Nginx Berjalan...")
	StartSyncWorker(esClient, cfg.HttpFloodThreshold)
}

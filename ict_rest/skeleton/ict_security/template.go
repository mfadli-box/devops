package ict_security

import (
	"context"
	"database/sql"
	"time"
)

/* ======================= ict_nginx_sla
  id                       String            @id @default(uuid())
  date                     String            @unique @db.VarChar(10)
  total_requests           BigInt            @default(0)
  successful_requests      BigInt            @default(0)
  client_errors            BigInt            @default(0)
  server_errors            BigInt            @default(0)
  attack_requests          BigInt            @default(0)
  avg_response_time        Decimal           @db.Decimal(10, 4) @default(0.0000)
  sla_percentage           Decimal           @db.Decimal(5, 2) @default(0.00)
========================== */

type SLAInfo struct {
	ID                 string  `json:"id"`
	Date               string  `json:"date"`
	TotalRequests      int64   `json:"total_requests"`
	SuccessfulRequests int64   `json:"successful_requests"`
	ClientErrors       int64   `json:"client_errors"`
	ServerErrors       int64   `json:"server_errors"`
	AttackRequests     int64   `json:"attack_requests"`
	AvgResponseTime    float64 `json:"avg_response_time"`
	SLAPercentage      float64 `json:"sla_percentage"`
}

type SLAMode struct {
	Date     string
	LogCount int64
}

/* ======================= ict_ip_whitelist
  id                       String            @id @default(uuid())
  ip_or_cidr               String            @unique @db.VarChar(50)
  description              String?           @db.Text
  created_at               DateTime          @default(now())
========================== */

type IPWInfo struct {
	ID          string    `json:"id"`
	IPOrCIDR    string    `json:"ip_or_cidr"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type IPWItem struct {
	IP          string `json:"ip"`
	Description string `json:"description"`
}

/* ======================= ict_ip_blacklist
  id                       String            @id @default(uuid())
  ip                       String            @unique @db.VarChar(45)
  threat_score             Int
  reason                   String            @db.VarChar(50)
  banned_at                DateTime          @default(now())
  expires_at               DateTime?
========================== */

type IPBInfo struct {
	ID          string     `json:"id"`
	IP          string     `json:"ip"`
	ThreatScore int        `json:"threat_score"`
	Reason      string     `json:"reason"`
	BannedAt    time.Time  `json:"banned_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

/* ======================= ict_nginx_atc_sum
  id                       String            @id @default(uuid())
  date                     String            @db.VarChar(10)
  client_ip                String            @db.VarChar(45)
  traffic_type             String            @db.VarChar(50)
  target_domain            String            @db.VarChar(255)
  total_hits               BigInt            @default(1)
  last_seen                DateTime          @default(now())

  @@unique([date, client_ip, traffic_type, target_domain], name: "uq_atc_sum_composite")
========================== */

type ATCList struct {
	ID           string    `json:"id"`
	Date         string    `json:"date"`
	ClientIP     string    `json:"client_ip"`
	TrafficType  string    `json:"traffic_type"`
	TargetDomain string    `json:"target_domain"`
	TotalHits    int64     `json:"total_hits"`
	LastSeen     time.Time `json:"last_seen"`
}

/* ======================= ict_waf_bypass_rule
  id                       String            @id @default(uuid())
  domain                   String            @db.VarChar(255)
  url_path                 String            @db.Text
  args_pattern             String?           @db.Text
  description              String?           @db.Text
  created_at               DateTime          @default(now())

  @@index([domain, url_path])
========================== */

type WAFList struct {
	ID          string    `json:"id"`
	Domain      string    `json:"domain"`
	URLPath     string    `json:"url_path"`
	ArgsPattern *string   `json:"args_pattern,omitempty"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type WAFItem struct {
	Domain      string  `json:"domain" binding:"required"`
	URLPath     string  `json:"url_path" binding:"required"`
	ArgsPattern *string `json:"args_pattern"`
	Description string  `json:"description"`
}

/* ======================= ict_nginx_log
   ======================= ict_nginx_app
   ======================= ict_nginx_atc
  id                       String            @id @default(uuid())
  timestamp                DateTime          @default(now())
  host                     String            @db.VarChar(255)
  server_ip                String            @db.VarChar(45)
  client_ip                String            @db.VarChar(45)
  country_iso              String?           @db.VarChar(5)
  xff                      String?           @db.Text
  traffic_type             String            @db.VarChar(50)
  domain                   String            @db.VarChar(255)
  url                      String            @db.Text
  referer                  String?           @db.Text
  args                     String?           @db.Text
  upstreamtime             String?           @db.VarChar(50)
  responsetime             String?           @db.VarChar(50)
  request_method           String            @db.VarChar(10)
  status                   String            @db.VarChar(5)
  size                     String?           @db.VarChar(50)
  request_body             String?           @db.Text
  request_length           Int?
  protocol                 String?           @db.VarChar(20)
  upstreamhost             String?           @db.VarChar(255)
  file_dir                 String?           @db.VarChar(255)
  http_user_agent          String?           @db.Text
  created_at               DateTime          @default(now())

  @@index([client_ip])
  @@index([country_iso])
  @@index([timestamp])
========================== */

type LOGList struct {
	ClientIP    string    `json:"client_ip"`
	Date        string    `json:"date"`
	TotalHits   int       `json:"total_hits"`
	TrafficType string    `json:"traffic_type_dominant,omitempty"`
	Logs        []LOGItem `json:"logs"`
}

type LOGItem struct {
	ID           string    `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	Domain       string    `json:"domain"`
	URL          string    `json:"url"`
	Status       string    `json:"status"`
	TrafficType  string    `json:"traffic_type"`
	CountryISO   string    `json:"country_iso"`
	Args         string    `json:"args,omitempty"`
	ResponseTime string    `json:"response_time"`
}

type LOGInfo struct {
	ID            string    `json:"id"`
	Timestamp     string    `json:"timestamp"`
	Host          string    `json:"host"`
	ServerIP      string    `json:"server_ip"`
	ClientIP      string    `json:"client_ip"`
	CountryISO    string    `json:"country_iso,omitempty"`
	XFF           string    `json:"xff,omitempty"`
	TrafficType   string    `json:"traffic_type"`
	Domain        string    `json:"domain"`
	URL           string    `json:"url"`
	Referer       string    `json:"referer,omitempty"`
	Args          string    `json:"args,omitempty"`
	UpstreamTime  string    `json:"upstreamtime,omitempty"`
	ResponseTime  string    `json:"responsetime,omitempty"`
	RequestMethod string    `json:"request_method"`
	Status        string    `json:"status"`
	Size          string    `json:"size,omitempty"`
	RequestBody   string    `json:"request_body,omitempty"`
	RequestLength int       `json:"request_length,omitempty"`
	Protocol      string    `json:"protocol,omitempty"`
	UpstreamHost  string    `json:"upstreamhost,omitempty"`
	FileDir       string    `json:"file_dir,omitempty"`
	HTTPUserAgent string    `json:"http_user_agent,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type Repository interface {
	NLSLA(ctx context.Context, limit, offset int) ([]SLAInfo, error)
	NLIPW(ctx context.Context, search string, limit, offset int) ([]IPWInfo, error)
	NCIPW(ctx context.Context, tx TxOrDB, id, ip, desc string) error
	NDIPW(ctx context.Context, ip string) error
	NLIPB(ctx context.Context, search string, limit, offset int) ([]IPBInfo, error)
	NDIPB(ctx context.Context, tx TxOrDB, ip string) error
	NLWAF(ctx context.Context, search string, limit, offset int) ([]WAFList, error)
	NCWAF(ctx context.Context, tx TxOrDB, id, domain, path string, args *string, desc string) error
	NDWAF(ctx context.Context, id string) error
	NLATC(ctx context.Context, date string) ([]ATCList, error)
	NLLOG(ctx context.Context, ip, date, table string) (*LOGList, error)
	XGATCDate(ctx context.Context, tx TxOrDB, ip string) ([]SLAMode, error)
	XGWAFDate(ctx context.Context, tx TxOrDB, domain, path string, args *string) ([]SLAMode, error)
	XMLOGSIp(ctx context.Context, tx TxOrDB, ip string) (int64, error)
	XMLOGSArg(ctx context.Context, tx TxOrDB, domain, path string, args *string) (int64, error)
	XDLOGSAtcIp(ctx context.Context, tx TxOrDB, ip string) error
	XDLOGSAtcArg(ctx context.Context, tx TxOrDB, domain, path string, args *string) error
	XDLOGSSum(ctx context.Context, tx TxOrDB, ip string) error
	XUSLASum(ctx context.Context, tx TxOrDB, count int64, date string) error
	XUSLAPct(ctx context.Context, tx TxOrDB, date string) error
}

type UseCase interface {
	NLSLA(ctx context.Context, limit, offset int) ([]SLAInfo, error)
	NLIPW(ctx context.Context, search string, limit, offset int) ([]IPWInfo, error)
	NDIPW(ctx context.Context, id string) error
	NLIPB(ctx context.Context, search string, limit, offset int) ([]IPBInfo, error)
	NCIPM(ctx context.Context, ip, desc string) error
	NLWAF(ctx context.Context, search string, limit, offset int) ([]WAFList, error)
	NCWAF(ctx context.Context, req WAFItem) error
	NDWAF(ctx context.Context, id string) error
	NLATC(ctx context.Context, date string) ([]ATCList, error)
	NLLOG(ctx context.Context, ip, date string) (*LOGList, error)
}

type TxOrDB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type Result interface {
	RowsAffected() (int64, error)
}

type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
}

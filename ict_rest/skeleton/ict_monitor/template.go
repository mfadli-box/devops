package ict_monitor

import (
	"context"
	"time"
)

/* ======================= ict_uptimerobot_log
  id                       String            @id @default(uuid())
  monitor_id               Int
  monitor_url              String
  monitor_friendly_name    String
  alert_type               Int
  alert_type_friendly_name String
  alert_details            String?
  alert_duration           Int?
  alert_datetime           DateTime
  monitor_alert_contacts   String
  ssl_expiry_date          DateTime?
  ssl_expiry_days_left     Int?
  dashboard_url            String
  monitor_type             String
  http_status_code         Int?
  monitoring_regions       Json
  monitor_tags             Json
  monitor_group            String?
  incident_start_time      DateTime
  incident_end_time        DateTime?
  response_time            Int?
  created_at               DateTime          @default(now())

  @@index([monitor_id])
  @@index([alert_datetime])
========================== */

type UptimeAlertInfo struct {
	ID                    string     `json:"id"`
	MonitorID             int        `json:"monitor_id"`
	MonitorURL            string     `json:"monitor_url"`
	MonitorFriendlyName   string     `json:"monitor_friendly_name"`
	AlertType             int        `json:"alert_type"`
	AlertTypeFriendlyName string     `json:"alert_type_friendly_name"`
	AlertDetails          *string    `json:"alert_details"`
	AlertDuration         *int       `json:"alert_duration"`
	AlertDateTime         time.Time  `json:"alert_date_time"`
	MonitorAlertContacts  string     `json:"monitor_alert_contacts"`
	SSLExpiryDate         *time.Time `json:"ssl_expiry_date"`
	SSLExpiryDaysLeft     *int       `json:"ssl_expiry_days_left"`
	DashboardURL          string     `json:"dashboard_url"`
	MonitorType           string     `json:"monitor_type"`
	HTTPStatusCode        *int       `json:"http_status_code"`
	MonitoringRegions     []byte     `json:"monitoring_regions"`
	MonitorTags           []byte     `json:"monitor_tags"`
	MonitorGroup          *string    `json:"monitor_group"`
	IncidentStartTime     time.Time  `json:"incident_start_time"`
	IncidentEndTime       *time.Time `json:"incident_end_time"`
	ResponseTime          *int       `json:"response_time"`
	CreatedAt             time.Time  `json:"created_at"`
}

type UptimeAlertItem struct {
	MonitorID             int    `form:"monitorid" json:"monitorid"`
	MonitorURL            string `form:"monitorurl" json:"monitorurl"`
	MonitorFriendlyName   string `form:"monitorfriendlyname" json:"monitorfriendlyname"`
	AlertType             int    `form:"alerttype" json:"alerttype"`
	AlertTypeFriendlyName string `form:"alerttypefriendlyname" json:"alerttypefriendlyname"`
	AlertDetails          string `form:"alertdetails" json:"alertdetails"`
	AlertDuration         string `form:"alertduration" json:"alertduration"`
	AlertDateTime         int64  `form:"alertdatetime" json:"alertdatetime"`
	MonitorAlertContacts  string `form:"monitoralertcontacts" json:"monitoralertcontacts"`
	SSLExpiryDate         int64  `form:"sslexpirydate" json:"sslexpirydate"`
	SSLExpiryDaysLeft     string `form:"sslexpirydaysleft" json:"sslexpirydaysleft"`
	DashboardURL          string `form:"dashboardurl" json:"dashboardurl"`
	MonitorType           string `form:"monitortype" json:"monitortype"`
	HTTPStatusCode        string `form:"httpstatuscode" json:"httpstatuscode"`
	MonitoringRegions     string `form:"monitoringregions" json:"monitoringregions"`
	MonitorTags           string `form:"monitortags" json:"monitortags"`
	MonitorGroup          string `form:"monitorgroup" json:"monitorgroup"`
	IncidentStartTime     int64  `form:"incidentstarttime" json:"incidentstarttime"`
	IncidentEndTime       string `form:"incidentendtime" json:"incidentendtime"`
	ResponseTime          string `form:"responsetime" json:"responsetime"`
}

type Repository interface {
	URHook(ctx context.Context, item UptimeAlertInfo) error
}

type UseCase interface {
	URHook(ctx context.Context, req UptimeAlertItem) error
}

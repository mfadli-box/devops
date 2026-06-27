package backbone

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type useMemory struct {
	db *sql.DB
}

type SessionMemory struct {
	UserId    string    `json:"user_id"`
	CompanyId string    `json:"company_id"`
	ExpiresAt time.Time `json:"expires_at"`
	IsActive  bool      `json:"is_active"`
	IsAdmin   bool      `json:"is_admin"`
	IsHris    bool      `json:"is_hris"`
}

func USLoad() gin.HandlerFunc {
	var db = PgSQL
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token diperlukan",
			})
			return
		}

		token := authHeader
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = strings.TrimSpace(authHeader[7:])
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token diperlukan",
			})
			return
		}

		query := `
			SELECT s.user_id, u.company_id, s.expires_at, u.is_active, u.is_admin, u.is_hris
			FROM   "dat_user_session" s
			JOIN   "dat_user" u ON s.user_id = u.id
			WHERE  s.token = $1
		`

		var session SessionMemory
		ctx := context.Background()
		err := db.QueryRowContext(ctx, query, token).Scan(
			&session.UserId,
			&session.CompanyId,
			&session.ExpiresAt,
			&session.IsActive,
			&session.IsAdmin,
			&session.IsHris,
		)

		if err != nil || session.ExpiresAt.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Sesi kadaluarsa atau tidak valid",
			})
			return
		}

		if !session.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Akun pengguna dinonaktifkan",
			})
			return
		}

		c.Set("userId", session.UserId)
		c.Set("companyId", session.CompanyId)
		c.Set("isAdmin", session.IsAdmin)
		c.Set("isHris", session.IsHris)
		c.Next()
	}
}

func USLock() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Akses khusus Admin",
			})
			return
		}
		c.Next()
	}
}

func USLogs(moduleCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Status() >= http.StatusBadRequest {
			return
		}

		userID := c.GetString("userId")
		if userID == "" {
			return
		}

		companyID := c.GetString("companyId")
		action := strings.TrimSpace(c.Request.Method)
		if action == "" {
			action = "ACTION"
		}
		module := strings.TrimSpace(moduleCode)
		path := strings.TrimSpace(c.Request.URL.Path)
		ipAddress := strings.TrimSpace(c.ClientIP())
		userAgent := strings.TrimSpace(c.GetHeader("User-Agent"))
		auditID := uuid.New().String()

		queryPrimary := `
			INSERT INTO "dat_user_action" (
				id, user_id, company_id, module_code, action, path, ip_address, user_agent, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		`

		_, err := PgSQL.ExecContext(context.Background(), queryPrimary,
			auditID,
			userID,
			companyID,
			module,
			action,
			path,
			ipAddress,
			userAgent,
		)
		if err == nil {
			return
		}

		queryFallback := `
			INSERT INTO "dat_user_action" (
				id, user_id, company_id, module_code, action, path, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, NOW())
		`
		_, err = PgSQL.ExecContext(context.Background(), queryFallback,
			auditID,
			userID,
			companyID,
			module,
			action,
			path,
		)
		if err == nil {
			return
		}

		queryMinimal := `
			INSERT INTO "dat_user_action" (
				user_id, action, path
			) VALUES ($1, $2, $3)
		`
		if _, err = PgSQL.ExecContext(context.Background(), queryMinimal,
			userID,
			action,
			path); err != nil {
			_ = c.Error(err)
		}
	}
}

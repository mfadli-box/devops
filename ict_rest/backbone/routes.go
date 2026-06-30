package backbone

import (
	"ict_rest/skeleton/dat_company"
	"ict_rest/skeleton/dat_module"
	"ict_rest/skeleton/dat_user"
	"ict_rest/skeleton/ict_monitor"
	"ict_rest/skeleton/ict_security"
	"time"

	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/gin-gonic/gin"
)

func SetRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	rest := gin.Default()
	rest.SetTrustedProxies([]string{"localhost", "172.99.66.6"})
	rest.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:36666", "http://172.99.66.6:36666"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	rest.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	rest.GET("/rest", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "rest",
		})
	})
	rest.GET("/hook", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "hook",
		})
	})

	SetDatabase()
	var db = PgSQL

	userR := dat_user.NRepository(db)
	userU := dat_user.NUseCase(userR)
	userH := dat_user.NHandler(userU)

	moduleR := dat_module.NRepository(db)
	moduleU := dat_module.NUseCase(moduleR)
	moduleH := dat_module.NHandler(moduleU)

	companyR := dat_company.NRepository(db)
	companyU := dat_company.NUseCase(companyR)
	companyH := dat_company.NHandler(companyU)

	rest.GET("/rest/company", companyH.PLCompany)
	rest.POST("/rest/user/login", userH.PELogin)
	rest.POST("/rest/user/logout", USLoad(), userH.PELogout)

	rest.GET("/rest/module", USLoad(), moduleH.PLModule)
	rest.GET("/rest/user/profile", USLoad(), userH.PLProfile)
	rest.GET("/rest/user/history", USLoad(), userH.PLHistory)
	rest.GET("/rest/user/company", USLoad(), companyH.PLCompanyUser)
	rest.PUT("/rest/user/profile", USLoad(), userH.PUProfile)
	rest.PUT("/rest/user/password", USLoad(), userH.PUPassword)

	rest.GET("/rest/admin/module", USLoad(), USLock(), moduleH.ALModule)
	rest.GET("/rest/admin/user", USLoad(), USLock(), userH.ALUser)
	rest.GET("/rest/admin/user-company", USLoad(), USLock(), userH.ALUserCompany)
	rest.GET("/rest/admin/user-privilege", USLoad(), USLock(), userH.ALUserPrivilege)
	rest.GET("/rest/admin/company", USLoad(), USLock(), companyH.ALCompany)
	rest.GET("/rest/admin/company-module", USLoad(), USLock(), companyH.ALCompanyModule)
	rest.POST("/rest/admin/user", USLoad(), USLock(), USLogs("c_user"), userH.ACUser)
	rest.POST("/rest/admin/user-company", USLoad(), USLock(), USLogs("c_user_company"), userH.ACUserCompany)
	rest.POST("/rest/admin/user-privilege", USLoad(), USLock(), USLogs("c_user_privilege"), userH.ACUserPrivilege)
	rest.POST("/rest/admin/module", USLoad(), USLock(), USLogs("c_module"), moduleH.ACModule)
	rest.POST("/rest/admin/company", USLoad(), USLock(), USLogs("c_company"), companyH.ACCompany)
	rest.POST("/rest/admin/company-module", USLoad(), USLock(), USLogs("c_company_module"), companyH.ACCompanyModule)
	rest.PUT("/rest/admin/module/:id", USLoad(), USLock(), USLogs("u_module"), moduleH.AUModule)
	rest.PUT("/rest/admin/user/:id", USLoad(), USLock(), USLogs("u_user"), userH.AUUser)
	rest.PUT("/rest/admin/user-company/:id", USLoad(), USLock(), USLogs("u_user_company"), userH.AUUserCompany)
	rest.PUT("/rest/admin/user-privilege/:id", USLoad(), USLock(), USLogs("u_user_privilege"), userH.AUUserPrivilege)
	rest.PUT("/rest/admin/company/:id", USLoad(), USLock(), USLogs("u_company"), companyH.AUCompany)
	rest.PUT("/rest/admin/company-module/:id", USLoad(), USLock(), USLogs("u_company_module"), companyH.AUCompanyModule)

	ictSecurityR := ict_security.NRepository(db)
	ictSecurityU := ict_security.NUseCase(ictSecurityR)
	ictSecurityH := ict_security.NHandler(ictSecurityU)

	ictMonitorR := ict_monitor.NRepository(db)
	ictMonitorU := ict_monitor.NUseCase(ictMonitorR)
	ictMonitorH := ict_monitor.NHandler(ictMonitorU)

	rest.GET("/rest/pages/NW01/sla", USLoad(), ictSecurityH.NLSLA)
	rest.GET("/rest/pages/NW01/waf", USLoad(), ictSecurityH.NLWAF)
	rest.GET("/rest/pages/NW01/attack", USLoad(), ictSecurityH.NLATC)
	rest.GET("/rest/pages/NW01/history", USLoad(), ictSecurityH.NLLOG)
	rest.GET("/rest/pages/NW01/whitelist", USLoad(), ictSecurityH.NLIPW)
	rest.GET("/rest/pages/NW01/blacklist", USLoad(), ictSecurityH.NLIPB)
	rest.POST("/rest/pages/NW01/waf", USLoad(), USLogs("c_NW01_waf"), ictSecurityH.NCWAF)
	rest.POST("/rest/pages/NW01/whitelist", USLoad(), USLogs("c_NW01_whitelist"), ictSecurityH.NCIPM)
	rest.DELETE("/rest/pages/NW01/waf/:id", USLoad(), USLogs("d_NW01_waf"), ictSecurityH.NDWAF)
	rest.DELETE("/rest/pages/NW01/whitelist/:id", USLoad(), USLogs("d_NW01_whitelist"), ictSecurityH.NDIPW)

	rest.GET("/rest/pages/NW02/sla", USLoad(), ictMonitorH.URSla)
	rest.GET("/rest/pages/NW02/summary", USLoad(), ictMonitorH.URSum)
	rest.GET("/rest/pages/NW02/logs", USLoad(), ictMonitorH.URLog)
	rest.DELETE("/rest/pages/NW02/logs/:id", USLoad(), USLogs("d_NW02_log"), ictMonitorH.DURLog)

	rest.POST("/hook/uptimerobot", ictMonitorH.URHook)

	return rest
}

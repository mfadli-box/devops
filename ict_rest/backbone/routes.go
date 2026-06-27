package backbone

import (
	"ict_rest/skeleton/dat_company"
	"ict_rest/skeleton/dat_module"
	"ict_rest/skeleton/dat_user"
	"ict_rest/skeleton/ict_monitor"
	"ict_rest/skeleton/ict_security"

	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/gin-gonic/gin"
)

func SetRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	rest := gin.Default()
	rest.SetTrustedProxies([]string{"127.0.0.1"})

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

	ictMonitorR := ict_monitor.NRepository(db)
	ictMonitorU := ict_monitor.NUseCase(ictMonitorR)
	ictMonitorH := ict_monitor.NHandler(ictMonitorU)

	ictSecurityR := ict_security.NRepository(db)
	ictSecurityU := ict_security.NUseCase(ictSecurityR)
	ictSecurityH := ict_security.NHandler(ictSecurityU)

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

	rest.GET("/rest/security/sla", USLoad(), USLock(), ictSecurityH.NLSLA)
	rest.GET("/rest/security/waf", USLoad(), USLock(), ictSecurityH.NLWAF)
	rest.GET("/rest/security/attack", USLoad(), USLock(), ictSecurityH.NLATC)
	rest.GET("/rest/security/history", USLoad(), USLock(), ictSecurityH.NLLOG)
	rest.GET("/rest/security/whitelist", USLoad(), USLock(), ictSecurityH.NLIPW)
	rest.GET("/rest/security/blacklist", USLoad(), USLock(), ictSecurityH.NLIPB)
	rest.POST("/rest/security/waf", USLoad(), USLock(), USLogs("c_sec_waf"), ictSecurityH.NCWAF)
	rest.POST("/rest/security/whitelist", USLoad(), USLock(), USLogs("c_sec_whitelist"), ictSecurityH.NCIPM)
	rest.DELETE("/rest/security/waf/:id", USLoad(), USLock(), USLogs("d_sec_waf"), ictSecurityH.NDWAF)
	rest.DELETE("/rest/security/whitelist/:id", USLoad(), USLock(), USLogs("d_sec_whitelist"), ictSecurityH.NDIPW)

	rest.POST("/hook/uptimerobot", ictMonitorH.URHook)

	return rest
}

package dat_user

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase UseCase
}

func NHandler(u UseCase) *Handler {
	return &Handler{usecase: u}
}

func NPages(c *gin.Context) (limit, offset int) {
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset = (page - 1) * limit
	return limit, offset
}

func (h *Handler) PELogin(c *gin.Context) {
	var req UserLoginItem
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	res, err := h.usecase.PELogin(c.Request.Context(), req, ipAddress, userAgent)
	if err != nil {
		if errors.Is(err, errors.New("Nama Pengguna atau kata sandi salah")) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}
		if errors.Is(err, errors.New("Akun anda telah dinonaktifkan")) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Terjadi kesalahan server",
		})
		return
	}

	cookiePayload, err := json.Marshal(map[string]string{
		"token":      res.Token,
		"expires_at": res.ExpiresAt.Format(time.RFC3339),
	})
	if err == nil {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "sessionMemorySave",
			Value:    url.QueryEscape(string(cookiePayload)),
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int(time.Until(res.ExpiresAt).Seconds()),
			Expires:  res.ExpiresAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil Masuk Sistem",
		"data":    res,
	})
}

func (h *Handler) PELogout(c *gin.Context) {
	token := GetTokenLogin(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Token diperlukan untuk keluar",
		})
		return
	}

	err := h.usecase.PELogout(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menghapus sesi",
		})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "sessionMemorySave",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   0,
		Expires:  time.Unix(0, 0),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Sesi berhasil dihapus dari basis data",
	})
}

func (h *Handler) PLProfile(c *gin.Context) {
	userID := c.GetString("userId")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Sesi tidak valid",
		})
		return
	}

	profile, err := h.usecase.PLProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil profil",
		})
		return
	}

	profile.Password = ""
	c.JSON(http.StatusOK, gin.H{
		"data": profile,
	})
}

func (h *Handler) PUProfile(c *gin.Context) {
	userID := c.GetString("userId")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Sesi tidak valid",
		})
		return
	}

	var req UserProfileEdit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	profile, err := h.usecase.PUProfile(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	profile.Password = ""
	c.JSON(http.StatusOK, gin.H{
		"message": "Profil berhasil diperbarui",
		"data":    profile,
	})
}

func (h *Handler) PUPassword(c *gin.Context) {
	userID := c.GetString("userId")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Sesi tidak valid",
		})
		return
	}

	var req UserPasswordEdit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := h.usecase.PUPassword(c.Request.Context(), userID, req)
	if err != nil {
		if err.Error() == "Fitur ubah kata sandi tidak tersedia untuk pengguna HRIS" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kata sandi berhasil diperbarui",
	})
}

func (h *Handler) PLHistory(c *gin.Context) {
	userID := c.GetString("userId")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Sesi tidak valid",
		})
		return
	}

	list, err := h.usecase.PLHistory(c.Request.Context(), userID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil riwayat pengguna",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": list,
	})
}

func (h *Handler) ALUser(c *gin.Context) {
	list, err := h.usecase.ALUser(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data pengguna",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": list,
	})
}

func (h *Handler) ACUser(c *gin.Context) {
	var req UserEdit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := h.usecase.ACUser(c.Request.Context(), req); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, errors.New("Nama Pengguna, email, dan nama lengkap wajib diisi")) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Pengguna berhasil ditambahkan",
	})
}

func (h *Handler) AUUser(c *gin.Context) {
	id := c.Param("id")
	var req UserEdit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := h.usecase.AUUser(c.Request.Context(), id, req); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, errors.New("Nama Pengguna, email, dan nama lengkap wajib diisi")) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Pengguna berhasil diperbarui",
	})
}

func (h *Handler) ALUserCompany(c *gin.Context) {
	userID := c.Query("user_id")
	list, err := h.usecase.ALUserCompany(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil hak akses perusahaan pengguna",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": list,
	})
}

func (h *Handler) ACUserCompany(c *gin.Context) {
	var req UserCompanyEdit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := h.usecase.ACUserCompany(c.Request.Context(), req); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, errors.New("ID, ID Pengguna, dan ID Perusahaan wajib diisi")) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Hak akses perusahaan pengguna berhasil ditambahkan",
	})
}

func (h *Handler) AUUserCompany(c *gin.Context) {
	id := c.Param("id")
	var req UserCompanyEdit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := h.usecase.AUUserCompany(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Hak akses perusahaan pengguna berhasil diperbarui",
	})
}

func (h *Handler) ALUserPrivilege(c *gin.Context) {
	userCompanyID := c.Query("user_company_id")
	list, err := h.usecase.ALUserPrivilege(c.Request.Context(), userCompanyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil hak akses modul pengguna",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": list,
	})
}

func (h *Handler) ACUserPrivilege(c *gin.Context) {
	var req UserPrivilegeEdit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := h.usecase.ACUserPrivilege(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Hak akses modul pengguna berhasil ditambahkan",
	})
}

func (h *Handler) AUUserPrivilege(c *gin.Context) {
	id := c.Param("id")
	var req UserPrivilegeEdit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := h.usecase.AUUserPrivilege(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Hak akses modul pengguna berhasil diperbarui",
	})
}

func GetTokenLogin(c *gin.Context) string {
	token := c.GetHeader("Authorization")
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}

	if token != "" {
		return token
	}

	cookie, err := c.Request.Cookie("sessionMemorySave")
	if err != nil {
		return ""
	}

	decoded, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return ""
	}

	var payload struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal([]byte(decoded), &payload); err != nil {
		return ""
	}

	return payload.Token
}

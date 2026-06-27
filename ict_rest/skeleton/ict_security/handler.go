package ict_security

import (
	"net/http"
	"strconv"

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

func (h *Handler) NLSLA(c *gin.Context) {
	limit, offset := NPages(c)

	result, err := h.usecase.NLSLA(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil daftar SLA",
		})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) NLIPW(c *gin.Context) {
	search := c.DefaultQuery("search", "")
	limit, offset := NPages(c)

	result, err := h.usecase.NLIPW(c.Request.Context(), search, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil daftar Whitelist",
		})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) NLIPB(c *gin.Context) {
	search := c.DefaultQuery("search", "")
	limit, offset := NPages(c)

	result, err := h.usecase.NLIPB(c.Request.Context(), search, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil daftar Blacklist",
		})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) NCIPM(c *gin.Context) {
	var req IPWItem
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "JSON payload tidak valid",
		})
		return
	}

	err := h.usecase.NCIPM(c.Request.Context(), req.IP, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Konversi digagalkan secara aman, silakan periksa berkas log sistem.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "IP " + req.IP + " sukses di-whitelist & performa SLA berhasil diperbarui.",
	})
}

func (h *Handler) NDIPW(c *gin.Context) {
	iP := c.Param("ip")

	if iP == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "IP wajib dilampirkan",
		})
		return
	}

	err := h.usecase.NDIPW(c.Request.Context(), iP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menghapus IP dari whitelist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "IP berhasil dihapus dari whitelist.",
	})
}

func (h *Handler) NLWAF(c *gin.Context) {
	search := c.DefaultQuery("search", "")
	limit, offset := NPages(c)

	result, err := h.usecase.NLWAF(c.Request.Context(), search, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil daftar aturan",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) NCWAF(c *gin.Context) {
	var req WAFItem

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON payload tidak valid"})
		return
	}

	err := h.usecase.NCWAF(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Transaksi dibatalkan secara aman. Silakan periksa berkas log sistem.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bypass rule berhasil dibuat. Struktur serangan historis yang sesuai telah dibersihkan dan SLA telah diseimbangkan.",
	})
}

func (h *Handler) NDWAF(c *gin.Context) {
	ruleID := c.Param("id")

	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID wajib dilampirkan",
		})
		return
	}

	err := h.usecase.NDWAF(c.Request.Context(), ruleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menghapus aturan WAF",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Aturan WAF berhasil dihapus.",
	})
}

func (h *Handler) NLATC(c *gin.Context) {
	date := c.DefaultQuery("date", "")

	result, err := h.usecase.NLATC(c.Request.Context(), date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil daftar Serangan",
		})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) NLLOG(c *gin.Context) {
	ipTarget := c.Query("ip")
	dateTarget := c.Query("date")

	if ipTarget == "" || dateTarget == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Parameter 'ip' dan 'date' wajib dilampirkan",
		})
		return
	}

	result, err := h.usecase.NLLOG(c.Request.Context(), ipTarget, dateTarget)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat rincian log harian",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

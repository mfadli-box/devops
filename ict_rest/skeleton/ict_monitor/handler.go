package ict_monitor

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

func (h *Handler) URHook(c *gin.Context) {
	var req UptimeAlertItem

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Format request tidak valid",
		})
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.URHook(ctx, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menyimpan alert",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data berhasil disimpan",
	})
}

func (h *Handler) URLog(c *gin.Context) {
	var filter FilterParams
	_ = c.ShouldBindQuery(&filter)

	res, err := h.usecase.URLog(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil detail log",
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) URSla(c *gin.Context) {
	var filter FilterParams
	_ = c.ShouldBindQuery(&filter)

	res, err := h.usecase.URSla(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data SLA harian",
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) URSum(c *gin.Context) {
	var filter FilterParams
	_ = c.ShouldBindQuery(&filter)

	res, err := h.usecase.URSum(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data summary domain",
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) DURLog(c *gin.Context) {
	logID := c.Param("id")

	ctx := c.Request.Context()
	if err := h.usecase.DURLog(ctx, logID); err != nil {
		if err.Error() == "Log tidak ditemukan" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Log berhasil dihapus",
	})
}

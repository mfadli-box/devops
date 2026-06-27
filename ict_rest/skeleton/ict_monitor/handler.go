package ict_monitor

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase UseCase
}

func NHandler(u UseCase) *Handler {
	return &Handler{usecase: u}
}

func (h *Handler) URHook(c *gin.Context) {
	var req UptimeAlertItem

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.New("Format request tidak valid"),
		})
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.URHook(ctx, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.New("Gagal menyimpan alert"),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Data berhasil disimpan",
	})
}

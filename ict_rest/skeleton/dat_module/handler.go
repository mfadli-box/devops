package dat_module

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

func (h *Handler) PLModule(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetString("userId")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": errors.New("Token diperlukan"),
		})
		return
	}

	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("companyId")
	}

	tree, err := h.usecase.PLModule(ctx, userID, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.New("Gagal memuat daftar modul"),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": tree,
	})
}

func (h *Handler) ALModule(c *gin.Context) {
	list, err := h.usecase.ALModule(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.New("Gagal memuat daftar modul"),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": list,
	})
}

func (h *Handler) ACModule(c *gin.Context) {
	var req ModuleEdit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := h.usecase.ACModule(c.Request.Context(), req); err != nil {
		if errors.Is(err, errors.New("Kode, nama, dan lokasi wajib diisi")) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": errors.New("Kode, nama, dan lokasi wajib diisi"),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Modul berhasil ditambahkan",
	})
}

func (h *Handler) AUModule(c *gin.Context) {
	id := c.Param("id")
	var req ModuleEdit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := h.usecase.AUModule(c.Request.Context(), id, req); err != nil {
		if errors.Is(err, errors.New("ID, kode, nama, dan lokasi wajib diisi")) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": errors.New("ID, kode, nama, dan lokasi wajib diisi"),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Modul berhasil diperbarui",
	})
}

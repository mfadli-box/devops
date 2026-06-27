package dat_company

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

func (h *Handler) PLCompany(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.usecase.PLCompany(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.New("Gagal memuat daftar perusahaan"),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": list,
	})
}

func (h *Handler) PLCompanyUser(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetString("userId")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": errors.New("Token diperlukan"),
		})
		return
	}
	list, err := h.usecase.PLCompanyUser(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.New("Gagal memuat daftar perusahaan untuk pengguna"),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *Handler) ALCompany(c *gin.Context) {
	list, err := h.usecase.ALCompany(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.New("Gagal memuat daftar perusahaan"),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": list,
	})
}

func (h *Handler) ACCompany(c *gin.Context) {
	var req CompanyEdit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := h.usecase.ACCompany(c.Request.Context(), req); err != nil {
		if errors.Is(err, errors.New("Kode dan nama wajib diisi")) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": errors.New("Kode dan nama wajib diisi"),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Perusahaan berhasil ditambahkan",
	})
}

func (h *Handler) AUCompany(c *gin.Context) {
	id := c.Param("id")
	var req CompanyEdit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := h.usecase.AUCompany(c.Request.Context(), id, req); err != nil {
		if errors.Is(err, errors.New("ID, Kode dan nama wajib diisi")) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": errors.New("ID, Kode dan nama wajib diisi"),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Perusahaan berhasil diperbarui",
	})
}

func (h *Handler) ALCompanyModule(c *gin.Context) {
	companyID := c.Query("company_id")
	list, err := h.usecase.ALCompanyModule(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.New("Gagal memuat daftar modul perusahaan"),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": list,
	})
}

func (h *Handler) ACCompanyModule(c *gin.Context) {
	var req CompanyModuleEdit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := h.usecase.ACCompanyModule(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Hak akses perusahaan berhasil ditambahkan",
	})
}

func (h *Handler) AUCompanyModule(c *gin.Context) {
	id := c.Param("id")
	var req CompanyModuleEdit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := h.usecase.AUCompanyModule(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Hak akses perusahaan berhasil diperbarui",
	})
}

package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"wellness_tracker/internal/models"
)

type CheckinHandler struct {
	db *gorm.DB
}

func NewCheckinHandler(db *gorm.DB) *CheckinHandler {
	return &CheckinHandler{db: db}
}

type createCheckinRequest struct {
	UserID          uint    `json:"userId" binding:"required"`
	MoodScore       uint    `json:"moodScore" binding:"required"`
	SleepHours      float32 `json:"sleepHours" binding:"required"`
	EnergyLevel     uint    `json:"energyLevel" binding:"required"`
	MedicationTaken bool    `json:"medicationTaken"`
	Note            string  `json:"note"`
}

type updateCheckinRequest struct {
	MoodScore       *uint    `json:"moodScore"`
	SleepHours      *float32 `json:"sleepHours"`
	EnergyLevel     *uint    `json:"energyLevel"`
	MedicationTaken *bool    `json:"medicationTaken"`
	Note            *string  `json:"note"`
}

// List returns a paginated list of check-ins.
func (h *CheckinHandler) List(c *gin.Context) {
	ctx := c.Request.Context()

	page, pageSize := parsePagination(c, 1, 20, 100)
	var checkins []models.Checkin

	tx := h.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize)

	if err := tx.Find(&checkins).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list check-ins")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": checkins,
		"meta": gin.H{
			"page":      page,
			"pageSize":  pageSize,
			"itemCount": len(checkins),
		},
	})
}

// Get returns a single check-in by ID.
func (h *CheckinHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	var checkin models.Checkin
	if err := h.db.WithContext(ctx).First(&checkin, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondError(c, http.StatusNotFound, "check-in not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to load check-in")
		return
	}

	c.JSON(http.StatusOK, checkin)
}

// Create inserts a new check-in.
func (h *CheckinHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	var req createCheckinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Check if a check-in already exists for this user today
	var existing models.Checkin
	err := h.db.WithContext(ctx).
		Where("user_id = ? AND DATE(created_at) = CURRENT_DATE", req.UserID).
		First(&existing).Error

	if err == nil {
		respondError(c, http.StatusConflict, "you already have a check-in for today, please update it instead")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		respondError(c, http.StatusInternalServerError, "failed to validate check-in")
		return
	}

	checkin := models.Checkin{
		UserID:          req.UserID,
		MoodScore:       req.MoodScore,
		SleepHours:      req.SleepHours,
		EnergyLevel:     req.EnergyLevel,
		MedicationTaken: req.MedicationTaken,
		Note:            req.Note,
	}

	if err := h.db.WithContext(ctx).Create(&checkin).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create check-in")
		return
	}

	c.JSON(http.StatusCreated, checkin)
}

// Update performs a partial update on an existing check-in.
func (h *CheckinHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	var req updateCheckinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	var checkin models.Checkin
	if err := h.db.WithContext(ctx).First(&checkin, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(c, http.StatusNotFound, "check-in not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to load check-in")
		return
	}

	applyCheckinUpdates(&checkin, req)

	if err := h.db.WithContext(ctx).Save(&checkin).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update check-in")
		return
	}

	c.JSON(http.StatusOK, checkin)
}

// applyCheckinUpdates applies non-nil fields from the request to the checkin model.
func applyCheckinUpdates(checkin *models.Checkin, req updateCheckinRequest) {
	if req.Note != nil {
		checkin.Note = *req.Note
	}
	if req.MoodScore != nil {
		checkin.MoodScore = *req.MoodScore
	}
	if req.SleepHours != nil {
		checkin.SleepHours = *req.SleepHours
	}
	if req.EnergyLevel != nil {
		checkin.EnergyLevel = *req.EnergyLevel
	}
	if req.MedicationTaken != nil {
		checkin.MedicationTaken = *req.MedicationTaken
	}
}

// Delete removes a check-in by ID.
func (h *CheckinHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	tx := h.db.WithContext(ctx).Delete(&models.Checkin{}, id)
	if tx.Error != nil {
		respondError(c, http.StatusInternalServerError, "failed to delete check-in")
		return
	}

	if tx.RowsAffected == 0 {
		respondError(c, http.StatusNotFound, "check-in not found")
		return
	}

	c.Status(http.StatusNoContent)
}

func respondError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": message,
	})
}

func parseIDParam(c *gin.Context) (uint, bool) {
	raw := c.Param("id")
	id64, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id64 == 0 {
		respondError(c, http.StatusBadRequest, "invalid id parameter")
		return 0, false
	}
	return uint(id64), true
}

func parsePagination(c *gin.Context, defaultPage, defaultPageSize, maxPageSize int) (int, int) {
	page := defaultPage
	pageSize := defaultPageSize

	if v := c.Query("page"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if v := c.Query("page_size"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	return page, pageSize
}

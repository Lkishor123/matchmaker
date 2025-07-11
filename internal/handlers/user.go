package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"matchmaker/internal/database"
	"matchmaker/internal/httputil"
	"matchmaker/internal/logging"
	"matchmaker/internal/models"
)

// CreateUser creates a user if it does not exist and returns the ID.
func CreateUser(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		logging.Log.WithError(err).Warn("invalid create user payload")
		httputil.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	var user models.User
	err := database.DB.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			user = models.User{Email: req.Email}
			if err := database.DB.Create(&user).Error; err != nil {
				logging.Log.WithError(err).Error("failed to create user")
				httputil.JSONError(c, http.StatusInternalServerError, "create failed")
				return
			}
			c.JSON(http.StatusCreated, gin.H{"id": user.ID})
			return
		}
		logging.Log.WithError(err).Error("failed to query user")
		httputil.JSONError(c, http.StatusInternalServerError, "database error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": user.ID})
}

// GetMe returns the authenticated user's profile.
func GetMe(c *gin.Context) {
	uid := c.GetUint("user_id")
	var user models.User
	if err := database.DB.Preload("BirthDetail").First(&user, uid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			httputil.JSONError(c, http.StatusNotFound, "user not found")
			return
		}
		logging.Log.WithError(err).Error("failed to fetch user")
		httputil.JSONError(c, http.StatusInternalServerError, "database error")
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateMe updates profile fields for the authenticated user.
func UpdateMe(c *gin.Context) {
	uid := c.GetUint("user_id")
	var req struct {
		Gender   string `json:"gender"`
		Location string `json:"location"`
		PhotoURL string `json:"photoURL"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Log.WithError(err).Warn("invalid update payload")
		httputil.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}
	var user models.User
	if err := database.DB.First(&user, uid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			httputil.JSONError(c, http.StatusNotFound, "user not found")
			return
		}
		logging.Log.WithError(err).Error("failed to fetch user")
		httputil.JSONError(c, http.StatusInternalServerError, "database error")
		return
	}
	if req.Gender != "" {
		user.Gender = req.Gender
	}
	if req.Location != "" {
		user.Location = req.Location
	}
	if req.PhotoURL != "" {
		user.PhotoURL = req.PhotoURL
	}
	if err := database.DB.Save(&user).Error; err != nil {
		logging.Log.WithError(err).Error("failed to update user")
		httputil.JSONError(c, http.StatusInternalServerError, "update failed")
		return
	}
	c.JSON(http.StatusOK, user)
}

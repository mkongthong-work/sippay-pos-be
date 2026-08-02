package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pos-backend/db"
	"pos-backend/models"
)

func ListTables(c *gin.Context) {
	var tables []models.DiningTable
	db.DB.Order("name asc").Find(&tables)
	c.JSON(http.StatusOK, tables)
}

func CreateTable(c *gin.Context) {
	var table models.DiningTable
	if err := c.ShouldBindJSON(&table); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if table.Status == "" {
		table.Status = "available"
	}
	if err := db.DB.Create(&table).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, table)
}

func UpdateTable(c *gin.Context) {
	id := c.Param("id")
	var table models.DiningTable
	if err := db.DB.First(&table, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "table not found"})
		return
	}
	var input models.DiningTable
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	table.Name = input.Name
	table.Zone = input.Zone
	table.Capacity = input.Capacity
	if input.Status != "" {
		table.Status = input.Status
	}
	db.DB.Save(&table)
	c.JSON(http.StatusOK, table)
}

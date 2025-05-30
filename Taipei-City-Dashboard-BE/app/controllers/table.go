// Package controllers stores all the controllers for the Gin router.
package controllers

import (
	"TaipeiCityDashboardBE/app/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// POST /api/v1/tables/import
func ImportTableData(c *gin.Context) {
    tableName := c.PostForm("tableName")
    columnsJson := c.PostForm("columns")
    file, _, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing file"})
        return
    }
    defer file.Close()

    if err := models.ImportTableFromCSV(tableName, columnsJson, file); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Table created and data imported"})
}

// GET /api/v1/tables
func ListTables(c *gin.Context) {
    names, err := models.ListAllTableNames()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"tables": names})
}

// GET /api/v1/tables/:tableName/sample?limit={n}
func GetTableSample(c *gin.Context) {
    table := c.Param("tableName")
    limit := 3
    if limStr := c.Query("limit"); limStr != "" {
        if lim, err := strconv.Atoi(limStr); err == nil && lim > 0 {
            limit = lim
        }
    }
	
    cols, rows, err := models.GetSampleRows(table, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"table": table, "columns": cols, "rows": rows})
}

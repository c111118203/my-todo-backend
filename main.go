package main

import (
	"os"
	"net/http"
	"github.com/gin-contrib/cors"   // CORS 套件
	"github.com/gin-gonic/gin"      // Gin 框架
	"github.com/glebarez/sqlite"    // 純 Go 版 SQLite (Windows 友善)
	"gorm.io/gorm"                  // GORM ORM
	
)

// Todo 結構體定義
type Todo struct {
	gorm.Model        // 自動加入 ID, CreatedAt, UpdatedAt, DeletedAt
	Title  string `json:"title"`
	Status bool   `json:"status"`
}
var db *gorm.DB
// 初始化資料庫
func initDB() {
	var err error
	// 使用 glebarez/sqlite 連接資料庫
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("資料庫連線失敗: " + err.Error())
	}

	// 自動建立資料表
	db.AutoMigrate(&Todo{})
}

func main() {
	// 1. 初始化資料庫
	initDB()

	// 2. 初始化 Gin 引擎
	router := gin.Default()

	// 3. 設定 CORS (跨域通行證)
	// 這行非常重要，是為了讓下週的 Vue 前端可以連線
	router.Use(cors.Default())

	// --- 路由設定 ---

	// 獲取所有事項
	router.GET("/todos", func(c *gin.Context) {
		var todos []Todo
		db.Find(&todos)
		c.JSON(http.StatusOK, todos)
	})

	// 新增事項
	router.POST("/todos", func(c *gin.Context) {
		var newTodo Todo
		if err := c.ShouldBindJSON(&newTodo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.Create(&newTodo)
		c.JSON(http.StatusCreated, newTodo)
	})

	// 更新事項狀態 (切換完成/未完成)
	router.PUT("/todos/:id", func(c *gin.Context) {
		id := c.Param("id")
		var todo Todo
		if err := db.First(&todo, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "找不到資料"})
			return
		}
		todo.Status = !todo.Status
		db.Save(&todo)
		c.JSON(http.StatusOK, todo)
	})

	// 刪除事項
	router.DELETE("/todos/:id", func(c *gin.Context) {
		id := c.Param("id")
		var todo Todo

		// 第一步：先找看看這筆資料在不在
		if err := db.First(&todo, id).Error; err != nil {
			// 如果找不到，db.First 會回傳 record not found 錯誤
			c.JSON(http.StatusNotFound, gin.H{"error": "刪除失敗：找不到 ID 為 " + id + " 的資料"})
			return
		}

		// 第二步：確定在，才執行刪除
		db.Unscoped().Delete(&todo)
		
		c.JSON(http.StatusOK, gin.H{"message": "ID " + id + " 已成功永久刪除"})
	})
	// 取得雲端平台分配的 Port，如果沒有則預設 8080
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		// 啟動伺服器，改為 ":"+port
		router.Run(":" + port)
}
package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Users struct {
	Id       int    `json:"id" gorm:"column:id;"`
	Username string `json:"username" gorm:"column:username;"`
	Name     string `json:"name" gorm:"column:name;"`
	Phone    string `json:"phone" gorm:"column:phone;"`
}

type ObjRequest struct {
	RequestId   string      `json:"requestId"`
	RequestTime string      `json:"requestTime"`
	Data        DataRequest `json:"data"`
}

type DataRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
}

type ObjResponse struct {
	ResponseId      string       `json:"responseId"`
	ResponseTime    string       `json:"responseTime"`
	ResponseCode    string       `json:"responseCode"`
	ResponseMessage string       `json:"responseMessage"`
	Data            DateResponse `json:"data"`
}

type DateResponse struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
}

func (Users) TableName() string { return "users" }

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/buoi2?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalln("Cannot connect to MySQL:", err)
	}

	log.Println("Connected to MySQL:", db)

	router := gin.Default()

	v1 := router.Group("/v1")
	{
		v1.POST("/users", createUser(db))           // create user
		v1.GET("/users", getListOfUsers(db))        // list users
		v1.GET("/users/:id", readUserById(db))      // get an user by ID
		v1.PUT("/users/:id", editUserById(db))      // edit an user by ID
		v1.DELETE("/users/:id", deleteUserById(db)) // delete an user by ID
	}

	router.Run()
}

func createUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var dataUser Users

		if err := c.ShouldBind(&dataUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// preprocess title - trim all spaces
		dataUser.Username = strings.TrimSpace(dataUser.Username)
		dataUser.Name = strings.TrimSpace(dataUser.Name)
		dataUser.Phone = strings.TrimSpace(dataUser.Phone)

		if dataUser.Username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username cannot be blank"})
			return
		}

		if dataUser.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name cannot be blank"})
			return
		}

		if dataUser.Phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone cannot be blank"})
			return
		}

		var count int64
		if db.Model(&Users{}).Where("username = ?", dataUser.Username).Count(&count); count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate data"})
			return
		}

		if err := db.Create(&dataUser).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": dataUser.Id})
	}
}

func readUserById(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var dataUser Users

		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Where("id = ?", id).First(&dataUser).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": dataUser})
	}
}

func getListOfUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type DataPaging struct {
			Page  int   `json:"page" form:"page"`
			Limit int   `json:"limit" form:"limit"`
			Total int64 `json:"total" form:"-"`
		}

		var paging DataPaging

		if err := c.ShouldBind(&paging); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if paging.Page <= 0 {
			paging.Page = 1
		}

		if paging.Limit <= 0 {
			paging.Limit = 10
		}

		offset := (paging.Page - 1) * paging.Limit

		var result []Users

		if err := db.Table(Users{}.TableName()).
			Count(&paging.Total).
			Offset(offset).
			Order("id desc").
			Find(&result).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": result})
	}
}

func editUserById(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var dataUser Users

		// preprocess title - trim all spaces
		dataUser.Username = strings.TrimSpace(dataUser.Username)
		dataUser.Name = strings.TrimSpace(dataUser.Name)
		dataUser.Phone = strings.TrimSpace(dataUser.Phone)

		if dataUser.Username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username cannot be blank"})
			return
		}

		if dataUser.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name cannot be blank"})
			return
		}

		if dataUser.Phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone cannot be blank"})
			return
		}

		// check have id exist in db
		if err := db.Where("id = ?", id).First(&dataUser).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// map
		if err := c.ShouldBind(&dataUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Where("id = ?", id).Updates(&dataUser).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

func deleteUserById(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var dataUser Users

		// check have id exist in db
		if err := db.Where("id = ?", id).First(&dataUser).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Table(Users{}.TableName()).
			Where("id = ?", id).
			Delete(nil).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

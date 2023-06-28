package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		v1.POST("/users", createUser(db))         // create user
		v1.GET("/users", getListOfUsers(db))      // list users
		v1.POST("/search-user", readUserById(db)) // get an user by ID
		v1.POST("/update-user", editUserById(db)) // edit an user by ID
		v1.DELETE("/users", deleteUserById(db))   // delete an user by ID
	}

	router.Run()
}

func createUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var objReq ObjRequest
		datetime := time.Now().UTC()

		if err := c.ShouldBind(&objReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// preprocess title - trim all spaces
		objReq.Data.Username = strings.TrimSpace(objReq.Data.Username)
		objReq.Data.Name = strings.TrimSpace(objReq.Data.Name)
		objReq.Data.Phone = strings.TrimSpace(objReq.Data.Phone)

		objRes := ObjResponse{ResponseId: uuid.New().String(), ResponseTime: datetime.Format(time.RFC3339)}

		if objReq.Data.Username == "" {
			objRes.ResponseCode = "01"
			objRes.ResponseMessage = "Invalid Username"

			c.JSON(http.StatusBadRequest, objRes)
			return
		}

		if objReq.Data.Name == "" {
			objRes.ResponseCode = "02"
			objRes.ResponseMessage = "Invalid name"

			c.JSON(http.StatusBadRequest, objRes)
			return
		}

		if objReq.Data.Phone == "" {
			objRes.ResponseCode = "03"
			objRes.ResponseMessage = "Invalid Phone"

			c.JSON(http.StatusBadRequest, objRes)
			return
		}

		data := Users{Username: objReq.Data.Username, Name: objReq.Data.Name, Phone: objReq.Data.Phone}

		var count int64
		if db.Model(&Users{}).Where("username = ?", data.Username).Count(&count); count > 0 {
			objRes.ResponseCode = "04"
			objRes.ResponseMessage = "Duplicate data"

			c.JSON(http.StatusBadRequest, objRes)
			return
		}

		if err := db.Create(&data).Error; err != nil {
			objRes.ResponseCode = "05"
			objRes.ResponseMessage = "Cannot insert db, pls contact administrator"
			log.Fatalln("Create user: ", err.Error())
			c.JSON(http.StatusBadRequest, objRes)
			return
		}

		objRes.ResponseCode = "00"
		objRes.ResponseMessage = "Success"

		c.JSON(http.StatusOK, objRes)
	}
}

func readUserById(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var objReq ObjRequest
		c.BindJSON(&objReq)
		datetime := time.Now().UTC()

		//id, err := strconv.Atoi(c.Param("id"))

		objRes := ObjResponse{ResponseId: uuid.New().String(), ResponseTime: datetime.Format(time.RFC3339)}

		// if err != nil {
		// 	objRes.ResponseCode = "06"
		// 	objRes.ResponseMessage = "BadRequest"
		// 	c.JSON(http.StatusBadRequest, objRes)
		// 	return
		// }

		// preprocess title - trim all spaces
		objReq.Data.Username = strings.TrimSpace(objReq.Data.Username)

		log.Printf("Check username: " + objReq.Data.Username)
		if objReq.Data.Username == "" {
			objRes.ResponseCode = "01"
			objRes.ResponseMessage = "Invalid Username"

			c.JSON(http.StatusBadRequest, objRes)
			return
		}

		var dataUser Users
		if err := db.Where("username = ?", objReq.Data.Username).First(&dataUser).Error; err != nil {
			objRes.ResponseCode = "07"
			objRes.ResponseMessage = "Not found data with username: " + objReq.Data.Username
			c.JSON(http.StatusBadRequest, objRes)
			//c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		objRes.ResponseCode = "00"
		objRes.ResponseMessage = "Success"
		objRes.Data.Name = dataUser.Name
		objRes.Data.Username = dataUser.Username
		objRes.Data.Phone = dataUser.Phone

		c.JSON(http.StatusOK, objRes)
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
		var objReq ObjRequest
		c.BindJSON(&objReq)
		datetime := time.Now().UTC()
		// id, err := strconv.Atoi(c.Param("id"))

		// if err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	return
		// }

		objRes := ObjResponse{ResponseId: uuid.New().String(), ResponseTime: datetime.Format(time.RFC3339)}
		// preprocess title - trim all spaces
		objRes.Data.Username = strings.TrimSpace(objReq.Data.Username)
		objRes.Data.Name = strings.TrimSpace(objReq.Data.Name)
		objRes.Data.Phone = strings.TrimSpace(objReq.Data.Phone)

		log.Printf("Check username: " + objReq.Data.Username)
		if objRes.Data.Username == "" {
			objRes.ResponseCode = "01"
			objRes.ResponseMessage = "Invalid Username"

			c.JSON(http.StatusBadRequest, objRes)
			return
		}

		if objRes.Data.Name == "" {
			objRes.ResponseCode = "02"
			objRes.ResponseMessage = "Invalid name"

			c.JSON(http.StatusBadRequest, objRes)
			return
		}

		if objRes.Data.Phone == "" {
			objRes.ResponseCode = "03"
			objRes.ResponseMessage = "Invalid Phone"

			c.JSON(http.StatusBadRequest, objRes)
			return
		}

		dataUser := Users{Username: objRes.Data.Username, Name: objRes.Data.Name, Phone: objRes.Data.Phone}

		// preprocess title - trim all spaces
		// dataUser.Username = strings.TrimSpace(dataUser.Username)
		// dataUser.Name = strings.TrimSpace(dataUser.Name)
		// dataUser.Phone = strings.TrimSpace(dataUser.Phone)

		// check have id exist in db
		if err := db.Where("username = ?", dataUser.Username).First(&dataUser).Error; err != nil {
			objRes.ResponseCode = "09"
			objRes.ResponseMessage = "NotFound Data"

			c.JSON(http.StatusBadRequest, objRes)
			//c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// map
		// if err := c.ShouldBind(&dataUser); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	return
		// }

		if err := db.Where("username = ?", dataUser.Username).Updates(&dataUser).Error; err != nil {
			objRes.ResponseCode = "10"
			objRes.ResponseMessage = "Until error update data"

			c.JSON(http.StatusBadRequest, objRes)
			//c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		objRes.ResponseCode = "00"
		objRes.ResponseMessage = "Success"
		c.JSON(http.StatusOK, objRes)
	}
}

func deleteUserById(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var objReq ObjRequest
		c.BindJSON(&objReq)
		datetime := time.Now().UTC()
		// id, err := strconv.Atoi(c.Param("id"))

		// if err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	return
		// }

		objRes := ObjResponse{ResponseId: uuid.New().String(), ResponseTime: datetime.Format(time.RFC3339)}
		// preprocess title - trim all spaces
		objRes.Data.Username = strings.TrimSpace(objReq.Data.Username)
		objRes.Data.Name = strings.TrimSpace(objReq.Data.Name)
		objRes.Data.Phone = strings.TrimSpace(objReq.Data.Phone)

		dataUser := Users{Username: objRes.Data.Username, Name: objRes.Data.Name, Phone: objRes.Data.Phone}

		// check have id exist in db
		if err := db.Where("username = ?", dataUser.Username).First(&dataUser).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Table(Users{}.TableName()).Where("username = ?", dataUser.Username).Delete(nil).Error; err != nil {
			objRes.ResponseCode = "10"
			objRes.ResponseMessage = "Until error update data"
			//c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.JSON(http.StatusBadRequest, objRes)
			return
		}

		objRes.ResponseCode = "00"
		objRes.ResponseMessage = "Success"
		c.JSON(http.StatusOK, objRes)
	}
}

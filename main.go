package main

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "products-api/docs"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

type Product struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Categories  string  `json:"categories"`
}

func initDB() {
	dsn := "root:@tcp(localhost:3306)/bazadanix?charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_0900_ai_ci"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}
	db.AutoMigrate(&Product{})
}

// @Summary Создание нового продукта
// @Description Создает новый продукт с указанными параметрами
// @Accept json
// @Produce json
// @Param product body object true "Данные продукта" example={"name":"Laptop","description":"High-performance laptop","price":999.99,"categories":["Electronics","Computers"]}
// @Success 201 {object} Product "Успешно созданный продукт" example={"id":1,"name":"Laptop","description":"High-performance laptop","price":999.99,"categories":["Electronics, Computers"]}
// @Failure 400 {object} map[string]string "Ошибка валидации" example={"error":"Ошибка парсинга данных: name is required"}
// @Failure 500 {object} map[string]string "Ошибка сервера" example={"error":"Ошибка при сохранении в базу: database error"}
// @Router /products [post]
func createProduct(c *gin.Context) {
	var input struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Price       float64  `json:"price" binding:"required,gt=0"`
		Categories  []string `json:"categories"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка парсинга данных: " + err.Error()})
		return
	}

	categoriesStr := strings.Join(input.Categories, ", ")
	product := Product{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Categories:  categoriesStr,
	}

	log.Printf("Получены данные для создания: %+v\n", product)
	result := db.Create(&product)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении в базу: " + result.Error.Error()})
		return
	}
	c.JSON(http.StatusCreated, product)
}

// @Summary Обновление продукта
// @Description Обновляет существующий продукт по ID
// @Accept json
// @Produce json
// @Param id path string true "ID продукта" example="1"
// @Param product body object true "Данные продукта" example={"name":"Updated Laptop","description":"Updated description","price":1099.99,"categories":["Electronics","Gadgets"]}
// @Success 200 {object} Product "Обновленный продукт" example={"id":1,"name":"Updated Laptop","description":"Updated description","price":1099.99,"categories":"Electronics, Gadgets"}
// @Failure 400 {object} map[string]string "Ошибка валидации" example={"error":"name is required"}
// @Failure 404 {object} map[string]string "Продукт не найден" example={"error":"Product not found"}
// @Failure 500 {object} map[string]string "Ошибка сервера" example={"error":"Failed to save product"}
// @Router /products/{id} [put]
func updateProduct(c *gin.Context) {
	id := c.Param("id")
	var product Product
	if err := db.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var input struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Price       float64  `json:"price" binding:"required,gt=0"`
		Categories  []string `json:"categories"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product.Name = input.Name
	product.Description = input.Description
	product.Price = input.Price
	product.Categories = strings.Join(input.Categories, ", ")

	if err := db.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save product"})
		return
	}
	c.JSON(http.StatusOK, product)
}

// @Summary Удаление продукта
// @Description Удаляет продукт по ID
// @Produce json
// @Param id path string true "ID продукта" example="1"
// @Success 200 {object} map[string]string "Успешное удаление" example={"message":"Product deleted successfully"}
// @Failure 500 {object} map[string]string "Ошибка сервера" example={"error":"Failed to delete product"}
// @Router /products/{id} [delete]
func deleteProduct(c *gin.Context) {
	id := c.Param("id")
	if err := db.Delete(&Product{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// @Summary Получение всех продуктов
// @Description Возвращает список всех продуктов
// @Produce json
// @Success 200 {array} Product "Список продуктов" example=[{"id":1,"name":"Laptop","description":"High-performance laptop","price":999.99,"categories":"Electronics, Computers"}]
// @Router /products [get]
func getProducts(c *gin.Context) {
	var products []Product
	db.Find(&products)
	c.JSON(http.StatusOK, products)
}

// @Summary Получение продукта по ID
// @Description Возвращает информацию о конкретном продукте
// @Produce json
// @Param id path string true "ID продукта" example="1"
// @Success 200 {object} Product "Данные продукта" example={"id":1,"name":"Laptop","description":"High-performance laptop","price":999.99,"categories":"Electronics, Computers"}
// @Failure 404 {object} map[string]string "Продукт не найден" example={"error":"Product not found"}
// @Router /products/{id} [get]
func getProductByID(c *gin.Context) {
	id := c.Param("id")
	var product Product
	if err := db.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

// @Summary Получение списка категорий
// @Description Возвращает список уникальных категорий
// @Produce json
// @Success 200 {array} string "Список категорий" example=["Electronics","Computers","Gadgets"]
// @Router /categories [get]
func getCategories(c *gin.Context) {
	var products []Product
	db.Find(&products)

	categorySet := make(map[string]struct{})
	for _, product := range products {
		categories := strings.Split(product.Categories, ", ")
		for _, cat := range categories {
			if cat != "" {
				categorySet[cat] = struct{}{}
			}
		}
	}

	var categoryNames []string
	for cat := range categorySet {
		categoryNames = append(categoryNames, cat)
	}
	c.JSON(http.StatusOK, categoryNames)
}

func serveFrontend() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./frontend")))
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		panic("Failed to start frontend server: " + err.Error())
	}
}

func serveAdmin() {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./admin/"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join("./admin", "admin.html"))
		} else {
			fileServer.ServeHTTP(w, r)
		}
	})
	server := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		panic("Failed to start admin server: " + err.Error())
	}
}

func main() {
	initDB()
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://localhost:8081"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.POST("/products", createProduct)
	r.PUT("/products/:id", updateProduct)
	r.DELETE("/products/:id", deleteProduct)
	r.GET("/products", getProducts)
	r.GET("/products/:id", getProductByID)
	r.GET("/categories", getCategories)

	go func() {
		if err := r.Run(":3000"); err != nil {
			panic("Failed to start API server: " + err.Error())
		}
	}()

	go serveFrontend()
	serveAdmin()
}

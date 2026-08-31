package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	 amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)



type OrderCreatedEvent struct {
	Event   string `json:"event"`
	UserID  string `json:"userId"`
	OrderID string `json:"orderId"`
}

var ctx = context.Background()
var rdb *redis.Client

const JWT_SECRET = "rahasia_super_aman_kamu" // Samakan dengan User Service

type CartItem struct {
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Name      string  `json:"name"`
}

type AddCartRequest struct {
	ProductID int     `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
	Name      string  `json:"name" binding:"required"`
}

func main() {
	// 1. Inisialisasi Koneksi Redis
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "redis-cart"
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:6379", redisHost),
	})

	go initRabbitMQConsumer()	

	r := gin.Default()

	

	// Group Route Cart dengan Middleware Auth JWT
	cartRoutes := r.Group("/api/cart")
	cartRoutes.Use(JWTMiddleware())
	{
		cartRoutes.POST("/add", AddToCart)
		cartRoutes.GET("", GetCart)
		cartRoutes.DELETE("/item/:id", RemoveFromCart)
		cartRoutes.DELETE("/clear", ClearCart)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}

	fmt.Printf("Shopping Cart Service (Go) running on port %s\n", port)
	r.Run(":" + port)
}

// --- MIDDLEWARE JWT ---
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(JWT_SECRET), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
			c.Abort()
			return
		}

		// Ambil userId dari JWT payload
		userId, ok := claims["userId"].(string)
		if !ok || userId == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid userId format in token"})
			c.Abort()
			return
		}
		c.Set("userId", userId)
		c.Next()
	}
}

func initRabbitMQConsumer() {

	// menghubungkan ke server rabbit
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	var conn *amqp.Connection
	var err error

	for {
		conn , err = amqp.Dial(rabbitmqURL)
		if err == nil {
			break
		}
		log.Println("RabbitMQ not ready yet, retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil{
		log.Fatalf("Failed to open a channel: %v", err)	
	}

	defer ch.Close()

	q, err:= ch.QueueDeclare(
		"ORDER_CREATED_QUEUE", // Name	
		true,					 // Durable
		false,					// Delete when unused
		false,					// Exclusive
		false,					// No-wait
		nil,					// Arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,				// Queue
		"",					// Consumer
		true, 				// auto ack 
		false,	
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Println(" [*] Waiting for ORDER_CREATED messages in Cart Service...")

	for d := range msgs {
		log.Printf("Recieved Event: %s", d.Body)
		var event OrderCreatedEvent
		err := json.Unmarshal(d.Body, &event)
		if err == nil && event.UserID != "" {
			ctx := context.Background()
			cartKey := fmt.Sprintf("cart:%s", event.UserID)
			rdb.Del(ctx , cartKey)
			log.Printf("Sucessfully cleared cart for UserID: %s via Event-Driven RabbitMQ!", event.UserID)
		}
	}
}

// --- ENDPOINT: POST /api/cart/add ---
func AddToCart(c *gin.Context) {
	userId := c.MustGet("userId").(string)

	var req AddCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	cartKey := fmt.Sprintf("cart:%s", userId)

	// Ambil cart lama dari Redis
	existingCart, _ := rdb.Get(ctx, cartKey).Result()
	var items []CartItem

	if existingCart != "" {
		json.Unmarshal([]byte(existingCart), &items)
	}

	// Update quantity jika produk sudah ada di cart, atau append baru
	found := false
	for i, item := range items {
		if item.ProductID == req.ProductID {
			items[i].Quantity += req.Quantity
			found = true
			break
		}
	}

	if !found {
		items = append(items, CartItem{
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
			Price:     req.Price,
			Name:      req.Name,
		})
	}

	// Simpan kembali ke Redis
	updatedCartJson, _ := json.Marshal(items)
	err := rdb.Set(ctx, cartKey, updatedCartJson, 0).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product added to cart",
		"data":    items,
	})
}

// --- ENDPOINT: GET /api/cart ---
func GetCart(c *gin.Context) {
	userId := c.MustGet("userId").(string)
	cartKey := fmt.Sprintf("cart:%s", userId)

	cartJson, err := rdb.Get(ctx, cartKey).Result()
	if err == redis.Nil {
		c.JSON(http.StatusOK, gin.H{"items": []CartItem{}, "total_price": 0})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch cart"})
		return
	}

	var items []CartItem
	json.Unmarshal([]byte(cartJson), &items)

	var totalPrice float64
	for _, item := range items {
		totalPrice += item.Price * float64(item.Quantity)
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":     userId,
		"items":       items,
		"total_price": totalPrice,
	})
}

// --- ENDPOINT: DELETE /api/cart/item/:id ---
func RemoveFromCart(c *gin.Context) {
	userId := c.MustGet("userId").(string)
	productIdStr := c.Param("id")
	cartKey := fmt.Sprintf("cart:%s", userId)

	cartJson, err := rdb.Get(ctx, cartKey).Result()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Cart is empty"})
		return
	}

	var items []CartItem
	json.Unmarshal([]byte(cartJson), &items)

	var updatedItems []CartItem
	for _, item := range items {
		if fmt.Sprintf("%d", item.ProductID) != productIdStr {
			updatedItems = append(updatedItems, item)
		}
	}

	updatedCartJson, _ := json.Marshal(updatedItems)
	rdb.Set(ctx, cartKey, updatedCartJson, 0)

	c.JSON(http.StatusOK, gin.H{
		"message": "Item removed from cart",
		"data":    updatedItems,
	})

	
}

func ClearCart(c * gin.Context) {

	userId := c.MustGet("userId").(string) 

	cartKey := fmt.Sprintf("cart:%s", userId)

	err := rdb.Del(ctx, cartKey).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghapus cart",
			"error":   err.Error(),
		})
		
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Keranjang belanja berhasil dikosongkan",
	})

	}

	
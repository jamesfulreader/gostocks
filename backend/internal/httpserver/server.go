package httpserver

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jamesfulreader/gostocks/internal/auth"
	"github.com/jamesfulreader/gostocks/internal/stocks"
	"github.com/jamesfulreader/gostocks/internal/users"
)

type Server struct {
	addr        string
	provider    stocks.Provider
	router      *gin.Engine
	subManager  *SubscriptionManager
	userService *users.Service
}

func New(provider stocks.Provider, db *pgxpool.Pool, addr string) *Server {
	router := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowMethods = []string{"GET", "POST", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	userRepo := users.NewPostgresRepository(db)
	userService := users.NewService(userRepo)

	s := &Server{
		addr:        addr,
		provider:    provider,
		router:      router,
		subManager:  NewSubscriptionManager(),
		userService: userService,
	}

	// Start subscription manager and ticker
	go s.subManager.Run()
	s.StartTicker()

	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	s.router.GET("/ws", s.handleWebSocket)

	api := s.router.Group("/api")
	{
		api.GET("/quote", s.handleQuote)
		api.GET("/intraday", s.handleIntraday)

		// Auth routes
		api.POST("/register", s.handleRegister)
		api.POST("/login", s.handleLogin)

		// Protected routes
		protected := api.Group("/")
		protected.Use(auth.AuthMiddleware())
		{
			protected.GET("/portfolio", s.handleGetPortfolio)
			protected.POST("/portfolio", s.handleAddToPortfolio)
			protected.DELETE("/portfolio", s.handleRemoveFromPortfolio)
		}
	}
}

func (s *Server) ListenAndServe() error {
	return s.router.Run(s.addr)
}

func (s *Server) handleQuote(c *gin.Context) {
	symbol := strings.ToUpper(c.Query("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing symbol"})
		return
	}
	q, err := s.provider.Quote(c.Request.Context(), symbol)
	if err != nil {
		log.Printf("quote error: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, q)
}

func (s *Server) handleIntraday(c *gin.Context) {
	symbol := strings.ToUpper(c.Query("symbol"))
	interval := c.Query("interval")
	if interval == "" {
		interval = "1min"
	}
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing symbol"})
		return
	}
	points, err := s.provider.Intraday(c.Request.Context(), symbol, interval, 100)
	if err != nil {
		log.Printf("intraday error: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, points)
}

func (s *Server) handleRegister(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := s.userService.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (s *Server) handleLogin(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := s.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func (s *Server) handleGetPortfolio(c *gin.Context) {
	userID := c.GetInt("userID")
	portfolio, err := s.userService.GetPortfolio(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get portfolio"})
		return
	}
	c.JSON(http.StatusOK, portfolio)
}

func (s *Server) handleAddToPortfolio(c *gin.Context) {
	userID := c.GetInt("userID")
	var req struct {
		Symbol string `json:"symbol"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := s.userService.AddToPortfolio(c.Request.Context(), userID, req.Symbol); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add to portfolio"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) handleRemoveFromPortfolio(c *gin.Context) {
	userID := c.GetInt("userID")
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing symbol"})
		return
	}

	if err := s.userService.RemoveFromPortfolio(c.Request.Context(), userID, symbol); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove from portfolio"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

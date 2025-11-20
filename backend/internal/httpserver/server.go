package httpserver

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jamesfulreader/gostocks/internal/stocks"
)

type Server struct {
	addr       string
	provider   stocks.Provider
	router     *gin.Engine
	subManager *SubscriptionManager
}

func New(provider stocks.Provider, addr string) *Server {
	router := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type"}
	router.Use(cors.New(config))

	s := &Server{
		addr:       addr,
		provider:   provider,
		router:     router,
		subManager: NewSubscriptionManager(),
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

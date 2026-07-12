package api

import (
	_ "chaosguard/docs" // Load generated Swagger docs
	"chaosguard/internal/api/handlers"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter creates and configures the Gin engine with all routes and middlewares
func SetupRouter(h *handlers.Handler) *gin.Engine {
	// Set Gin to release mode to prevent default debug stdout prints
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// Register custom middlewares
	r.Use(RequestIDMiddleware())
	r.Use(LoggerMiddleware())
	r.Use(CORSMiddleware())
	r.Use(RecoveryMiddleware())

	// Health endpoint
	r.GET("/health", h.GetHealth)

	// Containers endpoints
	r.GET("/containers", h.GetContainers)
	r.GET("/containers/:id", h.GetContainerByID)

	// Experiments endpoints
	r.GET("/experiments", h.GetExperiments)
	r.GET("/experiments/:id", h.GetExperimentByID)
	r.POST("/experiments", h.CreateExperiment)
	r.DELETE("/experiments/:id", h.DeleteExperiment)

	// Scheduler endpoints
	r.GET("/scheduler/status", h.GetSchedulerStatus)
	r.POST("/scheduler/start", h.StartScheduler)
	r.POST("/scheduler/stop", h.StopScheduler)

	// Runtime status endpoint
	r.GET("/runtime", h.GetRuntime)
	r.POST("/runtime/stop", h.StopRuntime)

	// Live logs endpoint
	r.GET("/logs", h.GetLogs)

	// Reuse existing Prometheus handler
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation interactive UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}

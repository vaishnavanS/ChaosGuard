package handlers

import (
	"context"
	"fmt"
	"time"

	"chaosguard/internal/api/requests"
	"chaosguard/internal/api/responses"
	"chaosguard/internal/domain"
	"chaosguard/internal/usecase/scheduler"
	"chaosguard/pkg/config"
	"chaosguard/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RuntimeStateProvider defines the contract for fetching the app lifecycle state
type RuntimeStateProvider interface {
	GetState() string
}

// Handler handles all HTTP API endpoints for ChaosGuard
type Handler struct {
	cfg            *config.Config
	containerRepo  domain.ContainerRepository
	containerCli   domain.ContainerController
	experimentRepo domain.ExperimentRepository
	attackMgr      domain.AttackManager
	recoveryMgr    domain.RecoveryManager
	scheduler      *scheduler.Scheduler
	stateProvider  RuntimeStateProvider
	version        string
	stopFunc       func()
}

// NewHandler creates a new Handler instance
func NewHandler(
	cfg *config.Config,
	containerRepo domain.ContainerRepository,
	containerCli domain.ContainerController,
	experimentRepo domain.ExperimentRepository,
	attackMgr domain.AttackManager,
	recoveryMgr domain.RecoveryManager,
	sched *scheduler.Scheduler,
	stateProvider RuntimeStateProvider,
	version string,
) *Handler {
	return &Handler{
		cfg:            cfg,
		containerRepo:  containerRepo,
		containerCli:   containerCli,
		experimentRepo: experimentRepo,
		attackMgr:      attackMgr,
		recoveryMgr:    recoveryMgr,
		scheduler:      sched,
		stateProvider:  stateProvider,
		version:        version,
	}
}

// GetHealth returns application status, state, and version
// @Summary Get Health status
// @Description Returns the application health status and runtime state
// @Tags Health
// @Produce json
// @Success 200 {object} responses.SuccessResponse{data=responses.HealthResponse}
// @Router /health [get]
func (h *Handler) GetHealth(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{
		Success: true,
		Data: responses.HealthResponse{
			Status:  "healthy",
			State:   h.stateProvider.GetState(),
			Version: h.version,
		},
	})
}

// GetContainers lists all discovered containers
// @Summary List Containers
// @Description Returns a list of target containers and their monitoring status
// @Tags Containers
// @Produce json
// @Success 200 {object} responses.SuccessResponse{data=[]responses.ContainerResponse}
// @Router /containers [get]
func (h *Handler) GetContainers(c *gin.Context) {
	list, err := h.containerRepo.List()
	if err != nil {
		c.JSON(500, responses.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	res := make([]responses.ContainerResponse, 0, len(list))
	for _, item := range list {
		uptime := "0s"
		if item.Status == "running" {
			uptime = time.Since(item.CreatedAt).Round(time.Second).String()
		}
		res = append(res, responses.ContainerResponse{
			ID:          item.ID,
			Name:        item.Name,
			Image:       item.Image,
			Status:      item.Status,
			IsMonitored: item.IsMonitored,
			Uptime:      uptime,
			Labels:      map[string]string{"app": item.Name},
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}
	c.JSON(200, responses.SuccessResponse{Success: true, Data: res})
}

// GetContainerByID retrieves a single container's detailed state
// @Summary Get Container
// @Description Returns details for a single target container
// @Tags Containers
// @Produce json
// @Param id path string true "Container ID"
// @Success 200 {object} responses.SuccessResponse{data=responses.ContainerResponse}
// @Failure 404 {object} responses.ErrorResponse
// @Router /containers/{id} [get]
func (h *Handler) GetContainerByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.containerRepo.Get(id)
	if err != nil {
		c.JSON(404, responses.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	uptime := "0s"
	if item.Status == "running" {
		uptime = time.Since(item.CreatedAt).Round(time.Second).String()
	}

	c.JSON(200, responses.SuccessResponse{
		Success: true,
		Data: responses.ContainerResponse{
			ID:          item.ID,
			Name:        item.Name,
			Image:       item.Image,
			Status:      item.Status,
			IsMonitored: item.IsMonitored,
			Uptime:      uptime,
			Labels:      map[string]string{"app": item.Name},
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		},
	})
}

// GetExperiments lists all registered chaos experiments
// @Summary List Experiments
// @Description Returns a list of chaos experiments executed or scheduled
// @Tags Experiments
// @Produce json
// @Success 200 {object} responses.SuccessResponse{data=[]responses.ExperimentResponse}
// @Router /experiments [get]
func (h *Handler) GetExperiments(c *gin.Context) {
	list, err := h.experimentRepo.List()
	if err != nil {
		c.JSON(500, responses.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	res := make([]responses.ExperimentResponse, 0, len(list))
	for _, item := range list {
		res = append(res, responses.ExperimentResponse{
			ID:                item.ID,
			TargetContainerID: item.TargetContainerID,
			ContainerName:     item.ContainerName,
			AttackType:        item.AttackType,
			Duration:          item.Duration,
			Status:            item.Status,
			RecoveryStatus:    item.RecoveryStatus,
			Parameters:        item.Parameters,
			StartedAt:         item.StartedAt,
			EndedAt:           item.EndedAt,
			ErrorMessage:      item.ErrorMessage,
		})
	}
	c.JSON(200, responses.SuccessResponse{Success: true, Data: res})
}

// GetExperimentByID retrieves details of a single experiment
// @Summary Get Experiment
// @Description Returns details for a single chaos experiment
// @Tags Experiments
// @Produce json
// @Param id path string true "Experiment ID"
// @Success 200 {object} responses.SuccessResponse{data=responses.ExperimentResponse}
// @Failure 404 {object} responses.ErrorResponse
// @Router /experiments/{id} [get]
func (h *Handler) GetExperimentByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.experimentRepo.Get(id)
	if err != nil {
		c.JSON(404, responses.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(200, responses.SuccessResponse{
		Success: true,
		Data: responses.ExperimentResponse{
			ID:                item.ID,
			TargetContainerID: item.TargetContainerID,
			ContainerName:     item.ContainerName,
			AttackType:        item.AttackType,
			Duration:          item.Duration,
			Status:            item.Status,
			RecoveryStatus:    item.RecoveryStatus,
			Parameters:        item.Parameters,
			StartedAt:         item.StartedAt,
			EndedAt:           item.EndedAt,
			ErrorMessage:      item.ErrorMessage,
		},
	})
}

// CreateExperiment starts an ad-hoc chaos experiment
// @Summary Inject Failure
// @Description Starts a manual/ad-hoc chaos attack against a container
// @Tags Experiments
// @Accept json
// @Produce json
// @Param body body requests.CreateExperimentRequest true "Attack Details"
// @Success 211 {object} responses.SuccessResponse{data=responses.ExperimentResponse}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Router /experiments [post]
func (h *Handler) CreateExperiment(c *gin.Context) {
	var req requests.CreateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, responses.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	target, err := h.containerRepo.Get(req.TargetContainerID)
	if err != nil {
		c.JSON(404, responses.ErrorResponse{Success: false, Error: fmt.Sprintf("target container not found: %v", err)})
		return
	}

	_, err = h.attackMgr.Get(req.AttackType)
	if err != nil {
		c.JSON(400, responses.ErrorResponse{Success: false, Error: fmt.Sprintf("invalid attack type: %v", err)})
		return
	}

	duration := time.Duration(req.DurationSec) * time.Second
	experimentID := uuid.New().String()
	exp := &domain.Experiment{
		ID:                experimentID,
		TargetContainerID: target.ID,
		ContainerName:     target.Name,
		AttackType:        req.AttackType,
		Duration:          int64(req.DurationSec),
		Status:            domain.ExperimentStatusPending,
		Parameters:        fmt.Sprintf(`{"duration":"%s"}`, duration.String()),
		StartedAt:         time.Now(),
	}

	if err := h.experimentRepo.Save(exp); err != nil {
		c.JSON(500, responses.ErrorResponse{Success: false, Error: fmt.Sprintf("failed to save experiment: %v", err)})
		return
	}

	if h.recoveryMgr != nil {
		h.recoveryMgr.TrackExperiment(exp)
	}

	// Trigger attack execution asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), duration+10*time.Second)
		defer cancel()

		if err := h.attackMgr.Execute(ctx, exp); err != nil {
			logger.Error(err, "API: Failed to execute chaos attack %s on %s", req.AttackType, target.Name)
			return
		}

		time.Sleep(duration)

		logger.Info("API: Initiating recovery for experiment %s", exp.ID)
		if err := h.recoveryMgr.Recover(context.Background(), exp); err != nil {
			logger.Error(err, "API: Failed to recover container %s after attack", target.Name)
		}
	}()

	c.JSON(201, responses.SuccessResponse{
		Success: true,
		Data: responses.ExperimentResponse{
			ID:                exp.ID,
			TargetContainerID: exp.TargetContainerID,
			ContainerName:     exp.ContainerName,
			AttackType:        exp.AttackType,
			Duration:          exp.Duration,
			Status:            exp.Status,
			Parameters:        exp.Parameters,
			StartedAt:         exp.StartedAt,
		},
	})
}

// DeleteExperiment cancels and cleans up an experiment
// @Summary Cancel Experiment
// @Description Stops a running experiment early and deletes its history record
// @Tags Experiments
// @Produce json
// @Param id path string true "Experiment ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 404 {object} responses.ErrorResponse
// @Router /experiments/{id} [delete]
func (h *Handler) DeleteExperiment(c *gin.Context) {
	id := c.Param("id")
	exp, err := h.experimentRepo.Get(id)
	if err != nil {
		c.JSON(404, responses.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	// If currently active/running, recover container state early
	if exp.Status == domain.ExperimentStatusRunning || exp.Status == domain.ExperimentStatusPending {
		logger.Info("API: Early recovery triggered for deleted experiment %s", exp.ID)
		if err := h.recoveryMgr.Recover(context.Background(), exp); err != nil {
			logger.Error(err, "API: Failed to run early recovery for deleted experiment %s", exp.ID)
		}
	}

	if err := h.experimentRepo.Delete(id); err != nil {
		c.JSON(500, responses.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(200, responses.SuccessResponse{Success: true, Data: map[string]interface{}{"id": id}})
}

// GetSchedulerStatus returns scheduler state and details
// @Summary Get Scheduler Status
// @Description Returns the status and scheduling settings of the automated scheduler
// @Tags Scheduler
// @Produce json
// @Success 200 {object} responses.SuccessResponse{data=responses.SchedulerStatusResponse}
// @Router /scheduler/status [get]
func (h *Handler) GetSchedulerStatus(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{
		Success: true,
		Data: responses.SchedulerStatusResponse{
			Running:        h.scheduler.IsRunning(),
			Mode:           h.cfg.Scheduler.Mode,
			AttackInterval: h.cfg.Scheduler.AttackInterval,
			AttackDuration: h.cfg.Scheduler.AttackDuration,
		},
	})
}

// StartScheduler triggers the scheduler loop
// @Summary Start Scheduler
// @Description Resumes the background chaos failure injection loop
// @Tags Scheduler
// @Produce json
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Router /scheduler/start [post]
func (h *Handler) StartScheduler(c *gin.Context) {
	if h.scheduler.IsRunning() {
		c.JSON(400, responses.ErrorResponse{Success: false, Error: "scheduler is already running"})
		return
	}

	if err := h.scheduler.Start(context.Background()); err != nil {
		c.JSON(500, responses.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(200, responses.SuccessResponse{Success: true, Data: "scheduler started successfully"})
}

// StopScheduler halts the scheduler loop
// @Summary Stop Scheduler
// @Description Pauses the automated scheduler loop
// @Tags Scheduler
// @Produce json
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Router /scheduler/stop [post]
func (h *Handler) StopScheduler(c *gin.Context) {
	if !h.scheduler.IsRunning() {
		c.JSON(400, responses.ErrorResponse{Success: false, Error: "scheduler is not running"})
		return
	}

	if err := h.scheduler.Stop(); err != nil {
		c.JSON(500, responses.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(200, responses.SuccessResponse{Success: true, Data: "scheduler stopped successfully"})
}

// GetRuntime returns composition lifecycle status
// @Summary Get Runtime State
// @Description Returns the lifecycle state of the ChaosGuard daemon
// @Tags Runtime
// @Produce json
// @Success 200 {object} responses.SuccessResponse{data=responses.RuntimeResponse}
// @Router /runtime [get]
func (h *Handler) GetRuntime(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{
		Success: true,
		Data: responses.RuntimeResponse{
			State: h.stateProvider.GetState(),
		},
	})
}

// SetStopFunc registers the runtime shutdown callback
func (h *Handler) SetStopFunc(stopFunc func()) {
	h.stopFunc = stopFunc
}

// StopRuntime triggers a graceful shutdown of the application
// @Summary Stop Runtime
// @Description Triggers a graceful shutdown of the ChaosGuard daemon
// @Tags Runtime
// @Produce json
// @Success 200 {object} responses.SuccessResponse
// @Router /runtime/stop [post]
func (h *Handler) StopRuntime(c *gin.Context) {
	if h.stopFunc == nil {
		c.JSON(500, responses.ErrorResponse{Success: false, Error: "shutdown handler not registered"})
		return
	}

	c.JSON(200, responses.SuccessResponse{
		Success: true,
		Data:    "Daemon shutdown initiated",
	})

	go func() {
		time.Sleep(100 * time.Millisecond)
		h.stopFunc()
	}()
}

package service

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
	"tj-routes/internal/utils"

	"go.uber.org/zap"
)

type BulkUploadService interface {
	UploadCSV(entityType string, file *multipart.FileHeader, userID uint) (*models.BulkUploadLog, error)
	GetUploadStatus(uploadID uint) (*models.BulkUploadLog, error)
	ListUploads(offset, limit int, filters map[string]interface{}) ([]models.BulkUploadLog, int64, error)
	RecoverStuckJobs() error
	ResetStuckJob(uploadID uint) error
}

type bulkUploadService struct {
	bulkUploadLogRepo repository.BulkUploadLogRepository
	fileStorage       utils.FileStorage
	jobQueueClient    *utils.JobQueueClient
	config            *config.Config
	logger            *zap.Logger
}

func NewBulkUploadService(
	bulkUploadLogRepo repository.BulkUploadLogRepository,
	fileStorage utils.FileStorage,
	jobQueueClient *utils.JobQueueClient,
	cfg *config.Config,
	logger *zap.Logger,
) BulkUploadService {
	return &bulkUploadService{
		bulkUploadLogRepo: bulkUploadLogRepo,
		fileStorage:       fileStorage,
		jobQueueClient:    jobQueueClient,
		config:            cfg,
		logger:            logger,
	}
}

func (s *bulkUploadService) UploadCSV(entityType string, file *multipart.FileHeader, userID uint) (*models.BulkUploadLog, error) {
	// Validate entity type
	entityTypeLower := strings.ToLower(entityType)
	var validEntityType models.BulkUploadEntityType
	switch entityTypeLower {
	case "route":
		validEntityType = models.BulkUploadEntityTypeRoute
	case "stop":
		validEntityType = models.BulkUploadEntityTypeStop
	case "vehicle":
		validEntityType = models.BulkUploadEntityTypeVehicle
	default:
		return nil, fmt.Errorf("invalid entity type: %s. Must be 'route', 'stop', or 'vehicle'", entityType)
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".csv" {
		return nil, fmt.Errorf("invalid file type: %s. Only CSV files are allowed", ext)
	}

	// Save file to storage
	filePath, err := s.saveUploadFile(file, validEntityType)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Create bulk upload log entry
	uploadLog := &models.BulkUploadLog{
		EntityType:      validEntityType,
		FilePath:        filePath,
		Status:          models.BulkUploadStatusPending,
		UserID:          userID,
		LastProcessedRow: 0,
		LastUpdatedAt:   time.Now(),
	}

	if err := s.bulkUploadLogRepo.Create(uploadLog); err != nil {
		return nil, fmt.Errorf("failed to create upload log: %w", err)
	}

	// Queue background job
	jobID, err := s.jobQueueClient.EnqueueBulkUploadJob(uploadLog.ID)
	if err != nil {
		// Update status to failed
		uploadLog.Status = models.BulkUploadStatusFailed
		errorMsg := fmt.Sprintf("Failed to queue job: %v", err)
		uploadLog.ErrorMessage = &errorMsg
		s.bulkUploadLogRepo.Update(uploadLog)
		return nil, fmt.Errorf("failed to queue job: %w", err)
	}

	// Update log with job ID
	uploadLog.JobID = &jobID
	if err := s.bulkUploadLogRepo.Update(uploadLog); err != nil {
		s.logger.Warn("Failed to update upload log with job ID", zap.Error(err))
	}

	return uploadLog, nil
}

func (s *bulkUploadService) saveUploadFile(file *multipart.FileHeader, entityType models.BulkUploadEntityType) (string, error) {
	// Create a temporary ID for the file (we'll use timestamp)
	timestamp := time.Now().Unix()
	entityID := fmt.Sprintf("bulk_%d", timestamp)

	// Save file using file storage
	fileURL, err := s.fileStorage.SaveFile(file, "bulk_uploads", entityID)
	if err != nil {
		return "", err
	}

	// Extract relative path from URL
	// URL format: http://host:port/uploads/bulk_uploads/{entityID}/{filename}
	// We need to store the relative path for processing
	baseURL := fmt.Sprintf("http://%s:%s", s.config.Server.Host, s.config.Server.Port)
	if s.config.Server.Environment == "production" {
		if allowedOrigin := s.config.Server.AllowedOrigin; allowedOrigin != "*" && allowedOrigin != "" {
			baseURL = allowedOrigin
		}
	}

	// Extract relative path
	relativePath := strings.TrimPrefix(fileURL, baseURL+"/uploads/")
	
	// Return full path relative to uploads directory
	return filepath.Join(s.config.FileStorage.UploadPath, relativePath), nil
}

func (s *bulkUploadService) GetUploadStatus(uploadID uint) (*models.BulkUploadLog, error) {
	return s.bulkUploadLogRepo.FindByID(uploadID)
}

func (s *bulkUploadService) ListUploads(offset, limit int, filters map[string]interface{}) ([]models.BulkUploadLog, int64, error) {
	return s.bulkUploadLogRepo.List(offset, limit, filters)
}

func (s *bulkUploadService) RecoverStuckJobs() error {
	thresholdMinutes := s.config.JobQueue.StuckJobThresholdMinutes
	stuckJobs, err := s.bulkUploadLogRepo.FindStuckJobs(thresholdMinutes)
	if err != nil {
		return fmt.Errorf("failed to find stuck jobs: %w", err)
	}

	if len(stuckJobs) == 0 {
		s.logger.Info("No stuck jobs found")
		return nil
	}

	s.logger.Info("Recovering stuck jobs", zap.Int("count", len(stuckJobs)))

	for _, job := range stuckJobs {
		if err := s.ResetStuckJob(job.ID); err != nil {
			s.logger.Error("Failed to reset stuck job",
				zap.Uint("upload_id", job.ID),
				zap.Error(err),
			)
			continue
		}
	}

	return nil
}

func (s *bulkUploadService) ResetStuckJob(uploadID uint) error {
	uploadLog, err := s.bulkUploadLogRepo.FindByID(uploadID)
	if err != nil {
		return fmt.Errorf("failed to find upload log: %w", err)
	}

	// Reset status to pending
	uploadLog.Status = models.BulkUploadStatusPending
	uploadLog.LastUpdatedAt = time.Now()
	// Keep LastProcessedRow so job can resume from where it left off

	// Re-queue job
	jobID, err := s.jobQueueClient.EnqueueBulkUploadJob(uploadID)
	if err != nil {
		return fmt.Errorf("failed to re-queue job: %w", err)
	}

	uploadLog.JobID = &jobID
	if err := s.bulkUploadLogRepo.Update(uploadLog); err != nil {
		return fmt.Errorf("failed to update upload log: %w", err)
	}

	s.logger.Info("Reset stuck job",
		zap.Uint("upload_id", uploadID),
		zap.String("job_id", jobID),
		zap.Int("last_processed_row", uploadLog.LastProcessedRow),
	)

	return nil
}


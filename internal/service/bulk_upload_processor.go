package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"tj-routes/internal/cache"
	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
	"tj-routes/internal/utils"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	batchSize = 100 // Process records in batches
)

type BulkUploadProcessor struct {
	bulkUploadLogRepo repository.BulkUploadLogRepository
	routeRepo         repository.RouteRepository
	stopRepo          repository.StopRepository
	vehicleRepo       repository.VehicleRepository
	cache             cache.Cache
	config            *config.Config
	logger            *zap.Logger
}

func NewBulkUploadProcessor(
	bulkUploadLogRepo repository.BulkUploadLogRepository,
	routeRepo repository.RouteRepository,
	stopRepo repository.StopRepository,
	vehicleRepo repository.VehicleRepository,
	cacheInstance cache.Cache,
	cfg *config.Config,
	logger *zap.Logger,
) *BulkUploadProcessor {
	return &BulkUploadProcessor{
		bulkUploadLogRepo: bulkUploadLogRepo,
		routeRepo:         routeRepo,
		stopRepo:          stopRepo,
		vehicleRepo:       vehicleRepo,
		cache:             cacheInstance,
		config:            cfg,
		logger:            logger,
	}
}

// ProcessBulkUpload processes a bulk upload job
func (p *BulkUploadProcessor) ProcessBulkUpload(ctx context.Context, task *asynq.Task) error {
	var payload utils.BulkUploadJobPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	uploadID := payload.UploadID

	// Get upload log
	uploadLog, err := p.bulkUploadLogRepo.FindByID(uploadID)
	if err != nil {
		return fmt.Errorf("failed to find upload log: %w", err)
	}

	// Update status to processing
	uploadLog.Status = models.BulkUploadStatusProcessing
	uploadLog.LastUpdatedAt = time.Now()
	if err := p.bulkUploadLogRepo.Update(uploadLog); err != nil {
		return fmt.Errorf("failed to update upload log: %w", err)
	}

	// Process based on entity type
	switch uploadLog.EntityType {
	case models.BulkUploadEntityTypeRoute:
		err = p.processRoutes(uploadLog)
	case models.BulkUploadEntityTypeStop:
		err = p.processStops(uploadLog)
	case models.BulkUploadEntityTypeVehicle:
		err = p.processVehicles(uploadLog)
	default:
		err = fmt.Errorf("unknown entity type: %s", uploadLog.EntityType)
	}

	// Update final status
	if err != nil {
		uploadLog.Status = models.BulkUploadStatusFailed
		errorMsg := err.Error()
		uploadLog.ErrorMessage = &errorMsg
	} else {
		uploadLog.Status = models.BulkUploadStatusCompleted
	}
	uploadLog.LastUpdatedAt = time.Now()
	p.bulkUploadLogRepo.Update(uploadLog)

	// Invalidate cache
	ctx2 := context.Background()
	p.cache.InvalidatePattern(ctx2, cache.RoutePattern())
	p.cache.InvalidatePattern(ctx2, cache.StopPattern())
	p.cache.InvalidatePattern(ctx2, cache.VehiclePattern())

	return err
}

func (p *BulkUploadProcessor) processRoutes(uploadLog *models.BulkUploadLog) error {
	// Parse CSV
	rows, err := utils.ParseRoutesCSV(uploadLog.FilePath)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	uploadLog.TotalRows = len(rows)
	startRow := uploadLog.LastProcessedRow

	// Process in batches
	for i := startRow; i < len(rows); i += batchSize {
		endRow := i + batchSize
		if endRow > len(rows) {
			endRow = len(rows)
		}

		batch := rows[i:endRow]
		successCount := 0
		duplicateCount := 0
		errorCount := 0

		for _, row := range batch {
			// Check for duplicate
			existing, err := p.routeRepo.FindByRouteNumber(row.RouteNumber)
			if err == nil && existing != nil {
				duplicateCount++
				uploadLog.DuplicateCount++
				continue
			}
			if err != nil && err != gorm.ErrRecordNotFound {
				errorCount++
				uploadLog.ErrorCount++
				p.logger.Warn("Error checking duplicate route",
					zap.String("route_number", row.RouteNumber),
					zap.Error(err),
				)
				continue
			}

			// Create new route
			route := &models.Route{
				RouteNumber: row.RouteNumber,
				Name:        row.Name,
				Description: row.Description,
				Status:      models.Status(row.Status),
			}
			if route.Status == "" {
				route.Status = models.StatusActive
			}

			if err := p.routeRepo.Create(route); err != nil {
				errorCount++
				uploadLog.ErrorCount++
				p.logger.Warn("Failed to create route",
					zap.String("route_number", row.RouteNumber),
					zap.Error(err),
				)
				continue
			}

			successCount++
			uploadLog.SuccessCount++
		}

		// Update progress after each batch
		uploadLog.LastProcessedRow = endRow
		uploadLog.LastUpdatedAt = time.Now()
		if err := p.bulkUploadLogRepo.Update(uploadLog); err != nil {
			p.logger.Error("Failed to update progress", zap.Error(err))
		}

		p.logger.Info("Processed batch",
			zap.Uint("upload_id", uploadLog.ID),
			zap.Int("start_row", i),
			zap.Int("end_row", endRow),
			zap.Int("success", successCount),
			zap.Int("duplicate", duplicateCount),
			zap.Int("error", errorCount),
		)
	}

	return nil
}

func (p *BulkUploadProcessor) processStops(uploadLog *models.BulkUploadLog) error {
	// Parse CSV
	rows, err := utils.ParseStopsCSV(uploadLog.FilePath)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	uploadLog.TotalRows = len(rows)
	startRow := uploadLog.LastProcessedRow

	// Process in batches
	for i := startRow; i < len(rows); i += batchSize {
		endRow := i + batchSize
		if endRow > len(rows) {
			endRow = len(rows)
		}

		batch := rows[i:endRow]
		successCount := 0
		duplicateCount := 0
		errorCount := 0

		for _, row := range batch {
			// Check for duplicate
			existing, err := p.stopRepo.FindByLatitudeAndLongitude(row.Latitude, row.Longitude)
			if err == nil && existing != nil {
				duplicateCount++
				uploadLog.DuplicateCount++
				continue
			}
			if err != nil && err != gorm.ErrRecordNotFound {
				errorCount++
				uploadLog.ErrorCount++
				p.logger.Warn("Error checking duplicate stop",
					zap.Float64("latitude", row.Latitude),
					zap.Float64("longitude", row.Longitude),
					zap.Error(err),
				)
				continue
			}

			// Create new stop
			stop := &models.Stop{
				Name:      row.Name,
				Type:      models.StopType(row.Type),
				Latitude:  row.Latitude,
				Longitude: row.Longitude,
				Address:   row.Address,
				City:      row.City,
				District:  row.District,
				Status:    models.Status(row.Status),
			}
			if stop.Status == "" {
				stop.Status = models.StatusActive
			}

			// Handle facilities JSON
			if row.Facilities != "" {
				facilitiesJSON, err := utils.ValidateFacilitiesJSON(row.Facilities)
				if err != nil {
					p.logger.Warn("Invalid facilities JSON, skipping",
						zap.String("stop_name", row.Name),
						zap.Error(err),
					)
				} else {
					stop.Facilities = facilitiesJSON
				}
			}

			if err := p.stopRepo.Create(stop); err != nil {
				errorCount++
				uploadLog.ErrorCount++
				p.logger.Warn("Failed to create stop",
					zap.String("name", row.Name),
					zap.Error(err),
				)
				continue
			}

			successCount++
			uploadLog.SuccessCount++
		}

		// Update progress after each batch
		uploadLog.LastProcessedRow = endRow
		uploadLog.LastUpdatedAt = time.Now()
		if err := p.bulkUploadLogRepo.Update(uploadLog); err != nil {
			p.logger.Error("Failed to update progress", zap.Error(err))
		}

		p.logger.Info("Processed batch",
			zap.Uint("upload_id", uploadLog.ID),
			zap.Int("start_row", i),
			zap.Int("end_row", endRow),
			zap.Int("success", successCount),
			zap.Int("duplicate", duplicateCount),
			zap.Int("error", errorCount),
		)
	}

	return nil
}

func (p *BulkUploadProcessor) processVehicles(uploadLog *models.BulkUploadLog) error {
	// Parse CSV
	rows, err := utils.ParseVehiclesCSV(uploadLog.FilePath)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	uploadLog.TotalRows = len(rows)
	startRow := uploadLog.LastProcessedRow

	// Process in batches
	for i := startRow; i < len(rows); i += batchSize {
		endRow := i + batchSize
		if endRow > len(rows) {
			endRow = len(rows)
		}

		batch := rows[i:endRow]
		successCount := 0
		duplicateCount := 0
		errorCount := 0

		for _, row := range batch {
			// Check for duplicate
			existing, err := p.vehicleRepo.FindByVehiclePlate(row.VehiclePlate)
			if err == nil && existing != nil {
				duplicateCount++
				uploadLog.DuplicateCount++
				continue
			}
			if err != nil && err != gorm.ErrRecordNotFound {
				errorCount++
				uploadLog.ErrorCount++
				p.logger.Warn("Error checking duplicate vehicle",
					zap.String("vehicle_plate", row.VehiclePlate),
					zap.Error(err),
				)
				continue
			}

			// Verify route exists
			_, err = p.routeRepo.FindByID(row.RouteID)
			if err != nil {
				errorCount++
				uploadLog.ErrorCount++
				p.logger.Warn("Route not found for vehicle",
					zap.String("vehicle_plate", row.VehiclePlate),
					zap.Uint("route_id", row.RouteID),
					zap.Error(err),
				)
				continue
			}

			// Create new vehicle
			vehicle := &models.Vehicle{
				VehiclePlate: row.VehiclePlate,
				RouteID:      row.RouteID,
				VehicleType:  row.VehicleType,
				Capacity:     row.Capacity,
				Status:       models.Status(row.Status),
			}
			if vehicle.Status == "" {
				vehicle.Status = models.StatusActive
			}

			if err := p.vehicleRepo.Create(vehicle); err != nil {
				errorCount++
				uploadLog.ErrorCount++
				p.logger.Warn("Failed to create vehicle",
					zap.String("vehicle_plate", row.VehiclePlate),
					zap.Error(err),
				)
				continue
			}

			successCount++
			uploadLog.SuccessCount++
		}

		// Update progress after each batch
		uploadLog.LastProcessedRow = endRow
		uploadLog.LastUpdatedAt = time.Now()
		if err := p.bulkUploadLogRepo.Update(uploadLog); err != nil {
			p.logger.Error("Failed to update progress", zap.Error(err))
		}

		p.logger.Info("Processed batch",
			zap.Uint("upload_id", uploadLog.ID),
			zap.Int("start_row", i),
			zap.Int("end_row", endRow),
			zap.Int("success", successCount),
			zap.Int("duplicate", duplicateCount),
			zap.Int("error", errorCount),
		)
	}

	return nil
}


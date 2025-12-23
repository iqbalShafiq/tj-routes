package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"tj-routes/internal/config"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// Job type for bulk upload processing
	JobTypeBulkUpload = "bulk:upload"
)

// BulkUploadJobPayload represents the payload for bulk upload jobs
type BulkUploadJobPayload struct {
	UploadID uint `json:"upload_id"`
}

// JobQueueClient wraps asynq client
type JobQueueClient struct {
	client *asynq.Client
}

// JobQueueServer wraps asynq server
type JobQueueServer struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

var (
	jobQueueClient *JobQueueClient
	jobQueueServer *JobQueueServer
)

// InitJobQueueClient initializes the asynq client
func InitJobQueueClient(cfg *config.Config) (*JobQueueClient, error) {
	// Test Redis connection first with retry logic
	testClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	defer testClient.Close()

	// Retry connection up to 3 times with exponential backoff
	var lastErr error
	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := testClient.Ping(ctx).Err()
		cancel()
		
		if err == nil {
			break // Success
		}
		
		lastErr = err
		if i < 2 { // Don't sleep on last attempt
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to connect to Redis for job queue after retries (host=%s, port=%s): %w", cfg.Redis.Host, cfg.Redis.Port, lastErr)
	}

	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	client := asynq.NewClient(redisOpt)
	jobQueueClient = &JobQueueClient{client: client}
	return jobQueueClient, nil
}

// InitJobQueueServer initializes the asynq server and mux
func InitJobQueueServer(cfg *config.Config, logger *zap.Logger) (*JobQueueServer, error) {
	// Test Redis connection first with retry logic
	testClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	defer testClient.Close()

	// Retry connection up to 3 times with exponential backoff
	var lastErr error
	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := testClient.Ping(ctx).Err()
		cancel()
		
		if err == nil {
			break // Success
		}
		
		lastErr = err
		if i < 2 { // Don't sleep on last attempt
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to connect to Redis for job queue server after retries (host=%s, port=%s): %w", cfg.Redis.Host, cfg.Redis.Port, lastErr)
	}

	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: cfg.JobQueue.Concurrency,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			StrictPriority: false,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.Error("Job processing error",
					zap.String("type", task.Type()),
					zap.Error(err),
				)
			}),
		},
	)

	mux := asynq.NewServeMux()
	jobQueueServer = &JobQueueServer{server: server, mux: mux}
	return jobQueueServer, nil
}

// EnqueueBulkUploadJob enqueues a bulk upload job
func (c *JobQueueClient) EnqueueBulkUploadJob(uploadID uint) (string, error) {
	payload, err := json.Marshal(BulkUploadJobPayload{UploadID: uploadID})
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(JobTypeBulkUpload, payload, asynq.Queue("default"))
	info, err := c.client.Enqueue(task)
	if err != nil {
		return "", fmt.Errorf("failed to enqueue task: %w", err)
	}

	return info.ID, nil
}

// RegisterBulkUploadHandler registers the handler for bulk upload jobs
func (s *JobQueueServer) RegisterBulkUploadHandler(handler func(context.Context, *asynq.Task) error) {
	s.mux.HandleFunc(JobTypeBulkUpload, handler)
}

// Start starts the job queue server
func (s *JobQueueServer) Start() error {
	return s.server.Start(s.mux)
}

// Shutdown gracefully shuts down the job queue server
func (s *JobQueueServer) Shutdown() {
	s.server.Shutdown()
}

// GetClient returns the global job queue client
func GetJobQueueClient() *JobQueueClient {
	return jobQueueClient
}

// GetServer returns the global job queue server
func GetJobQueueServer() *JobQueueServer {
	return jobQueueServer
}

// CancelJob cancels a job by ID
func (c *JobQueueClient) CancelJob(jobID string) error {
	// Note: asynq doesn't have a direct cancel method for already enqueued jobs
	// Jobs will be automatically retried or archived based on their retry policy
	// For cancellation, we rely on the job processor checking the status in the database
	return nil
}


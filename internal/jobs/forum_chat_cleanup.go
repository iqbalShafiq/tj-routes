package jobs

import (
	"context"
	"time"

	"tj-routes/internal/repository"

	"go.uber.org/zap"
)

type ForumChatCleanupJob struct {
	forumMessageRepo repository.ForumMessageRepository
	logger           *zap.Logger
}

func NewForumChatCleanupJob(
	forumMessageRepo repository.ForumMessageRepository,
	logger *zap.Logger,
) *ForumChatCleanupJob {
	return &ForumChatCleanupJob{
		forumMessageRepo: forumMessageRepo,
		logger:           logger,
	}
}

func (j *ForumChatCleanupJob) Execute(ctx context.Context) error {
	cutoff := time.Now().Add(-24 * time.Hour)

	err := j.forumMessageRepo.DeleteOlderThan(cutoff)
	if err != nil {
		j.logger.Error("Failed to delete old forum messages",
			zap.Time("cutoff", cutoff),
			zap.Error(err),
		)
		return err
	}

	j.logger.Info("Deleted old forum messages",
		zap.Time("cutoff", cutoff),
	)

	return nil
}

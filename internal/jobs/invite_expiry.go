package jobs

import (
	"context"

	"tj-routes/internal/repository"

	"go.uber.org/zap"
)

type InviteExpiryJob struct {
	groupInviteRepo repository.GroupInviteRepository
	logger          *zap.Logger
}

func NewInviteExpiryJob(
	groupInviteRepo repository.GroupInviteRepository,
	logger *zap.Logger,
) *InviteExpiryJob {
	return &InviteExpiryJob{
		groupInviteRepo: groupInviteRepo,
		logger:          logger,
	}
}

func (j *InviteExpiryJob) Execute(ctx context.Context) error {
	err := j.groupInviteRepo.MarkExpired()
	if err != nil {
		j.logger.Error("Failed to mark expired group invites",
			zap.Error(err),
		)
		return err
	}

	j.logger.Info("Marked expired group invites successfully")
	return nil
}

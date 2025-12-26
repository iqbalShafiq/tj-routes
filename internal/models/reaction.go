package models

import (
	"time"
)

type ReactionType string

const (
	ReactionUpvote   ReactionType = "upvote"
	ReactionDownvote ReactionType = "downvote"
)

type ReactionTargetType string

const (
	ReactionTargetReport    ReactionTargetType = "report"
	ReactionTargetComment   ReactionTargetType = "comment"
	ReactionTargetForumPost ReactionTargetType = "forum_post"
)

type Reaction struct {
	ID           uint              `gorm:"primaryKey" json:"id"`
	UserID       uint              `gorm:"not null;index" json:"user_id"`
	TargetType   ReactionTargetType `gorm:"type:varchar(20);not null;index" json:"target_type"`
	TargetID     uint              `gorm:"not null;index" json:"target_id"`
	ReactionType ReactionType      `gorm:"type:varchar(20);not null" json:"reaction_type"`
	CreatedAt    time.Time         `json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Unique constraint: (user_id, target_type, target_id) - enforced at application level


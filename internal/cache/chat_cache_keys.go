package cache

import (
	"fmt"
	"time"
)

const DefaultTTL = 30 * time.Minute

func ConversationKey(id uint) string {
	return fmt.Sprintf("chat:conversation:%d", id)
}

func ConversationListKey(userID uint, page, limit int) string {
	return fmt.Sprintf("chat:user:%d:conversations:%d:%d", userID, page, limit)
}

func GroupChatKey(id uint) string {
	return fmt.Sprintf("chat:group:%d", id)
}

func GroupMembersKey(groupID uint) string {
	return fmt.Sprintf("chat:group:%d:members", groupID)
}

func GroupListKey(userID uint, page, limit int) string {
	return fmt.Sprintf("chat:user:%d:groups:%d:%d", userID, page, limit)
}

func ForumChatMessagesKey(forumID uint) string {
	return fmt.Sprintf("chat:forum:%d:messages", forumID)
}

func UserPresenceKey(userID uint) string {
	return fmt.Sprintf("chat:presence:user:%d", userID)
}

func OnlineUsersKey() string {
	return "chat:online:users"
}

func TypingIndicatorKey(conversationID uint) string {
	return fmt.Sprintf("chat:typing:%d", conversationID)
}

func UnreadCountKey(userID uint) string {
	return fmt.Sprintf("chat:user:%d:unread", userID)
}

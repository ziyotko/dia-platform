package service

import (
	"errors"
	"strings"
	"time"

	"base/internal/models"
	"base/pkg/db"
	"base/pkg/notifier"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MessageService struct{}

// broadcastReadExists 子查询：判断广播消息对指定用户是否已有已读记录。
// 广播消息（receiver_id = 0）在库中只有一行数据，行级 is_read 会被所有人共享，
// 因此广播的已读状态存放在 base_message_read，定向消息继续使用 base_message.is_read。
const broadcastReadExists = "EXISTS (SELECT 1 FROM base_message_read mr WHERE mr.message_id = base_message.id AND mr.user_id = ? AND mr.deleted_at IS NULL)"

// CreateDraft 新建草稿。草稿不会进入任何人的收件箱，发送需走 /messages/send（新建发送）
// 或 /messages/:id/send（发送已有草稿）。
func (s MessageService) CreateDraft(senderID uint64, senderName string, tenantID uint64, in DraftInput) (*models.Message, error) {
	updates, err := s.normalizeDraft(tenantID, in)
	if err != nil {
		return nil, err
	}
	m := models.Message{
		TenantID:     tenantID,
		SenderID:     senderID,
		SenderName:   senderName,
		ReceiverID:   in.ReceiverID,
		ReceiverType: updates["receiver_type"].(string),
		Title:        updates["title"].(string),
		Content:      in.Content,
		Type:         in.Type,
		Priority:     in.Priority,
		Status:       models.MessageStatusDraft,
	}
	if err := db.DB.Create(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// MessageListQuery 消息列表查询条件。
// Box=inbox 收件箱（别人发给我的 + 广播）；Box=sent 发件箱（我发出的，按 sender_id）。
type MessageListQuery struct {
	TenantID uint64
	UserID   uint64
	Box      string
	Status   int // -1 全部
	IsRead   int // -1 全部，0 未读，1 已读
	Page     int
	Size     int
}

// DraftInput 草稿内容（新建/编辑共用）。
// ReceiverID = 0 表示全员广播草稿。
type DraftInput struct {
	ReceiverID uint64
	Title      string
	Content    string
	Type       string
	Priority   string
}

// normalizeDraft 校验并归一化草稿字段：标题/内容必填；指定接收人时必须同租户且启用。
func (s MessageService) normalizeDraft(tenantID uint64, in DraftInput) (map[string]interface{}, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, errors.New("请填写标题")
	}
	content := strings.TrimSpace(in.Content)
	if content == "" {
		return nil, errors.New("请填写内容")
	}
	receiverType := "all"
	if in.ReceiverID != 0 {
		if _, err := s.recipients(tenantID, []uint64{in.ReceiverID}); err != nil {
			return nil, err
		}
		receiverType = "user"
	}
	return map[string]interface{}{
		"receiver_id":   in.ReceiverID,
		"receiver_type": receiverType,
		"title":         title,
		"content":       content,
		"type":          in.Type,
		"priority":      in.Priority,
	}, nil
}

// UpdateDraft 编辑本人的草稿（仅草稿状态可编辑）。
func (s MessageService) UpdateDraft(id, senderID, tenantID uint64, in DraftInput) error {
	updates, err := s.normalizeDraft(tenantID, in)
	if err != nil {
		return err
	}
	query := db.DB.Model(&models.Message{}).Where("id = ? AND sender_id = ? AND status = ?", id, senderID, models.MessageStatusDraft)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	// 先确认草稿归属与状态，避免「值没变」时把 RowsAffected=0 误判为失败
	if err := ensureRecordExists(db.DB.Model(&models.Message{}).
		Where("id = ? AND sender_id = ? AND status = ?", id, senderID, models.MessageStatusDraft), "草稿不存在或已发送"); err != nil {
		return err
	}
	return query.Updates(updates).Error
}

// SendDraft 发送本人的草稿：重新校验接收人后置为已发送。
func (s MessageService) SendDraft(id, senderID, tenantID uint64) error {
	var draft models.Message
	query := db.DB.Where("id = ? AND sender_id = ? AND status = ?", id, senderID, models.MessageStatusDraft)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&draft).Error; err != nil {
		return errors.New("草稿不存在或已发送")
	}
	if draft.ReceiverID != 0 {
		if _, err := s.recipients(tenantID, []uint64{draft.ReceiverID}); err != nil {
			return err
		}
	}
	now := time.Now()
	return db.DB.Model(&models.Message{}).Where("id = ?", draft.ID).
		Updates(map[string]interface{}{
			"status":  models.MessageStatusSent,
			"send_at": now,
		}).Error
}

// GetByID 按 ID 读取消息，并校验可见性：
// 只有发送者、接收者本人，或已发送的广播消息可以被读取，避免同租户内互读私信与草稿。
func (s MessageService) GetByID(id, userID, tenantID uint64) (*models.Message, error) {
	var m models.Message
	query := db.DB.Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	if err := query.First(&m).Error; err != nil {
		return &m, err
	}
	if !messageVisibleTo(&m, userID) {
		return &m, errors.New("消息不存在")
	}
	if m.ReceiverID == 0 {
		// 广播消息的行级 is_read 是共享的，需按当前用户回填
		list := []models.Message{m}
		s.fillBroadcastReadState(list, userID)
		m.IsRead = list[0].IsRead
		m.ReadAt = list[0].ReadAt
	}
	return &m, nil
}

// messageVisibleTo 判断消息对该用户是否可见。
// 广播消息（receiver_id = 0）对所有人可见，但草稿广播仅发送者可见。
func messageVisibleTo(m *models.Message, userID uint64) bool {
	if m.SenderID == userID || m.ReceiverID == userID {
		return true
	}
	return m.ReceiverID == 0 && m.Status == models.MessageStatusSent
}

// Delete 删除消息。权限口径：
//   - 定向消息：发送者或接收者本人可删；
//   - 广播消息：库中只有一行、租户内共享，普通接收者不可删，仅发送者或管理员可删。
func (s MessageService) Delete(id, userID uint64, isAdmin bool, tenantID uint64) error {
	var m models.Message
	query := db.DB.Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	if err := query.First(&m).Error; err != nil {
		return errors.New("消息不存在")
	}
	switch {
	case m.ReceiverID == 0:
		if m.SenderID != userID && !isAdmin {
			return errors.New("广播消息不支持单个接收者删除")
		}
	case m.SenderID != userID && m.ReceiverID != userID:
		return errors.New("只能删除自己发送或接收的消息")
	}
	return db.DB.Delete(&m).Error
}

func (s MessageService) List(q MessageListQuery) ([]models.Message, int64, error) {
	var list []models.Message
	var total int64
	query := db.DB.Model(&models.Message{})
	if q.TenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", q.TenantID)
	}
	if q.Box == "sent" {
		// 发件箱：我发出的消息（草稿与已发送均在列）
		query = query.Where("sender_id = ?", q.UserID)
	} else {
		// 收件箱：别人发给我或广播给我，且已发送（不含草稿）
		query = query.Where("receiver_id = ? OR receiver_id = 0", q.UserID).Where("status = ?", models.MessageStatusSent)
	}
	if q.Status >= 0 {
		query = query.Where("status = ?", q.Status)
	}
	if q.IsRead >= 0 {
		// 定向消息看 is_read；广播消息看 base_message_read 中是否有本人的已读记录
		if q.IsRead == 1 {
			query = query.Where("is_read = ? OR "+broadcastReadExists, true, q.UserID)
		} else {
			query = query.Where("is_read = ? AND NOT "+broadcastReadExists, false, q.UserID)
		}
	}
	query.Count(&total)
	if err := query.Order("created_at DESC").Offset((q.Page - 1) * q.Size).Limit(q.Size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	// 广播消息的行级 is_read 是共享字段，需替换为「当前用户是否已读」
	s.fillBroadcastReadState(list, q.UserID)
	return list, total, nil
}

// fillBroadcastReadState 将广播消息的 is_read（共享字段）替换为当前用户的真实已读状态。
func (s MessageService) fillBroadcastReadState(list []models.Message, userID uint64) {
	ids := make([]uint64, 0, len(list))
	for i := range list {
		if list[i].ReceiverID == 0 {
			ids = append(ids, list[i].ID)
		}
	}
	if len(ids) == 0 {
		return
	}
	var records []models.MessageRead
	if err := db.DB.Where("user_id = ? AND message_id IN ?", userID, ids).Find(&records).Error; err != nil {
		return
	}
	read := make(map[uint64]models.MessageRead, len(records))
	for _, r := range records {
		read[r.MessageID] = r
	}
	for i := range list {
		if list[i].ReceiverID == 0 {
			if r, ok := read[list[i].ID]; ok {
				list[i].IsRead = true
				readAt := r.ReadAt
				list[i].ReadAt = &readAt
			} else {
				list[i].IsRead = false
				list[i].ReadAt = nil
			}
		}
	}
}

// GetUnreadCount 当前用户未读数：
// 定向消息（receiver_id = 我）取 is_read，广播消息（receiver_id = 0）取 base_message_read 中无本人记录的那些。
func (s MessageService) GetUnreadCount(receiverID uint64, tenantID uint64) (int64, error) {
	direct := db.DB.Model(&models.Message{}).
		Where("receiver_id = ? AND status = ? AND is_read = ?", receiverID, models.MessageStatusSent, false)
	broadcast := db.DB.Model(&models.Message{}).
		Where("receiver_id = 0 AND status = ?", models.MessageStatusSent).
		Where("NOT "+broadcastReadExists, receiverID)
	if tenantID > 0 {
		direct = direct.Where("tenant_id = ? OR tenant_id = 0", tenantID)
		broadcast = broadcast.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	var directCount, broadcastCount int64
	if err := direct.Count(&directCount).Error; err != nil {
		return 0, err
	}
	if err := broadcast.Count(&broadcastCount).Error; err != nil {
		return 0, err
	}
	return directCount + broadcastCount, nil
}

// MarkRead 标记单条消息为已读（仅限发给自己的定向消息或广播消息）。
func (s MessageService) MarkRead(id, receiverID uint64, tenantID uint64) error {
	var m models.Message
	query := db.DB.Where("id = ? AND (receiver_id = ? OR receiver_id = 0)", id, receiverID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	if err := query.First(&m).Error; err != nil {
		return errors.New("消息不存在")
	}
	if m.ReceiverID == 0 {
		return s.markBroadcastRead(m.ID, receiverID)
	}
	return db.DB.Model(&models.Message{}).Where("id = ?", m.ID).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": time.Now(),
	}).Error
}

// markBroadcastRead 为广播消息写入「本人已读」记录（幂等）。
func (s MessageService) markBroadcastRead(messageID, userID uint64) error {
	var record models.MessageRead
	err := db.DB.Where("message_id = ? AND user_id = ?", messageID, userID).First(&record).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return db.DB.Create(&models.MessageRead{
		MessageID: messageID,
		UserID:    userID,
		ReadAt:    time.Now(),
	}).Error
}

// MarkAllRead 将当前用户的全部未读消息标为已读，返回影响条数（含广播的已读记录）。
func (s MessageService) MarkAllRead(receiverID uint64, tenantID uint64) (int64, error) {
	now := time.Now()

	direct := db.DB.Model(&models.Message{}).
		Where("receiver_id = ? AND status = ? AND is_read = ?", receiverID, models.MessageStatusSent, false)
	if tenantID > 0 {
		direct = direct.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	res := direct.Updates(map[string]interface{}{
		"is_read": true,
		"read_at": now,
	})
	if res.Error != nil {
		return 0, res.Error
	}
	affected := res.RowsAffected

	// 广播：为本人尚未读过的每条广播补一条已读记录
	broadcast := db.DB.Model(&models.Message{}).
		Where("receiver_id = 0 AND status = ?", models.MessageStatusSent).
		Where("NOT "+broadcastReadExists, receiverID)
	if tenantID > 0 {
		broadcast = broadcast.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	var pending []models.Message
	if err := broadcast.Select("id").Limit(1000).Find(&pending).Error; err != nil {
		return affected, err
	}
	if len(pending) == 0 {
		return affected, nil
	}
	records := make([]models.MessageRead, 0, len(pending))
	for _, m := range pending {
		records = append(records, models.MessageRead{MessageID: m.ID, UserID: receiverID, ReadAt: now})
	}
	if err := db.DB.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(records, 200).Error; err != nil {
		return affected, err
	}
	return affected + int64(len(records)), nil
}

// SendToUsers 定向发送：只允许发给「启用状态」且与发送者同租户的用户，
// 避免跨租户投递（收件人因 tenant_id 过滤看不到，却会真的收到邮件）。
func (s MessageService) SendToUsers(senderID uint64, senderName string, tenantID uint64, userIDs []uint64, title, content, msgType, priority string) error {
	users, err := s.recipients(tenantID, userIDs)
	if err != nil {
		return err
	}
	now := time.Now()
	msgs := make([]models.Message, 0, len(users))
	for _, u := range users {
		msgs = append(msgs, models.Message{
			TenantID:     tenantID,
			SenderID:     senderID,
			SenderName:   senderName,
			ReceiverID:   u.ID,
			ReceiverType: "user",
			Title:        title,
			Content:      content,
			Type:         msgType,
			Priority:     priority,
			Status:       models.MessageStatusSent,
			SendAt:       &now,
		})
	}
	return db.DB.CreateInBatches(msgs, 100).Error
}

// recipients 校验接收人：去重后必须全部存在、启用，且（非平台超管时）属于发送者所在租户。
func (s MessageService) recipients(tenantID uint64, userIDs []uint64) ([]models.User, error) {
	unique := make([]uint64, 0, len(userIDs))
	seen := make(map[uint64]struct{}, len(userIDs))
	for _, id := range userIDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return nil, errors.New("请选择接收用户")
	}
	query := db.DB.Model(&models.User{}).Where("id IN ? AND status = ?", unique, 1)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	if len(users) != len(unique) {
		return nil, errors.New("存在无效的接收用户（不存在、已停用或不属于当前租户）")
	}
	return users, nil
}

// Broadcast 全员广播：租户内共享一行数据，已读状态由 base_message_read 按人记录。
func (s MessageService) Broadcast(senderID uint64, senderName string, tenantID uint64, title, content, msgType, priority string) error {
	now := time.Now()
	return db.DB.Create(&models.Message{
		TenantID:     tenantID,
		SenderID:     senderID,
		SenderName:   senderName,
		ReceiverID:   0,
		ReceiverType: "all",
		Title:        title,
		Content:      content,
		Type:         msgType,
		Priority:     priority,
		Status:       models.MessageStatusSent,
		SendAt:       &now,
	}).Error
}

// SendExternal 根据渠道发送站外消息（邮件/短信/企微等）
func (s MessageService) SendExternal(channel, to, subject, content string) error {
	if channel == "" || channel == "in-app" {
		return nil
	}
	return notifier.Send(channel, notifier.Message{
		To:      to,
		Subject: subject,
		Body:    content,
	})
}

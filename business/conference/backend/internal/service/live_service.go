package service

import (
	"errors"
	"time"

	"conference/internal/models"
	"conference/pkg/db"

	"github.com/google/uuid"
)

type LiveService struct{}

// StartLive begins a live stream for a meeting
func (s *LiveService) StartLive(meetingID uint64) (*models.LiveConfig, error) {
	var config models.LiveConfig
	err := db.DB.Where("meeting_id = ?", meetingID).First(&config).Error
	streamKey := uuid.New().String()[:16]

	if err != nil {
		// Create new config
		config = models.LiveConfig{
			MeetingID: meetingID,
			StreamKey: streamKey,
			PushURL:   "rtmp://live.example.com/live/" + streamKey,
			PlayURL:   "http://live.example.com/hls/" + streamKey + ".m3u8",
			Status:    models.LiveStatusLive,
		}
		now := time.Now()
		config.StartedAt = &now
		if err := db.DB.Create(&config).Error; err != nil {
			return nil, errors.New("开启直播失败")
		}
	} else {
		now := time.Now()
		if err := db.DB.Model(&config).Updates(map[string]interface{}{
			"status":     models.LiveStatusLive,
			"started_at": &now,
			"stream_key": streamKey,
			"push_url":   "rtmp://live.example.com/live/" + streamKey,
			"play_url":   "http://live.example.com/hls/" + streamKey + ".m3u8",
		}).Error; err != nil {
			return nil, errors.New("开启直播失败")
		}
		config.Status = models.LiveStatusLive
	}

	return &config, nil
}

// StopLive ends a live stream
func (s *LiveService) StopLive(meetingID uint64) error {
	now := time.Now()
	return db.DB.Model(&models.LiveConfig{}).Where("meeting_id = ?", meetingID).Updates(map[string]interface{}{
		"status":   models.LiveStatusEnded,
		"ended_at": &now,
	}).Error
}

// GetLiveConfig returns live config for a meeting
func (s *LiveService) GetLiveConfig(meetingID uint64) (*models.LiveConfig, error) {
	var config models.LiveConfig
	if err := db.DB.Where("meeting_id = ?", meetingID).First(&config).Error; err != nil {
		return nil, errors.New("直播配置不存在")
	}
	return &config, nil
}

// GetPlayURL returns the play URL if user has access
func (s *LiveService) GetPlayURL(meetingID, userID uint64) (string, error) {
	// Verify user registered for this meeting
	var reg models.Registration
	if err := db.DB.Where("meeting_id = ? AND user_id = ? AND status = ?",
		meetingID, userID, models.RegStatusApproved).First(&reg).Error; err != nil {
		return "", errors.New("您未报名该会议或报名未通过，无法观看直播")
	}

	var config models.LiveConfig
	if err := db.DB.Where("meeting_id = ?", meetingID).First(&config).Error; err != nil {
		return "", errors.New("直播未开启")
	}
	if config.Status == models.LiveStatusOff {
		return "", errors.New("直播未开启")
	}
	return config.PlayURL, nil
}

// SendLiveMessage sends a chat message during live stream
func (s *LiveService) SendLiveMessage(meetingID, userID uint64, content string) error {
	if content == "" {
		return errors.New("消息不能为空")
	}
	msg := models.LiveMessage{
		MeetingID: meetingID,
		UserID:    userID,
		Content:   content,
	}
	return db.DB.Create(&msg).Error
}

// GetLiveMessages returns recent live messages
func (s *LiveService) GetLiveMessages(meetingID uint64, limit int) ([]models.LiveMessage, error) {
	var msgs []models.LiveMessage
	if limit <= 0 {
		limit = 50
	}
	err := db.DB.Preload("User").Where("meeting_id = ?", meetingID).
		Order("created_at DESC").Limit(limit).Find(&msgs).Error
	return msgs, err
}

// RecordViewingLog records viewing start
func (s *LiveService) RecordViewingStart(meetingID, userID uint64, viewType string) error {
	now := time.Now()
	log := models.ViewingLog{
		MeetingID: meetingID,
		UserID:    userID,
		StartTime: &now,
		Type:      viewType,
	}
	return db.DB.Create(&log).Error
}

// RecordViewingEnd records viewing end and calculates duration
func (s *LiveService) RecordViewingEnd(logID uint64) error {
	var log models.ViewingLog
	if err := db.DB.First(&log, logID).Error; err != nil {
		return errors.New("观看记录不存在")
	}
	now := time.Now()
	duration := int(now.Sub(*log.StartTime).Seconds())
	return db.DB.Model(&log).Updates(map[string]interface{}{
		"end_time": &now,
		"duration": duration,
	}).Error
}

// GetViewingLogs returns viewing logs for a meeting
func (s *LiveService) GetViewingLogs(meetingID uint64, page, size int) ([]models.ViewingLog, int64, error) {
	var list []models.ViewingLog
	var total int64
	query := db.DB.Model(&models.ViewingLog{}).Preload("User")
	if meetingID > 0 {
		query = query.Where("meeting_id = ?", meetingID)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// --- VOD ---

func (s *LiveService) UploadVod(meetingID uint64, videoURL string, replayExpiry *time.Time) error {
	var config models.VodConfig
	err := db.DB.Where("meeting_id = ?", meetingID).First(&config).Error
	if err != nil {
		config = models.VodConfig{
			MeetingID:    meetingID,
			VideoURL:     videoURL,
			AllowReplay:  true,
			ReplayExpiry: replayExpiry,
			Status:       "published",
		}
		return db.DB.Create(&config).Error
	}
	return db.DB.Model(&config).Updates(map[string]interface{}{
		"video_url":     videoURL,
		"allow_replay":  true,
		"replay_expiry": replayExpiry,
		"status":        "published",
	}).Error
}

func (s *LiveService) GetVodConfig(meetingID uint64) (*models.VodConfig, error) {
	var config models.VodConfig
	if err := db.DB.Where("meeting_id = ?", meetingID).First(&config).Error; err != nil {
		return nil, errors.New("录播不存在")
	}
	return &config, nil
}

func (s *LiveService) SetReplayStatus(meetingID uint64, allow bool) error {
	return db.DB.Model(&models.VodConfig{}).Where("meeting_id = ?", meetingID).
		Update("allow_replay", allow).Error
}

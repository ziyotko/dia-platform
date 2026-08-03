package service

import (
	"encoding/json"
	"errors"
	"time"

	"conference/internal/models"
	"conference/pkg/db"

	"gorm.io/gorm"
)

type VoteService struct{}

func (s *VoteService) Create(vote *models.Vote, options []models.VoteOption) error {
	tx := db.DB.Begin()
	if err := tx.Create(vote).Error; err != nil {
		tx.Rollback()
		return errors.New("创建投票失败")
	}
	for i := range options {
		options[i].VoteID = vote.ID
		options[i].ID = 0
	}
	if err := tx.Create(&options).Error; err != nil {
		tx.Rollback()
		return errors.New("创建选项失败")
	}
	return tx.Commit().Error
}

func (s *VoteService) Update(id uint64, updates map[string]interface{}) error {
	return db.DB.Model(&models.Vote{}).Where("id = ?", id).Updates(updates).Error
}

func (s *VoteService) Delete(id uint64) error {
	tx := db.DB.Begin()
	tx.Where("vote_id = ?", id).Delete(&models.VoteOption{})
	tx.Where("vote_id = ?", id).Delete(&models.VoteRecord{})
	tx.Delete(&models.Vote{}, id)
	return tx.Commit().Error
}

func (s *VoteService) GetByID(id uint64) (*models.Vote, error) {
	var v models.Vote
	if err := db.DB.Preload("Options", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).First(&v, id).Error; err != nil {
		return nil, errors.New("投票不存在")
	}
	return &v, nil
}

func (s *VoteService) List(meetingID uint64, status string, page, size int) ([]models.Vote, int64, error) {
	var list []models.Vote
	var total int64
	query := db.DB.Model(&models.Vote{})
	if meetingID > 0 {
		query = query.Where("meeting_id = ?", meetingID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s *VoteService) ListAvailable(userID uint64) ([]models.Vote, error) {
	now := time.Now()
	var list []models.Vote
	err := db.DB.Where("status = ? AND start_time <= ? AND end_time >= ?", "open", now, now).
		Order("start_time ASC").Find(&list).Error
	return list, err
}

func (s *VoteService) CastVote(voteID, userID uint64, optionID uint64, optionIDs []uint64) error {
	var vote models.Vote
	if err := db.DB.First(&vote, voteID).Error; err != nil {
		return errors.New("投票不存在")
	}
	if vote.Status != "open" {
		return errors.New("投票未开放")
	}
	now := time.Now()
	if vote.StartTime != nil && now.Before(*vote.StartTime) {
		return errors.New("投票尚未开始")
	}
	if vote.EndTime != nil && now.After(*vote.EndTime) {
		return errors.New("投票已结束")
	}

	// Check not already voted
	var count int64
	db.DB.Model(&models.VoteRecord{}).Where("vote_id = ? AND user_id = ?", voteID, userID).Count(&count)
	if count > 0 {
		return errors.New("您已投过票，不可重复投票")
	}

	record := models.VoteRecord{
		VoteID:  voteID,
		UserID:  userID,
		VotedAt: &now,
	}

	if vote.Type == models.VoteTypeSingle || vote.Type == models.VoteTypeEqual {
		record.OptionID = optionID
	} else {
		ids, _ := json.Marshal(optionIDs)
		record.OptionIDs = string(ids)
	}

	if err := db.DB.Create(&record).Error; err != nil {
		return errors.New("投票失败")
	}

	// Update option counts
	if vote.Type == models.VoteTypeSingle || vote.Type == models.VoteTypeEqual {
		db.DB.Model(&models.VoteOption{}).Where("id = ?", optionID).
			UpdateColumn("vote_count", gorm.Expr("vote_count + ?", 1))
	} else {
		for _, oid := range optionIDs {
			db.DB.Model(&models.VoteOption{}).Where("id = ?", oid).
				UpdateColumn("vote_count", gorm.Expr("vote_count + ?", 1))
		}
	}
	return nil
}

func (s *VoteService) GetResults(voteID uint64) (map[string]interface{}, error) {
	var vote models.Vote
	if err := db.DB.Preload("Options", func(db *gorm.DB) *gorm.DB {
		return db.Order("vote_count DESC")
	}).First(&vote, voteID).Error; err != nil {
		return nil, errors.New("投票不存在")
	}

	var totalVotes int64
	db.DB.Model(&models.VoteRecord{}).Where("vote_id = ?", voteID).Count(&totalVotes)

	return map[string]interface{}{
		"vote":        vote,
		"total_votes": totalVotes,
		"options":     vote.Options,
	}, nil
}

func (s *VoteService) HasUserVoted(voteID, userID uint64) bool {
	var count int64
	db.DB.Model(&models.VoteRecord{}).Where("vote_id = ? AND user_id = ?", voteID, userID).Count(&count)
	return count > 0
}

func (s *VoteService) GetMyVoteRecord(voteID, userID uint64) (*models.VoteRecord, error) {
	var record models.VoteRecord
	if err := db.DB.Where("vote_id = ? AND user_id = ?", voteID, userID).First(&record).Error; err != nil {
		return nil, errors.New("未投票")
	}
	return &record, nil
}

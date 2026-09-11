package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
	"member/pkg/utils"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type MemberService struct{}

// checkableMemberFields 允许查重的会员字段（防止 SQL 注入）
var checkableMemberFields = map[string]bool{
	"username":       true,
	"mobile":         true,
	"email":          true,
	"contact_mobile": true,
	"company_name":   true,
	"credit_code":    true,
}

// CheckFieldExists 校验会员字段值是否已被占用（admin 新增会员查重）
func (s *MemberService) CheckFieldExists(field, value string) (bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return false, nil
	}
	if !checkableMemberFields[field] {
		return false, errors.New("不支持的字段")
	}
	if field == "credit_code" {
		value = strings.ToUpper(value)
	}
	var count int64
	if err := db.DB.Model(&models.Member{}).Where(field+" = ?", value).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetMember returns a member by ID
func (s *MemberService) GetMember(id uint64) (*models.Member, error) {
	var m models.Member
	if err := db.DB.First(&m, id).Error; err != nil {
		return nil, errors.New("会员不存在")
	}
	// 填充主入会机构名称
	members := []models.Member{m}
	_ = s.fillOrgNames(members)
	m.OrgName = members[0].OrgName
	return &m, nil
}

// fillOrgNames fills each member's OrgName with its primary joined organization
// (主入会机构). Priority:
//  1. 该会员进行中的入会申请（待审核/已通过）所选机构；
//  2. 兜底：最近一条已缴费记录所属机构（覆盖无申请记录的历史/管理员直录会员）。
func (s *MemberService) fillOrgNames(members []models.Member) error {
	if len(members) == 0 {
		return nil
	}
	ids := make([]uint64, 0, len(members))
	index := make(map[uint64]*models.Member, len(members))
	for i := range members {
		ids = append(ids, members[i].ID)
		index[members[i].ID] = &members[i]
	}

	// 1) 入会申请（待审核/已通过）对应机构；同一会员同时至多存在一个进行中的申请
	var appRows []struct {
		MemberID uint64
		OrgName  string
	}
	err := db.DB.Model(&models.Application{}).
		Select("member_applications.member_id AS member_id, mo.name AS org_name").
		Joins("JOIN member_organizations mo ON mo.id = member_applications.org_id").
		Where("member_applications.member_id IN ? AND member_applications.status IN ?", ids,
			[]string{models.AppStatusApproved, models.AppStatusPendingReview}).
		Order("member_applications.created_at DESC, member_applications.id DESC").
		Scan(&appRows).Error
	if err != nil {
		return err
	}
	for _, r := range appRows {
		if m, ok := index[r.MemberID]; ok && m.OrgName == "" {
			m.OrgName = r.OrgName
		}
	}

	// 2) 兜底：取该会员最近一条“已缴费”记录所属机构
	var feeRows []struct {
		MemberID uint64
		OrgName  string
	}
	err = db.DB.Model(&models.FeeRecord{}).
		Select("member_id, org_name").
		Where("member_id IN ? AND status = ? AND org_name <> ''", ids, models.FeeStatusPaid).
		Order("year DESC, id DESC").
		Scan(&feeRows).Error
	if err != nil {
		return err
	}
	for _, r := range feeRows {
		if m, ok := index[r.MemberID]; ok && m.OrgName == "" {
			m.OrgName = r.OrgName
		}
	}
	return nil
}

// ListMembers returns paginated member list (admin)
func (s *MemberService) ListMembers(page, size int, keyword, status, memberType string) ([]models.Member, int64, error) {
	var members []models.Member
	var total int64

	query := db.DB.Model(&models.Member{})
	// 不显示管理员
	query = query.Where("is_admin = ?", false)
	if keyword != "" {
		kw := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR company_name LIKE ? OR name LIKE ? OR mobile LIKE ? OR email LIKE ?",
			kw, kw, kw, kw, kw)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if memberType != "" {
		query = query.Where("member_type = ?", memberType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&members).Error; err != nil {
		return nil, 0, err
	}
	// 填充主入会机构名称
	if err := s.fillOrgNames(members); err != nil {
		return nil, 0, err
	}
	return members, total, nil
}

// UpdateMemberStatus updates a member's status (admin)
func (s *MemberService) UpdateMemberStatus(id uint64, status string) error {
	validStatuses := map[string]bool{
		models.MemberStatusRegistering:   true,
		models.MemberStatusPendingReview: true,
		models.MemberStatusPendingPay:    true,
		models.MemberStatusActive:        true,
		models.MemberStatusRejected:      true,
		models.MemberStatusExpired:       true,
	}
	if !validStatuses[status] {
		return errors.New("无效的状态值")
	}
	return db.DB.Model(&models.Member{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateMemberLevel updates a member's level (admin). Only active members may change
// level, and the target level must come from the member's paid memberships (会籍).
// Every successful change is recorded in the membership change log.
func (s *MemberService) UpdateMemberLevel(id uint64, levelID uint64, operator string, reason string) error {
	var m models.Member
	if err := db.DB.First(&m, id).Error; err != nil {
		return errors.New("会员不存在")
	}
	if m.Status != models.MemberStatusActive {
		return errors.New("仅正式会员可变更等级")
	}
	if levelID == 0 {
		return errors.New("请选择会员等级")
	}

	// 查找目标等级并校验
	var lvl models.MemberLevel
	if err := db.DB.First(&lvl, levelID).Error; err != nil {
		return errors.New("会员等级不存在")
	}
	oldLevelID := parseUint(m.MemberLevel)
	if oldLevelID == lvl.ID {
		return errors.New("新旧会员等级不能相同")
	}

	// 目标等级必须来自该会员已缴费加入的机构所支持的等级（member_org_levels）
	var cnt int64
	if err := db.DB.Model(&models.MemberLevel{}).
		Joins("JOIN member_org_levels mol ON mol.level_id = member_levels.id").
		Joins("JOIN member_fee_records fr ON fr.org_id = mol.org_id").
		Where("fr.member_id = ? AND fr.status = ? AND member_levels.id = ?", id, models.FeeStatusPaid, levelID).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return errors.New("该等级不在该会员已缴费加入的机构所支持的等级中")
	}

	// 原始会籍名称
	oldLevelName := ""
	if oldLevelID > 0 {
		var oldLvl models.MemberLevel
		if err := db.DB.First(&oldLvl, oldLevelID).Error; err == nil {
			oldLevelName = oldLvl.Name
		}
	}

	if err := db.DB.Model(&models.Member{}).Where("id = ?", id).Update("member_level", lvl.ID).Error; err != nil {
		return err
	}

	// 同步更新该会员生效证书的等级
	db.DB.Model(&models.Certificate{}).
		Where("member_id = ? AND status = ?", id, models.CertStatusActive).
		Updates(map[string]interface{}{"level_id": lvl.ID, "level_name": lvl.Name})

	// 主入会机构
	orgID, orgName := s.primaryOrg(&m)

	// 记录会籍变更
	change := models.MemberLevelChange{
		MemberID:     m.ID,
		Username:     m.Username,
		MemberName:   memberDisplayName(&m),
		MemberType:   m.MemberType,
		ChangeYear:   time.Now().Year(),
		OrgID:        orgID,
		OrgName:      orgName,
		OldLevelID:   oldLevelID,
		OldLevelName: oldLevelName,
		NewLevelID:   lvl.ID,
		NewLevelName: lvl.Name,
		Reason:       reason,
		Operator:     operator,
	}
	if err := db.DB.Create(&change).Error; err != nil {
		return err
	}

	return nil
}

// primaryOrg returns the member's primary joined organization (id + name).
func (s *MemberService) primaryOrg(m *models.Member) (uint64, string) {
	members := []models.Member{*m}
	_ = s.fillOrgNames(members)
	name := members[0].OrgName
	if name == "" {
		return 0, ""
	}
	var org models.Organization
	if err := db.DB.Where("name = ?", name).First(&org).Error; err != nil {
		return 0, name
	}
	return org.ID, name
}

// ListLevelChanges returns paginated membership change records (admin).
func (s *MemberService) ListLevelChanges(page, size int, keyword, memberType string) ([]models.MemberLevelChange, int64, error) {
	var list []models.MemberLevelChange
	var total int64

	query := db.DB.Model(&models.MemberLevelChange{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR member_name LIKE ? OR old_level_name LIKE ? OR new_level_name LIKE ?",
			kw, kw, kw, kw)
	}
	if memberType != "" {
		query = query.Where("member_type = ?", memberType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListProfileChanges returns paginated profile change records (资料变更记录, admin).
func (s *MemberService) ListProfileChanges(page, size int, keyword string) ([]models.ProfileChange, int64, error) {
	var list []models.ProfileChange
	var total int64

	query := db.DB.Model(&models.ProfileChange{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		query = query.Where("operator LIKE ? OR name LIKE ?", kw, kw)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetMemberLevelChanges returns a single member's level change history, newest first.
func (s *MemberService) GetMemberLevelChanges(memberID uint64, page, size int) ([]models.MemberLevelChange, int64, error) {
	var list []models.MemberLevelChange
	var total int64

	query := db.DB.Model(&models.MemberLevelChange{}).Where("member_id = ?", memberID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC, id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// memberDisplayName returns the company name for unit members or personal name.
func memberDisplayName(m *models.Member) string {
	if m.MemberType == models.MemberTypePersonal {
		if m.Name != "" {
			return m.Name
		}
		return m.Username
	}
	if m.CompanyName != "" {
		return m.CompanyName
	}
	return m.Username
}

// parseUint parses a string into uint64, returning 0 on error or empty input.
func parseUint(s string) uint64 {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// GetMemberAvailableLevels returns the member levels supported by the organizations
// the member has paid to join (已缴费加入机构的关联等级), used for the change-level dropdown.
func (s *MemberService) GetMemberAvailableLevels(memberID uint64) ([]models.MemberLevel, error) {
	var levels []models.MemberLevel
	err := db.DB.Raw(`
		SELECT DISTINCT ml.id, ml.name, ml.level, ml.description, ml.created_at, ml.updated_at
		FROM member_levels ml
		JOIN member_org_levels mol ON mol.level_id = ml.id
		JOIN member_fee_records fr ON fr.org_id = mol.org_id
		WHERE fr.member_id = ? AND fr.status = ?
		ORDER BY ml.level ASC
	`, memberID, models.FeeStatusPaid).Scan(&levels).Error
	if err != nil {
		return nil, err
	}
	return levels, nil
}

// MemberJoinedOrgInfo 会员加入组织机构信息（管理端查看）
type MemberJoinedOrgInfo struct {
	Paid      []MemberPaidJoin      `json:"paid"`      // 付费加入
	Voluntary []MemberVoluntaryJoin `json:"voluntary"` // 主动加入
}

// MemberPaidJoin 付费加入：已缴费会费记录对应的机构
type MemberPaidJoin struct {
	OrgID      uint64  `json:"org_id"`
	OrgName    string  `json:"org_name"`
	LevelID    uint64  `json:"level_id"`
	LevelName  string  `json:"level_name"`
	Year       int     `json:"year"`
	Amount     float64 `json:"amount"`      // 应缴金额
	PaidAmount float64 `json:"paid_amount"` // 实缴金额
	PaidDate   string  `json:"paid_date"`
	PaidAt     string  `json:"paid_at"`
}

// MemberVoluntaryJoin 主动加入：会员自主加入组织记录
type MemberVoluntaryJoin struct {
	ID        uint64 `json:"id"`
	OrgID     uint64 `json:"org_id"`
	OrgName   string `json:"org_name"`
	LevelID   uint64 `json:"level_id"`
	LevelName string `json:"level_name"`
	JoinedAt  string `json:"joined_at"`
}

// GetMemberJoinedOrgs returns the organizations a member has joined (admin):
//   - paid: 已缴费（status=paid）的会费记录所属机构；
//   - voluntary: 会员自主加入的组织（member_user_orgs）。
func (s *MemberService) GetMemberJoinedOrgs(memberID uint64) (*MemberJoinedOrgInfo, error) {
	info := &MemberJoinedOrgInfo{
		Paid:      make([]MemberPaidJoin, 0),
		Voluntary: make([]MemberVoluntaryJoin, 0),
	}

	// 1) 付费加入：已缴费会费记录
	var fees []models.FeeRecord
	if err := db.DB.Where("member_id = ? AND status = ? AND org_name <> ''", memberID, models.FeeStatusPaid).
		Order("year DESC, id DESC").Find(&fees).Error; err != nil {
		return nil, err
	}
	for _, f := range fees {
		paidAt := ""
		if f.PaidAt != nil && !f.PaidAt.IsZero() {
			paidAt = f.PaidAt.Format("2006-01-02 15:04:05")
		}
		info.Paid = append(info.Paid, MemberPaidJoin{
			OrgID:      f.OrgID,
			OrgName:    f.OrgName,
			LevelID:    f.LevelID,
			LevelName:  f.LevelName,
			Year:       f.Year,
			Amount:     f.Amount,
			PaidAmount: f.PaidAmount,
			PaidDate:   f.PaidDate,
			PaidAt:     paidAt,
		})
	}

	// 2) 主动加入：会员自主加入组织记录
	var orgs []models.MemberOrganization
	if err := db.DB.Preload("Org").Where("member_id = ?", memberID).
		Order("joined_at DESC, id DESC").Find(&orgs).Error; err != nil {
		return nil, err
	}

	levelIDs := make([]uint64, 0)
	for _, o := range orgs {
		if o.LevelID > 0 {
			levelIDs = append(levelIDs, o.LevelID)
		}
	}
	levelNames := map[uint64]string{}
	if len(levelIDs) > 0 {
		var lvls []models.MemberLevel
		if err := db.DB.Where("id IN ?", levelIDs).Find(&lvls).Error; err == nil {
			for _, l := range lvls {
				levelNames[l.ID] = l.Name
			}
		}
	}
	for _, o := range orgs {
		orgName := ""
		if o.Org.ID > 0 {
			orgName = o.Org.Name
		}
		info.Voluntary = append(info.Voluntary, MemberVoluntaryJoin{
			ID:        o.ID,
			OrgID:     o.OrgID,
			OrgName:   orgName,
			LevelID:   o.LevelID,
			LevelName: levelNames[o.LevelID],
			JoinedAt:  o.JoinedAt.Format("2006-01-02 15:04"),
		})
	}

	return info, nil
}

// DeleteMember 物理删除会员（admin）
// 仅“注册中”状态的会员可删除，防止误删正式会员。
func (s *MemberService) DeleteMember(id uint64) error {
	var m models.Member
	if err := db.DB.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("会员不存在")
		}
		return err
	}
	if m.Status != models.MemberStatusRegistering {
		return errors.New("仅“注册中”状态的会员可删除")
	}
	return db.DB.Delete(&m).Error
}

// DefaultPassword is the default password set when an admin resets a member's password.
const DefaultPassword = "Abcd@1234"

// ResetMemberPassword resets a member's password to the default (admin).
func (s *MemberService) ResetMemberPassword(id uint64) error {
	var m models.Member
	if err := db.DB.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("会员不存在")
		}
		return err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(DefaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	return db.DB.Model(&models.Member{}).Where("id = ?", id).Update("password", string(hashed)).Error
}

// CreateMemberRequest is the admin request for creating a member directly.
type CreateMemberRequest struct {
	Username      string   `json:"username" binding:"required"`
	Password      string   `json:"password"`
	Mobile        string   `json:"mobile"`
	Email         string   `json:"email"`
	MemberType    string   `json:"member_type"`
	Status        string   `json:"status"`
	CompanyName   string   `json:"company_name"`
	CreditCode    string   `json:"credit_code"`
	LegalPerson   string   `json:"legal_person"`
	ContactPerson string   `json:"contact_person"`
	ContactTitle  string   `json:"contact_title"`
	ContactMobile string   `json:"contact_mobile"`
	Industry      string   `json:"industry"`
	Address       string   `json:"address"`
	Name          string   `json:"name"`
	IDCard        string   `json:"id_card"`
	LevelID       uint64   `json:"level_id"`
	RootOrgID     uint64   `json:"root_org_id"`
	OrgIDs        []uint64 `json:"org_ids"`
}

// CreateMember creates a new member directly (admin).
func (s *MemberService) CreateMember(req CreateMemberRequest, operator string) (*models.Member, error) {
	if req.MemberType == "" {
		req.MemberType = models.MemberTypeUnit
	}
	if req.MemberType != models.MemberTypeUnit && req.MemberType != models.MemberTypePersonal {
		return nil, errors.New("无效的会员类型")
	}

	// 规范化并校验基础信息
	req.Username = strings.TrimSpace(req.Username)
	req.Mobile = strings.TrimSpace(req.Mobile)
	req.Email = strings.TrimSpace(req.Email)
	req.ContactMobile = strings.TrimSpace(req.ContactMobile)
	req.CreditCode = strings.ToUpper(strings.TrimSpace(req.CreditCode))
	req.IDCard = strings.ToUpper(strings.TrimSpace(req.IDCard))
	req.Name = strings.TrimSpace(req.Name)
	req.CompanyName = strings.TrimSpace(req.CompanyName)

	if req.Username == "" {
		return nil, errors.New("请输入用户名")
	}
	if !utils.IsValidMobile(req.Mobile) {
		return nil, errors.New("请输入合法手机号")
	}
	if !utils.IsValidEmail(req.Email) {
		return nil, errors.New("请输入合法邮箱")
	}
	if req.MemberType == models.MemberTypeUnit {
		if !utils.IsValidPhone(req.ContactMobile) {
			return nil, errors.New("请输入合法联系电话")
		}
		if !utils.IsValidCreditCode(req.CreditCode) {
			return nil, errors.New("请输入合法的统一社会信用代码")
		}
	} else {
		req.CreditCode = ""
		req.ContactMobile = ""
		if req.Name == "" {
			return nil, errors.New("请输入姓名")
		}
		if !utils.IsValidIDCard(req.IDCard) {
			return nil, errors.New("请输入合法身份证号")
		}
	}

	if req.Status == "" {
		req.Status = models.MemberStatusActive
	}
	validStatuses := map[string]bool{
		models.MemberStatusRegistering:   true,
		models.MemberStatusPendingReview: true,
		models.MemberStatusPendingPay:    true,
		models.MemberStatusActive:        true,
		models.MemberStatusRejected:      true,
		models.MemberStatusExpired:       true,
	}
	if !validStatuses[req.Status] {
		return nil, errors.New("无效的会员状态")
	}

	// 唯一性校验
	var exist models.Member
	if err := db.DB.Where("username = ?", req.Username).First(&exist).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}
	if req.Mobile != "" {
		var m models.Member
		if err := db.DB.Where("mobile = ?", req.Mobile).First(&m).Error; err == nil {
			return nil, errors.New("手机号已被注册")
		}
	}
	if req.Email != "" {
		var e models.Member
		if err := db.DB.Where("email = ?", req.Email).First(&e).Error; err == nil {
			return nil, errors.New("邮箱已被注册")
		}
	}
	if req.ContactMobile != "" {
		var cm models.Member
		if err := db.DB.Where("contact_mobile = ?", req.ContactMobile).First(&cm).Error; err == nil {
			return nil, errors.New("该联系电话已被使用")
		}
	}
	if req.CompanyName != "" {
		var cn models.Member
		if err := db.DB.Where("company_name = ?", req.CompanyName).First(&cn).Error; err == nil {
			return nil, errors.New("该公司名称已存在")
		}
	}
	if req.CreditCode != "" {
		var cc models.Member
		if err := db.DB.Where("credit_code = ?", req.CreditCode).First(&cc).Error; err == nil {
			return nil, errors.New("该统一社会信用代码已存在")
		}
	}

	password := req.Password
	if password == "" {
		password = DefaultPassword
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	// 加入总会必须选择默认会员等级；分会自动沿用总会等级
	memberLevel := ""
	levelName := ""
	if req.LevelID > 0 {
		var lvl models.MemberLevel
		if err := db.DB.First(&lvl, req.LevelID).Error; err != nil {
			return nil, errors.New("会员等级不存在")
		}
		// 校验所选等级属于所选总会
		if req.RootOrgID > 0 {
			var cnt int64
			db.DB.Model(&models.MemberOrgLevel{}).Where("org_id = ? AND level_id = ?", req.RootOrgID, req.LevelID).Count(&cnt)
			if cnt == 0 {
				return nil, errors.New("所选会员等级不属于所选总会")
			}
		}
		memberLevel = strconv.FormatUint(lvl.ID, 10)
		levelName = lvl.Name
	}

	member := models.Member{
		Username:      req.Username,
		Password:      string(hashed),
		Mobile:        req.Mobile,
		Email:         req.Email,
		MemberType:    req.MemberType,
		Status:        req.Status,
		MemberLevel:   memberLevel,
		CompanyName:   req.CompanyName,
		CreditCode:    req.CreditCode,
		LegalPerson:   req.LegalPerson,
		ContactPerson: req.ContactPerson,
		ContactTitle:  req.ContactTitle,
		ContactMobile: req.ContactMobile,
		Industry:      req.Industry,
		Address:       req.Address,
		Name:          req.Name,
		IDCard:        req.IDCard,
	}
	if err := db.DB.Create(&member).Error; err != nil {
		return nil, err
	}

	// 总会通过缴费表记录（免缴、已缴费），不再写入 member_user_orgs；
	// 分会/代表机构直接写入 member_user_orgs。
	joined := make(map[uint64]bool)
	if req.RootOrgID == 0 {
		// 向后兼容：未指定总会时自动加入所有根组织
		var roots []models.Organization
		if err := db.DB.Where("parent_id = ?", 0).Find(&roots).Error; err == nil {
			for _, root := range roots {
				if root.ID == 0 || joined[root.ID] {
					continue
				}
				db.DB.Create(&models.MemberOrganization{MemberID: member.ID, OrgID: root.ID, LevelID: req.LevelID})
				joined[root.ID] = true
			}
		}
	}
	for _, orgID := range req.OrgIDs {
		if orgID == 0 || joined[orgID] {
			continue
		}
		var org models.Organization
		if err := db.DB.First(&org, orgID).Error; err != nil {
			continue
		}
		db.DB.Create(&models.MemberOrganization{MemberID: member.ID, OrgID: orgID, LevelID: req.LevelID})
		joined[orgID] = true
	}

	// 将加入总会的信息写入缴费表（免缴、已缴费）
	rootOrgName := ""
	if req.RootOrgID > 0 && req.LevelID > 0 {
		var rootOrg models.Organization
		if err := db.DB.Where("id = ? AND parent_id = ?", req.RootOrgID, 0).First(&rootOrg).Error; err == nil {
			rootOrgName = rootOrg.Name
			now := time.Now()
			// 会费金额 = 该总会所选会员等级对应年度的会费标准
			var amount float64
			var fs models.MemberFeeStandard
			if err := db.DB.Where("level_id = ? AND year = ?", req.LevelID, now.Year()).First(&fs).Error; err == nil {
				amount = fs.Amount
			}
			fee := models.FeeRecord{
				MemberID:  member.ID,
				Year:      now.Year(),
				Amount:    amount,
				Status:    models.FeeStatusPaid,
				PaidAt:    &models.LocalTime{Time: now},
				Remark:    "免缴",
				OrgID:     rootOrg.ID,
				OrgName:   rootOrg.Name,
				LevelID:   req.LevelID,
				LevelName: levelName,
			}
			db.DB.Create(&fee)
		}
	}

	// 生成会员证书（带等级与证书模板）
	if req.LevelID > 0 {
		now := time.Now()
		var tpl models.MemberCertificateTemplate
		if err := db.DB.Where("level_id = ?", req.LevelID).First(&tpl).Error; err != nil {
			// 未配置该等级证书样式时，回退到最低等级样式
			db.DB.Joins("JOIN member_levels ml ON ml.id = member_certificate_templates.level_id").
				Order("ml.level ASC").
				First(&tpl)
		}
		cert := models.Certificate{
			MemberID:       member.ID,
			CertNo:         "XXXXXX-" + now.Format("2006") + "-" + padLeftGen(member.ID),
			IssuedAt:       &models.LocalTime{Time: now},
			ExpireAt:       &models.LocalTime{Time: time.Date(now.Year(), 12, 31, 23, 59, 59, 0, now.Location())},
			Status:         models.CertStatusActive,
			LevelID:        req.LevelID,
			LevelName:      levelName,
			CertTemplateID: tpl.ID,
		}
		db.DB.Create(&cert)
	}

	// 插入会籍变更记录（按所选总会新增入会）
	if req.LevelID > 0 {
		change := models.MemberLevelChange{
			MemberID:     member.ID,
			Username:     member.Username,
			MemberName:   memberDisplayName(&member),
			MemberType:   member.MemberType,
			ChangeYear:   time.Now().Year(),
			OrgID:        req.RootOrgID,
			OrgName:      rootOrgName,
			NewLevelID:   req.LevelID,
			NewLevelName: levelName,
			Reason:       "新增会员",
			Operator:     operator,
		}
		db.DB.Create(&change)
	}

	return &member, nil
}

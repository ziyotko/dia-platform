package service

import (
	"errors"

	"base/internal/models"
	"base/pkg/db"

	"gorm.io/gorm"
)

// WorkflowService 流程定义（含节点编排）服务。
// 租户口径与 user/role/workflow_role 保持一致：读写严格按租户隔离，平台超管可跨租户管理。
type WorkflowService struct{}

type WorkflowListQuery struct {
	TenantID       uint64
	FilterTenantID uint64
	Keyword        string
	Status         *int
	Page           int
	Size           int
}

// WorkflowUserOption 审批人选择器用的用户精简信息
type WorkflowUserOption struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	RealName string `json:"realName"`
	Status   int    `json:"status"`
}

// WorkflowApproverOptions 节点审批人可选值（流程角色 + 用户）
type WorkflowApproverOptions struct {
	Roles []models.WorkflowRole `json:"roles"`
	Users []WorkflowUserOption  `json:"users"`
}

func (s WorkflowService) Create(w *models.Workflow) error {
	var count int64
	if err := db.DB.Model(&models.Workflow{}).
		Where("tenant_id = ? AND code = ?", w.TenantID, w.Code).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该租户下流程编码已存在")
	}
	return db.DB.Create(w).Error
}

func (s WorkflowService) Update(w *models.Workflow, tenantID uint64) error {
	query := db.DB.Model(&models.Workflow{}).Where("id = ?", w.ID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("流程定义不存在")
	}

	var dup int64
	if err := db.DB.Model(&models.Workflow{}).
		Where("tenant_id = ? AND code = ? AND id <> ?", w.TenantID, w.Code, w.ID).
		Count(&dup).Error; err != nil {
		return err
	}
	if dup > 0 {
		return errors.New("该租户下流程编码已存在")
	}

	return db.DB.Model(&models.Workflow{}).Where("id = ?", w.ID).Updates(map[string]interface{}{
		"name":        w.Name,
		"code":        w.Code,
		"description": w.Description,
		"status":      w.Status,
	}).Error
}

// Delete 删除流程定义：存在审批中的实例时拒绝删除；同时清理节点。
func (s WorkflowService) Delete(id uint64, tenantID uint64) error {
	var wf models.Workflow
	query := db.DB.Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&wf).Error; err != nil {
		return err
	}

	var running int64
	if err := db.DB.Model(&models.WorkflowInstance{}).
		Where("workflow_id = ? AND status = ?", id, models.WorkflowInstanceRunning).
		Count(&running).Error; err != nil {
		return err
	}
	if running > 0 {
		return errors.New("该流程还有审批中的实例，请先处理完成后再删除")
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workflow_id = ?", id).Delete(&models.WorkflowNode{}).Error; err != nil {
			return err
		}
		// 必须带主键/条件删除：GORM 会拦截无条件的全局删除
		return tx.Where("id = ?", wf.ID).Delete(&models.Workflow{}).Error
	})
}

func (s WorkflowService) GetByID(id uint64, tenantID uint64) (*models.Workflow, error) {
	var wf models.Workflow
	query := db.DB.Preload("Nodes", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sort ASC, id ASC")
	}).Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&wf).Error; err != nil {
		return &wf, err
	}
	fillApproverNames(wf.Nodes)
	return &wf, nil
}

// List 分页查询流程定义。tenantID 为当前登录用户所属租户：
//   - 普通租户用户（tenantID > 0）只能看到本租户；
//   - 平台超管（tenantID == 0）不传 filterTenantID 时查看全部租户。
func (s WorkflowService) List(q WorkflowListQuery) ([]models.Workflow, int64, error) {
	var list []models.Workflow
	var total int64
	query := db.DB.Model(&models.Workflow{})
	if q.TenantID > 0 {
		query = query.Where("tenant_id = ?", q.TenantID)
	} else if q.FilterTenantID > 0 {
		query = query.Where("tenant_id = ?", q.FilterTenantID)
	}
	if q.Keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+q.Keyword+"%", "%"+q.Keyword+"%")
	}
	if q.Status != nil {
		query = query.Where("status = ?", *q.Status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Preload("Nodes", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sort ASC, id ASC")
	}).Order("created_at DESC").Offset((q.Page - 1) * q.Size).Limit(q.Size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	for i := range list {
		fillApproverNames(list[i].Nodes)
	}
	return list, total, nil
}

// Options 启用中的流程定义选项（供「发起流程」选择，不分页、仅返回必要字段）。
func (s WorkflowService) Options(tenantID uint64) ([]models.Workflow, error) {
	var list []models.Workflow
	query := db.DB.Preload("Nodes", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sort ASC, id ASC")
	}).Where("status = ?", 1)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.Order("name ASC").Find(&list).Error
	return list, err
}

// ApproverOptions 节点审批人可选值：本租户的流程角色 + 用户。
func (s WorkflowService) ApproverOptions(tenantID uint64) (*WorkflowApproverOptions, error) {
	opts := &WorkflowApproverOptions{Roles: []models.WorkflowRole{}, Users: []WorkflowUserOption{}}

	roleQuery := db.DB.Model(&models.WorkflowRole{}).Order("name ASC")
	userQuery := db.DB.Model(&models.User{}).Order("username ASC")
	if tenantID > 0 {
		roleQuery = roleQuery.Where("tenant_id = ?", tenantID)
		userQuery = userQuery.Where("tenant_id = ?", tenantID)
	}
	if err := roleQuery.Find(&opts.Roles).Error; err != nil {
		return nil, err
	}
	var users []models.User
	if err := userQuery.Find(&users).Error; err != nil {
		return nil, err
	}
	for _, u := range users {
		opts.Users = append(opts.Users, WorkflowUserOption{
			ID:       u.ID,
			Username: u.Username,
			RealName: u.RealName,
			Status:   u.Status,
		})
	}
	return opts, nil
}

// SaveNodes 覆盖式保存流程节点（按传入顺序重新编号 Sort，保证顺序连续）。
// 存在审批中的实例时拒绝修改：节点行会被重建，运行中实例的 node_id 将失效。
func (s WorkflowService) SaveNodes(workflowID uint64, nodes []models.WorkflowNode, tenantID uint64) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var wf models.Workflow
		query := tx.Where("id = ?", workflowID)
		if tenantID > 0 {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.First(&wf).Error; err != nil {
			return errors.New("流程定义不存在")
		}

		var running int64
		if err := tx.Model(&models.WorkflowInstance{}).
			Where("workflow_id = ? AND status = ?", workflowID, models.WorkflowInstanceRunning).
			Count(&running).Error; err != nil {
			return err
		}
		if running > 0 {
			return errors.New("该流程还有审批中的实例，节点已被运行中的实例引用，请处理完成后再修改")
		}

		if err := tx.Where("workflow_id = ?", workflowID).Delete(&models.WorkflowNode{}).Error; err != nil {
			return err
		}
		if len(nodes) == 0 {
			return nil
		}
		cleaned := make([]models.WorkflowNode, 0, len(nodes))
		for i, node := range nodes {
			node.ID = 0
			node.WorkflowID = workflowID
			node.Sort = i + 1
			if node.Name == "" {
				return errors.New("节点名称不能为空")
			}
			switch node.ApproverType {
			case models.ApproverTypeRole:
				if node.ApproverID == 0 {
					return errors.New("请为节点选择流程角色")
				}
			case models.ApproverTypeUser:
				if node.ApproverID == 0 {
					return errors.New("请为节点选择审批人")
				}
			case models.ApproverTypeInitiator:
				node.ApproverID = 0
			default:
				return errors.New("节点审批人类型不合法")
			}
			cleaned = append(cleaned, node)
		}
		return tx.Create(&cleaned).Error
	})
}

// fillApproverNames 填充节点的审批人展示名，避免前端再查一次字典。
func fillApproverNames(nodes []models.WorkflowNode) {
	if len(nodes) == 0 {
		return
	}
	roleIDs := make([]uint64, 0, len(nodes))
	userIDs := make([]uint64, 0, len(nodes))
	for _, n := range nodes {
		switch n.ApproverType {
		case models.ApproverTypeRole:
			roleIDs = append(roleIDs, n.ApproverID)
		case models.ApproverTypeUser:
			userIDs = append(userIDs, n.ApproverID)
		}
	}

	roleNames := map[uint64]string{}
	if len(roleIDs) > 0 {
		var roles []models.WorkflowRole
		if err := db.DB.Where("id IN ?", roleIDs).Find(&roles).Error; err == nil {
			for _, r := range roles {
				roleNames[r.ID] = r.Name
			}
		}
	}
	userNames := map[uint64]string{}
	if len(userIDs) > 0 {
		var users []models.User
		if err := db.DB.Where("id IN ?", userIDs).Find(&users).Error; err == nil {
			for _, u := range users {
				if u.RealName != "" {
					userNames[u.ID] = u.RealName
				} else {
					userNames[u.ID] = u.Username
				}
			}
		}
	}

	for i := range nodes {
		switch nodes[i].ApproverType {
		case models.ApproverTypeRole:
			nodes[i].ApproverName = roleNames[nodes[i].ApproverID]
		case models.ApproverTypeUser:
			nodes[i].ApproverName = userNames[nodes[i].ApproverID]
		case models.ApproverTypeInitiator:
			nodes[i].ApproverName = "发起人本人"
		}
	}
}

package seed

import (
	"base/internal/models"
	"base/pkg/db"

	"github.com/sirupsen/logrus"
)

// MockOrganizations 插入模拟机构数据，仅在本地测试使用
func MockOrganizations() error {
	var count int64
	if err := db.DB.Model(&models.Organization{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		logrus.Infof("已有 %d 条机构数据，跳过模拟数据插入", count)
		return nil
	}

	orgs := []models.Organization{
		{TenantID: 0, ParentID: 0, Code: "headquarter", Name: "集团总部", Leader: "张总", Phone: "13800000001", Email: "hq@example.com", Sort: 1, Status: 1},
		{TenantID: 0, ParentID: 1, Code: "rd-center", Name: "研发中心", Leader: "李工", Phone: "13800000002", Email: "rd@example.com", Sort: 1, Status: 1},
		{TenantID: 0, ParentID: 1, Code: "product-center", Name: "产品中心", Leader: "王经理", Phone: "13800000003", Email: "product@example.com", Sort: 2, Status: 1},
		{TenantID: 0, ParentID: 1, Code: "market-center", Name: "市场中心", Leader: "赵经理", Phone: "13800000004", Email: "market@example.com", Sort: 3, Status: 1},
		{TenantID: 0, ParentID: 2, Code: "backend-team", Name: "后端研发部", Leader: "刘组长", Phone: "13800000005", Email: "backend@example.com", Sort: 1, Status: 1},
		{TenantID: 0, ParentID: 2, Code: "frontend-team", Name: "前端研发部", Leader: "陈组长", Phone: "13800000006", Email: "frontend@example.com", Sort: 2, Status: 1},
		{TenantID: 0, ParentID: 2, Code: "qa-team", Name: "测试部", Leader: "杨组长", Phone: "13800000007", Email: "qa@example.com", Sort: 3, Status: 1},
		{TenantID: 0, ParentID: 3, Code: "pm-team", Name: "产品一部", Leader: "周经理", Phone: "13800000008", Email: "pm1@example.com", Sort: 1, Status: 1},
		{TenantID: 0, ParentID: 4, Code: "sales-team", Name: "销售一部", Leader: "吴经理", Phone: "13800000009", Email: "sales1@example.com", Sort: 1, Status: 1},
		{TenantID: 0, ParentID: 4, Code: "sales-team2", Name: "销售二部", Leader: "郑经理", Phone: "13800000010", Email: "sales2@example.com", Sort: 2, Status: 1},
	}

	if err := db.DB.CreateInBatches(orgs, 100).Error; err != nil {
		return err
	}
	logrus.Infof("已插入 %d 条模拟机构数据", len(orgs))
	return nil
}

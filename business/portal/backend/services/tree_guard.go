package services

import (
	"errors"

	"server/utils"
)

// validateTreeParent 校验将节点 nodeID 的上级节点设置为 parentID 是否合法：
//  1. parentID == 0 表示提升为根节点，始终合法；
//  2. 上级节点不能是自身；
//  3. 上级节点必须存在（否则节点会成为无法从根到达的“孤儿”，从树中静默消失）；
//  4. 上级节点不能是 nodeID 的下级，否则父子关系成环。
//
// table 为物理表名（menu / department / organization），nodeID 为被更新节点的主键。
func validateTreeParent(table string, nodeID, parentID uint) error {
	if parentID == 0 {
		return nil
	}
	if parentID == nodeID {
		return errors.New("上级节点不能是自身")
	}

	exists, err := treeRecordExists(table, parentID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("上级节点不存在")
	}

	// 自 parentID 向上追溯，若途经 nodeID 说明把节点挂到了自己的后代下
	seen := map[uint]bool{parentID: true}
	cur := parentID
	for cur != 0 {
		var row struct {
			ParentID uint
		}
		if err := utils.DB.Table(table).Select("parent_id").Where("id = ?", cur).Scan(&row).Error; err != nil {
			return err
		}
		if row.ParentID == 0 {
			break
		}
		if row.ParentID == nodeID {
			return errors.New("上级节点不能是自身或其下级节点")
		}
		if seen[row.ParentID] {
			break // 数据中已有脏环，避免死循环
		}
		seen[row.ParentID] = true
		cur = row.ParentID
	}
	return nil
}

// treeRecordExists 判断指定表中是否存在某主键记录。
func treeRecordExists(table string, id uint) (bool, error) {
	var count int64
	if err := utils.DB.Table(table).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

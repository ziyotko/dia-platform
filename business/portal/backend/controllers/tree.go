package controllers

import "github.com/gin-gonic/gin"

// buildFlatTree 将扁平的父子节点列表组装为层级树（gin.H，含 children / hasChildren 字段）。
//   - idOf / parentOf 分别提取节点主键与父节点主键；
//   - newNode 负责构造单个节点（children / hasChildren 由本函数统一填充）；
//   - 父节点不在结果集中的“孤儿”节点会被提升为根节点，避免从树中静默消失。
//
// 供部门树（buildDeptTree）、机构树（buildOrgTree）等复用，避免多套重复的父子挂载逻辑。
func buildFlatTree[T any](
	list []T,
	idOf func(T) uint,
	parentOf func(T) uint,
	newNode func(T) gin.H,
) []gin.H {
	nodeMap := make(map[uint]*gin.H, len(list))
	var roots []gin.H

	for i := range list {
		item := list[i]
		node := newNode(item)
		nodeMap[idOf(item)] = &node
	}

	for i := range list {
		item := list[i]
		node := nodeMap[idOf(item)]
		pid := parentOf(item)
		if pid == 0 {
			roots = append(roots, *node)
			continue
		}
		parent, ok := nodeMap[pid]
		if !ok {
			// 上级节点缺失：作为根节点展示，保证节点不会丢失
			roots = append(roots, *node)
			continue
		}
		children := (*parent)["children"].([]gin.H)
		children = append(children, *node)
		(*parent)["children"] = children
		(*parent)["hasChildren"] = true
	}

	return roots
}

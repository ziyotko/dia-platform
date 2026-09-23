package db

import (
	"member/pkg/utils"
)

// EnsureUniqueMemberFeeIndex 幂等保证 member_fee_records 上存在 (member_id, year) 唯一索引 uk_member_year。
//
// 背景：模型里**故意**没有用 `uniqueIndex` 声明该索引——AutoMigrate 在既有表上补建唯一索引时，
// 一旦表里已有历史重复数据就会直接报错（启动 Fatal 退出）。因此改在启动时显式、幂等地执行：
//  1. 表不存在（全新库）→ 跳过（AutoMigrate 随后会建表）；
//  2. 索引已存在 → 直接返回；
//  3. 已有重复行 → 每组保留 id 最小的一条、删除多余行（并打印条数，避免“静默删数据”）；
//  4. 创建唯一索引。
//
// 任一步失败只记日志、不阻断启动：缺少索引时服务层仍有事务 + 行锁查重兜底。
func EnsureUniqueMemberFeeIndex() {
	if DB == nil {
		return
	}
	const (
		table     = "member_fee_records"
		indexName = "uk_member_year"
	)

	var tableCount int64
	if err := DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", table).
		Scan(&tableCount).Error; err != nil {
		utils.LogWarn("检查表 %s 是否存在失败（跳过唯一索引检查）：%v", table, err)
		return
	}
	if tableCount == 0 {
		return
	}

	var indexCount int64
	if err := DB.Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?", table, indexName).
		Scan(&indexCount).Error; err != nil {
		utils.LogWarn("检查表 %s 的索引 %s 失败：%v", table, indexName, err)
		return
	}
	if indexCount > 0 {
		return
	}

	// 历史重复行清理：按 (member_id, year) 分组，保留 id 最小的一条
	var dupGroups, extraRows int64
	dupQuery := "SELECT COUNT(*) AS dup_groups, IFNULL(SUM(c-1), 0) AS extra_rows FROM (" +
		"SELECT member_id, year, COUNT(*) c FROM " + table + " GROUP BY member_id, year HAVING c > 1) t"
	if err := DB.Raw(dupQuery).Row().Scan(&dupGroups, &extraRows); err != nil {
		utils.LogWarn("统计表 %s 的重复行失败：%v", table, err)
	}
	if extraRows > 0 {
		utils.LogWarn("表 %s 存在 %d 组重复的 (member_id, year)（多余 %d 行），按「保留 id 最小一条」清理后创建唯一索引",
			table, dupGroups, extraRows)
		delQuery := "DELETE f FROM " + table + " f JOIN (" +
			"SELECT MIN(id) AS keep_id, member_id, year FROM " + table +
			" GROUP BY member_id, year HAVING COUNT(*) > 1) d" +
			" ON d.member_id = f.member_id AND d.year = f.year WHERE f.id > d.keep_id"
		res := DB.Exec(delQuery)
		if res.Error != nil {
			utils.LogWarn("清理表 %s 的重复行失败（跳过创建索引）：%v", table, res.Error)
			return
		}
		utils.LogWarn("已清理表 %s 的重复行 %d 条", table, res.RowsAffected)
	}

	if err := DB.Exec("ALTER TABLE " + table + " ADD UNIQUE INDEX " + indexName + " (member_id, year)").Error; err != nil {
		utils.LogWarn("创建表 %s 的唯一索引 %s 失败（可手工执行 DEPLOY.md「手工 SQL」中的语句）：%v", table, indexName, err)
		return
	}
	utils.LogWarn("表 %s 已创建唯一索引 %s(member_id, year)", table, indexName)
}

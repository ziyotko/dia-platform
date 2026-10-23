// import-member-statistics 将 Excel「总表」导入会员专区年度统计表。
// 默认仅预览；必须显式传入 -apply 才会连接数据库并写入草稿。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"portal/config"
	"portal/internal/models"
	"portal/internal/services"
	"portal/pkg/utils"
)

func main() {
	file := flag.String("file", "", "Excel .xlsx 文件路径（只读取「总表」）")
	year := flag.Int("year", 2024, "统计年度")
	comparisonYear := flag.Int("compare-year", 2023, "对比年度（上一年）")
	column := flag.String("column", "数据中心", "目标会员栏目名称")
	title := flag.String("title", "", "数据集标题（可选）")
	apply := flag.Bool("apply", false, "确认写入数据库；默认只预览")
	flag.Parse()
	if strings.TrimSpace(*file) == "" {
		fmt.Fprintln(os.Stderr, "请通过 -file 指定 Excel 路径")
		flag.Usage()
		os.Exit(2)
	}
	parsed, err := services.ParseMemberStatisticsWorkbook(*file, *year, *comparisonYear)
	if err != nil {
		fmt.Fprintln(os.Stderr, "导入校验失败:", err)
		os.Exit(1)
	}
	preview, _ := json.MarshalIndent(parsed.Summary, "", "  ")
	fmt.Println(string(preview))
	if len(parsed.Summary.DuplicateNames) > 0 {
		fmt.Fprintln(os.Stderr, "注意：上述重复企业会按 Excel 原始行分别保留；记录数不等于去重企业数。")
	}
	if !*apply {
		fmt.Println("预览完成，未写入数据库。确认后添加 -apply 导入为草稿。")
		return
	}
	if err := config.InitImportConfig(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	utils.InitLogger()
	utils.InitDB()
	if err := utils.DB.AutoMigrate(&models.MemberStatisticsDataset{}, &models.MemberStatisticsRow{}); err != nil {
		fmt.Fprintln(os.Stderr, "创建统计数据表失败:", err)
		os.Exit(1)
	}
	id, err := services.ApplyMemberStatisticsImport(utils.DB, parsed, strings.TrimSpace(*column), strings.TrimSpace(*title))
	if err != nil {
		fmt.Fprintln(os.Stderr, "写入数据库失败（事务已回滚）:", err)
		os.Exit(1)
	}
	fmt.Printf("导入成功：数据集 ID=%d，%d 条记录，状态=草稿。\n", id, parsed.Summary.RecordCount)
}

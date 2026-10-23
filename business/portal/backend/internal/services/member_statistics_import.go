package services

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"portal/internal/models"
)

const memberStatisticsMaxFileSize = 20 << 20

type MemberStatisticsDuplicate struct {
	EnterpriseName string `json:"enterpriseName"`
	SourceRows     []int  `json:"sourceRows"`
}

type MemberStatisticsImportSummary struct {
	StatisticsYear  int                         `json:"statisticsYear"`
	ComparisonYear  int                         `json:"comparisonYear"`
	RecordCount     int                         `json:"recordCount"`
	EnterpriseCount int                         `json:"enterpriseCount"`
	ProvinceCount   int                         `json:"provinceCount"`
	RevenueYearWan  string                      `json:"revenueYearWan"`
	RevenuePriorWan string                      `json:"revenuePriorWan"`
	TotalYoYPercent string                      `json:"totalYoyPercent"`
	MissingRDYear   int                         `json:"missingRdYear"`
	MissingRDPrior  int                         `json:"missingRdPrior"`
	DuplicateNames  []MemberStatisticsDuplicate `json:"duplicateNames"`
	SourceSHA256    string                      `json:"sourceSha256"`
}

type MemberStatisticsImport struct {
	FileName string
	Rows     []models.MemberStatisticsRow
	Summary  MemberStatisticsImportSummary
}

// ParseMemberStatisticsWorkbook 只读取「总表」；数据透视表是派生结果，不重复导入。
// 此函数不连接数据库，可用于 -apply 前的无副作用预览。
func ParseMemberStatisticsWorkbook(path string, year, comparisonYear int) (*MemberStatisticsImport, error) {
	if year < 1900 || year > 2100 || comparisonYear != year-1 {
		return nil, errors.New("统计年度必须在 1900–2100 之间，对比年度须为上一年")
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("读取 Excel 文件失败: %w", err)
	}
	if info.IsDir() || info.Size() == 0 || info.Size() > memberStatisticsMaxFileSize {
		return nil, fmt.Errorf("Excel 文件须为非空文件，且不超过 %d MB", memberStatisticsMaxFileSize>>20)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 Excel 文件失败: %w", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(data), excelize.Options{})
	if err != nil {
		return nil, fmt.Errorf("打开 Excel 失败: %w", err)
	}
	defer f.Close()
	sheetIndex, err := f.GetSheetIndex("总表")
	if err != nil || sheetIndex < 0 {
		return nil, errors.New("Excel 缺少「总表」工作表")
	}
	rows, err := f.GetRows("总表", excelize.Options{RawCellValue: true})
	if err != nil {
		return nil, fmt.Errorf("读取「总表」失败: %w", err)
	}
	if len(rows) < 2 || len(rows) > 20001 {
		return nil, errors.New("「总表」须包含 1–20000 条数据")
	}
	if err := validateMemberStatisticsHeader(rows[0], year, comparisonYear); err != nil {
		return nil, err
	}

	sumYear, sumPrior := new(big.Rat), new(big.Rat)
	companyRows := map[string][]int{}
	provinces := map[string]struct{}{}
	importRows := make([]models.MemberStatisticsRow, 0, len(rows)-1)
	summary := MemberStatisticsImportSummary{
		StatisticsYear: year, ComparisonYear: comparisonYear,
		SourceSHA256:   fmt.Sprintf("%x", sha256.Sum256(data)),
		DuplicateNames: []MemberStatisticsDuplicate{},
	}
	for i, cells := range rows[1:] {
		rowNumber := i + 2
		if memberStatisticsEmptyRow(cells) {
			continue
		}
		row, err := parseMemberStatisticsRow(cells, rowNumber)
		if err != nil {
			return nil, err
		}
		companyRows[row.EnterpriseName] = append(companyRows[row.EnterpriseName], rowNumber)
		provinces[row.Province] = struct{}{}
		y, _ := new(big.Rat).SetString(row.RevenueYearWan)
		p, _ := new(big.Rat).SetString(row.RevenuePriorWan)
		sumYear.Add(sumYear, y)
		sumPrior.Add(sumPrior, p)
		if row.RDYearWan == nil {
			summary.MissingRDYear++
		}
		if row.RDPriorWan == nil {
			summary.MissingRDPrior++
		}
		importRows = append(importRows, row)
	}
	if len(importRows) == 0 {
		return nil, errors.New("「总表」没有有效数据行")
	}
	for name, sourceRows := range companyRows {
		if len(sourceRows) > 1 {
			summary.DuplicateNames = append(summary.DuplicateNames, MemberStatisticsDuplicate{name, sourceRows})
		}
	}
	sort.Slice(summary.DuplicateNames, func(i, j int) bool {
		return summary.DuplicateNames[i].SourceRows[0] < summary.DuplicateNames[j].SourceRows[0]
	})
	summary.RecordCount = len(importRows)
	summary.EnterpriseCount = len(companyRows)
	summary.ProvinceCount = len(provinces)
	summary.RevenueYearWan = trimDecimalZeros(sumYear.FloatString(10))
	summary.RevenuePriorWan = trimDecimalZeros(sumPrior.FloatString(10))
	if sumPrior.Sign() != 0 {
		yoy := new(big.Rat).Quo(sumYear, sumPrior)
		yoy.Sub(yoy, big.NewRat(1, 1))
		yoy.Mul(yoy, big.NewRat(100, 1))
		summary.TotalYoYPercent = yoy.FloatString(4)
	}
	return &MemberStatisticsImport{FileName: filepath.Base(path), Rows: importRows, Summary: summary}, nil
}

func validateMemberStatisticsHeader(header []string, year, prior int) error {
	want := []string{
		"企业名称", "主营业务领域", "地区", "一带", "一轴", "省市",
		fmt.Sprintf("主营业收入%d年（万元）", year),
		fmt.Sprintf("主营业收入%d年（万元）", prior),
		"同比增长",
		fmt.Sprintf("研发投入%d年（万元）", year),
		fmt.Sprintf("研发投入%d年（万元）", prior),
		fmt.Sprintf("研发投入占比%d", year),
		fmt.Sprintf("研发投入占比%d", prior),
		"数据来源",
	}
	if len(header) < len(want) {
		return fmt.Errorf("「总表」表头不完整：需要 %d 列，实际 %d 列", len(want), len(header))
	}
	for i, expected := range want {
		actual := normalizeMemberStatisticsHeader(header[i])
		if actual != normalizeMemberStatisticsHeader(expected) {
			return fmt.Errorf("「总表」第 %d 列表头应为 %q，实际为 %q", i+1, expected, header[i])
		}
	}
	return nil
}

func normalizeMemberStatisticsHeader(value string) string {
	value = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, value)
	return strings.ReplaceAll(value, "主营业务收入", "主营业收入")
}

func memberStatisticsEmptyRow(cells []string) bool {
	for _, cell := range cells {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func memberStatisticsCell(cells []string, index int) string {
	if index >= len(cells) {
		return ""
	}
	return strings.TrimSpace(cells[index])
}

func parseMemberStatisticsRow(cells []string, sourceRow int) (models.MemberStatisticsRow, error) {
	row := models.MemberStatisticsRow{SourceRow: sourceRow}
	row.EnterpriseName = memberStatisticsCell(cells, 0)
	row.SubField = memberStatisticsCell(cells, 1)
	row.Region = memberStatisticsCell(cells, 2)
	row.Province = memberStatisticsCell(cells, 5)
	row.Source = memberStatisticsCell(cells, 13)
	if row.EnterpriseName == "" || row.SubField == "" || row.Region == "" || row.Province == "" || row.Source == "" {
		return row, fmt.Errorf("第 %d 行：企业名称、领域、地区、省市和数据来源不能为空", sourceRow)
	}
	if !memberStatisticsFieldSet[row.SubField] || !memberStatisticsRegionSet[row.Region] || !memberStatisticsProvinceSet[row.Province] {
		return row, fmt.Errorf("第 %d 行：领域、地区或省市不在允许范围内", sourceRow)
	}
	for _, flag := range []struct {
		index int
		name  string
		dest  *int
	}{{3, "一带", &row.IsBelt}, {4, "一轴", &row.IsAxis}} {
		value := memberStatisticsCell(cells, flag.index)
		if value != "" && value != flag.name {
			return row, fmt.Errorf("第 %d 行：%s 列只能为空或填写「%s」", sourceRow, flag.name, flag.name)
		}
		if value != "" {
			*flag.dest = 1
		}
	}
	var err error
	if row.RevenueYearWan, err = memberStatisticsDecimal(memberStatisticsCell(cells, 6), false, 10); err != nil {
		return row, fmt.Errorf("第 %d 行：当年主营业务收入：%w", sourceRow, err)
	}
	if row.RevenuePriorWan, err = memberStatisticsDecimal(memberStatisticsCell(cells, 7), true, 10); err != nil {
		return row, fmt.Errorf("第 %d 行：上年主营业务收入：%w", sourceRow, err)
	}
	if row.YoYSource, err = memberStatisticsOptionalDecimal(memberStatisticsCell(cells, 8), 15); err != nil {
		return row, fmt.Errorf("第 %d 行：同比增长：%w", sourceRow, err)
	}
	if row.RDYearWan, err = memberStatisticsOptionalDecimal(memberStatisticsCell(cells, 9), 10); err != nil {
		return row, fmt.Errorf("第 %d 行：当年研发投入：%w", sourceRow, err)
	}
	if row.RDPriorWan, err = memberStatisticsOptionalDecimal(memberStatisticsCell(cells, 10), 10); err != nil {
		return row, fmt.Errorf("第 %d 行：上年研发投入：%w", sourceRow, err)
	}
	if row.RDShareYear, err = memberStatisticsOptionalDecimal(memberStatisticsCell(cells, 11), 15); err != nil {
		return row, fmt.Errorf("第 %d 行：当年研发投入占比：%w", sourceRow, err)
	}
	if row.RDSharePrior, err = memberStatisticsOptionalDecimal(memberStatisticsCell(cells, 12), 15); err != nil {
		return row, fmt.Errorf("第 %d 行：上年研发投入占比：%w", sourceRow, err)
	}
	if row.YoYSource != nil {
		current, _ := strconv.ParseFloat(row.RevenueYearWan, 64)
		prior, _ := strconv.ParseFloat(row.RevenuePriorWan, 64)
		provided, _ := strconv.ParseFloat(*row.YoYSource, 64)
		if math.Abs(current/prior-1-provided) > 0.001 {
			return row, fmt.Errorf("第 %d 行：同比增长与两年收入不一致", sourceRow)
		}
	}
	return row, nil
}

func memberStatisticsDecimal(value string, positive bool, places int) (string, error) {
	if value == "" {
		return "", errors.New("数值不能为空")
	}
	n, _, err := big.ParseFloat(value, 10, 256, big.ToNearestEven)
	if err != nil || n == nil || n.IsInf() || n.Sign() < 0 || (positive && n.Sign() == 0) || n.Cmp(new(big.Float).SetFloat64(1e17)) >= 0 {
		return "", fmt.Errorf("无效数值 %q", value)
	}
	return n.Text('f', places), nil
}

func memberStatisticsOptionalDecimal(value string, places int) (*string, error) {
	if value == "" {
		return nil, nil
	}
	n, _, err := big.ParseFloat(value, 10, 256, big.ToNearestEven)
	if err != nil || n == nil || n.IsInf() || n.Cmp(new(big.Float).SetFloat64(1e17)) >= 0 || n.Cmp(new(big.Float).SetFloat64(-1e17)) <= 0 {
		return nil, fmt.Errorf("无效数值 %q", value)
	}
	if places == 10 && n.Sign() < 0 {
		return nil, fmt.Errorf("金额不能为负数 %q", value)
	}
	formatted := n.Text('f', places)
	return &formatted, nil
}

func trimDecimalZeros(value string) string {
	return strings.TrimRight(strings.TrimRight(value, "0"), ".")
}

var memberStatisticsFieldSet = map[string]bool{
	"水": true, "气": true, "固废": true, "监测": true, "零部件": true,
	"材料药剂": true, "土壤": true, "噪声": true, "应急": true, "清洁生产": true, "环境服务": true,
}

var memberStatisticsRegionSet = map[string]bool{
	"华东": true, "华北": true, "华中": true, "华南": true,
	"西南": true, "东北": true, "西北": true,
}

var memberStatisticsProvinceSet = map[string]bool{
	"北京": true, "天津": true, "河北": true, "山西": true, "内蒙古": true,
	"辽宁": true, "吉林": true, "黑龙江": true, "上海": true, "江苏": true,
	"浙江": true, "安徽": true, "福建": true, "江西": true, "山东": true,
	"河南": true, "湖北": true, "湖南": true, "广东": true, "广西": true,
	"海南": true, "重庆": true, "四川": true, "贵州": true, "云南": true,
	"西藏": true, "陕西": true, "甘肃": true, "青海": true, "宁夏": true,
	"新疆": true,
}

// ApplyMemberStatisticsImport 在事务中替换同栏目、同年度的数据。
// 重复名称保留为不同来源行；重新导入后状态回到草稿，必须重新核对才能发布。
func ApplyMemberStatisticsImport(db *gorm.DB, parsed *MemberStatisticsImport, columnName, title string) (uint, error) {
	if db == nil || parsed == nil || len(parsed.Rows) == 0 {
		return 0, errors.New("没有可导入的数据")
	}
	var column models.MemberColumn
	if err := db.Where("name = ? AND status = ?", columnName, models.MemberColumnStatusEnabled).First(&column).Error; err != nil {
		return 0, fmt.Errorf("找不到已启用的会员栏目「%s」: %w", columnName, err)
	}
	if title == "" {
		title = fmt.Sprintf("环保装备制造业重点企业数据统计（%d年）", parsed.Summary.StatisticsYear)
	}
	if len([]rune(title)) > 200 {
		return 0, errors.New("数据集标题不能超过 200 字")
	}
	var datasetID uint
	err := db.Transaction(func(tx *gorm.DB) error {
		var dataset models.MemberStatisticsDataset
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("member_column_id = ? AND statistics_year = ?", column.ID, parsed.Summary.StatisticsYear).
			First(&dataset).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dataset = models.MemberStatisticsDataset{MemberColumnID: column.ID, StatisticsYear: parsed.Summary.StatisticsYear}
		}
		dataset.ComparisonYear = parsed.Summary.ComparisonYear
		dataset.Title = title
		dataset.SourceFileName = parsed.FileName
		dataset.SourceSHA256 = parsed.Summary.SourceSHA256
		dataset.ImportedAt = time.Now()
		dataset.Status = 0
		if dataset.ID == 0 {
			if err := tx.Create(&dataset).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&dataset).Updates(map[string]any{
				"comparison_year":  dataset.ComparisonYear,
				"title":            dataset.Title,
				"source_file_name": dataset.SourceFileName,
				"source_sha256":    dataset.SourceSHA256,
				"imported_at":      dataset.ImportedAt,
				"status":           0,
			}).Error; err != nil {
				return err
			}
			if err := tx.Where("dataset_id = ?", dataset.ID).Delete(&models.MemberStatisticsRow{}).Error; err != nil {
				return err
			}
		}
		datasetID = dataset.ID
		importRows := make([]models.MemberStatisticsRow, len(parsed.Rows))
		copy(importRows, parsed.Rows)
		for i := range importRows {
			importRows[i].ID = 0
			importRows[i].DatasetID = datasetID
		}
		return tx.CreateInBatches(importRows, 200).Error
	})
	return datasetID, err
}

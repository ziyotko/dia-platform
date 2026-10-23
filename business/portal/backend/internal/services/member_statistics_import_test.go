package services

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func memberStatisticsTestWorkbook(t *testing.T, corruptGrowth bool) string {
	t.Helper()
	f := excelize.NewFile()
	if err := f.SetSheetName("Sheet1", "总表"); err != nil {
		t.Fatal(err)
	}
	header := []any{
		"企业名称", "主营业务领域", "地区", "一带", "一轴", "省市",
		"主营业收入2024年（万元）", "主营业收入2023年（万元）", "同比增长",
		"研发投入2024年（万元）", "研发投入2023年（万元）",
		"研发投入占比2024", "研发投入占比2023", "数据来源",
	}
	first := []any{"甲企业 ", "水", "华东", "一带", "", "江苏", 120.5, 100, 0.205, "", 10, "", 0.1, "规范企业"}
	secondGrowth := 0.2
	if corruptGrowth {
		secondGrowth = 0.8
	}
	second := []any{"甲企业", "水", "华东", "", "一轴", "江苏", 60, 50, secondGrowth, 5, 4, 0.083333333333333, 0.08, "鼓励目录"}
	for i, row := range [][]any{header, first, second} {
		cell, err := excelize.CoordinatesToCellName(1, i+1)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.SetSheetRow("总表", cell, &row); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(t.TempDir(), "sample.xlsx")
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseMemberStatisticsWorkbookPreservesRowsAndMissingValues(t *testing.T) {
	parsed, err := ParseMemberStatisticsWorkbook(memberStatisticsTestWorkbook(t, false), 2024, 2023)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Summary.RecordCount != 2 || parsed.Summary.EnterpriseCount != 1 || parsed.Summary.ProvinceCount != 1 {
		t.Fatalf("unexpected counts: %+v", parsed.Summary)
	}
	if parsed.Summary.RevenueYearWan != "180.5" || parsed.Summary.RevenuePriorWan != "150" || parsed.Summary.TotalYoYPercent != "20.3333" {
		t.Fatalf("unexpected totals: %+v", parsed.Summary)
	}
	if parsed.Summary.MissingRDYear != 1 || parsed.Rows[0].RDYearWan != nil {
		t.Fatalf("missing R&D amount was not preserved: %+v", parsed.Rows[0])
	}
	if len(parsed.Summary.DuplicateNames) != 1 || parsed.Summary.DuplicateNames[0].SourceRows[0] != 2 || parsed.Summary.DuplicateNames[0].SourceRows[1] != 3 {
		t.Fatalf("duplicate enterprise not reported: %+v", parsed.Summary.DuplicateNames)
	}
	if parsed.Rows[0].SourceRow != 2 || parsed.Rows[1].SourceRow != 3 {
		t.Fatalf("source row numbers lost: %+v", parsed.Rows)
	}
}

func TestParseMemberStatisticsWorkbookRejectsMismatchedGrowth(t *testing.T) {
	_, err := ParseMemberStatisticsWorkbook(memberStatisticsTestWorkbook(t, true), 2024, 2023)
	if err == nil || !strings.Contains(err.Error(), "第 3 行") {
		t.Fatalf("expected row-specific growth error, got %v", err)
	}
}

package services

import (
	"testing"

	"portal/internal/models"
)

func TestBuildMemberStatisticsDetailUsesDistinctCompaniesAndTotalRevenue(t *testing.T) {
	parsed, err := ParseMemberStatisticsWorkbook(memberStatisticsTestWorkbook(t, false), 2024, 2023)
	if err != nil {
		t.Fatal(err)
	}
	dataset := models.MemberStatisticsDataset{StatisticsYear: 2024, ComparisonYear: 2023}
	detail, err := buildMemberStatisticsDetail(dataset, "数据中心", parsed.Rows)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Summary.RecordCount != 2 || detail.Summary.EnterpriseCount != 1 || detail.Summary.ProvinceCount != 1 {
		t.Fatalf("unexpected counts: %+v", detail.Summary)
	}
	if detail.Summary.RevenueYearWan != "180.5" || detail.Summary.RevenuePriorWan != "150" || detail.Summary.TotalYoYPercent != "20.3333" {
		t.Fatalf("unexpected totals: %+v", detail.Summary)
	}
	if detail.Summary.MissingRDYear != 1 || detail.Summary.MissingRDPrior != 0 {
		t.Fatalf("unexpected missing values: %+v", detail.Summary)
	}
	if len(detail.DuplicateNames) != 1 || detail.DuplicateNames[0].SourceRows[0] != 2 || detail.DuplicateNames[0].SourceRows[1] != 3 {
		t.Fatalf("unexpected duplicates: %+v", detail.DuplicateNames)
	}
}

func TestBuildMemberStatisticsDetailRejectsInvalidAmount(t *testing.T) {
	_, err := buildMemberStatisticsDetail(models.MemberStatisticsDataset{}, "数据中心", []models.MemberStatisticsRow{{RevenueYearWan: "bad", RevenuePriorWan: "100"}})
	if err == nil {
		t.Fatal("expected invalid amount to fail")
	}
}

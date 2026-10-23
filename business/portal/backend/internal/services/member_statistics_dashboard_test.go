package services

import (
	"os"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"portal/internal/models"
)

func TestMemberStatisticsDashboardWithMySQL(t *testing.T) {
	dsn := os.Getenv("DIA_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set DIA_TEST_MYSQL_DSN to verify the dashboard against MySQL")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewMemberStatisticsService(db).GetPublishedDashboard(MemberStatisticsDashboardQuery{}); err != nil {
		t.Fatalf("query published dashboard: %v", err)
	}
}

func TestMemberStatisticsDashboardAggregatesFilteredRows(t *testing.T) {
	rows := []models.MemberStatisticsRow{
		{EnterpriseName: "甲", Province: "广东", Region: "华南", SubField: "水", RevenueYearWan: "100.25", RevenuePriorWan: "80"},
		{EnterpriseName: "甲", Province: "广东", Region: "华南", SubField: "水", RevenueYearWan: "50.25", RevenuePriorWan: "40"},
		{EnterpriseName: "乙", Province: "江苏", Region: "华东", SubField: "气", RevenueYearWan: "10", RevenuePriorWan: "20"},
	}
	summary, err := memberStatisticsDashboardSummary(rows[:2])
	if err != nil {
		t.Fatal(err)
	}
	if summary.RecordCount != 2 || summary.EnterpriseCount != 1 || summary.ProvinceCount != 1 || summary.RevenueYearWan != "150.5" || summary.TotalYoYPercent != "25.4167" {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	groups, err := memberStatisticsDashboardGroups(rows, "region")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 || groups[0].Name != "华南" || groups[0].EnterpriseCount != 1 || groups[1].Name != "华东" || groups[1].YoYPercent != "-50.0000" {
		t.Fatalf("unexpected groups: %+v", groups)
	}
	filters := memberStatisticsFilterOptions(rows)
	if len(filters.Provinces) != 2 || len(filters.Regions) != 2 || len(filters.SubFields) != 2 {
		t.Fatalf("unexpected filters: %+v", filters)
	}
}

func TestMemberStatisticsDashboardBeltAxisFilters(t *testing.T) {
	rows := []models.MemberStatisticsRow{
		{Region: "华东", IsBelt: 0, IsAxis: 0},
		{Region: "华东", IsBelt: 1, IsAxis: 0},
		{Region: "华东", IsBelt: 0, IsAxis: 1},
		{Region: "华东", IsBelt: 1, IsAxis: 1},
	}
	zero, one := 0, 1
	for _, test := range []struct {
		name  string
		query MemberStatisticsDashboardQuery
		want  int
	}{
		{"全部", MemberStatisticsDashboardQuery{}, 4},
		{"一带是", MemberStatisticsDashboardQuery{IsBelt: &one}, 2},
		{"一带否", MemberStatisticsDashboardQuery{IsBelt: &zero}, 2},
		{"一轴是", MemberStatisticsDashboardQuery{IsAxis: &one}, 2},
		{"一轴否", MemberStatisticsDashboardQuery{IsAxis: &zero}, 2},
		{"两项都是", MemberStatisticsDashboardQuery{IsBelt: &one, IsAxis: &one}, 1},
		{"一带是且一轴否", MemberStatisticsDashboardQuery{IsBelt: &one, IsAxis: &zero, Region: "华东"}, 1},
		{"其他地区", MemberStatisticsDashboardQuery{IsBelt: &one, Region: "华南"}, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			count := 0
			for _, row := range rows {
				if memberStatisticsMatchesDashboardQuery(row, test.query) {
					count++
				}
			}
			if count != test.want {
				t.Fatalf("matched %d rows, want %d", count, test.want)
			}
		})
	}
}

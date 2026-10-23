package services

import (
	"math/big"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"portal/internal/models"
)

// MemberStatisticsDashboardQuery is the public, aggregate-only query for active member pages.
type MemberStatisticsDashboardQuery struct {
	Year     int
	Region   string
	Province string
	SubField string
	IsBelt   *int
	IsAxis   *int
}

type MemberStatisticsGroup struct {
	Name            string `json:"name"`
	RecordCount     int    `json:"recordCount"`
	EnterpriseCount int    `json:"enterpriseCount"`
	RevenueYearWan  string `json:"revenueYearWan"`
	RevenuePriorWan string `json:"revenuePriorWan"`
	YoYPercent      string `json:"yoyPercent"`
}

type MemberStatisticsDashboard struct {
	Years          []int                   `json:"years"`
	StatisticsYear int                     `json:"statisticsYear,omitempty"`
	ComparisonYear int                     `json:"comparisonYear,omitempty"`
	Title          string                  `json:"title,omitempty"`
	ImportedAt     string                  `json:"importedAt,omitempty"`
	Filters        MemberStatisticsFilters `json:"filters"`
	Summary        MemberStatisticsSummary `json:"summary"`
	Provinces      []MemberStatisticsGroup `json:"provinces"`
	Regions        []MemberStatisticsGroup `json:"regions"`
	SubFields      []MemberStatisticsGroup `json:"subFields"`
}

// GetPublishedDashboard never reads draft/offline datasets or disabled columns.
// The response contains aggregates only; enterprise names and source rows stay private.
func (s *MemberStatisticsService) GetPublishedDashboard(q MemberStatisticsDashboardQuery) (*MemberStatisticsDashboard, error) {
	var datasets []models.MemberStatisticsDataset
	err := s.db.Model(&models.MemberStatisticsDataset{}).
		Joins("JOIN member_column ON member_column.id = member_statistics_dataset.member_column_id").
		Where("member_column.name = ? AND member_column.status = ? AND member_statistics_dataset.status = ?",
			"数据中心", models.MemberColumnStatusEnabled, models.MemberStatisticsStatusPublished).
		Order("member_statistics_dataset.statistics_year DESC, member_statistics_dataset.id DESC").
		Find(&datasets).Error
	if err != nil {
		return nil, err
	}
	result := &MemberStatisticsDashboard{
		Years: []int{}, Filters: MemberStatisticsFilters{Provinces: []string{}, Regions: []string{}, SubFields: []string{}},
		Provinces: []MemberStatisticsGroup{}, Regions: []MemberStatisticsGroup{}, SubFields: []MemberStatisticsGroup{},
	}
	for _, dataset := range datasets {
		result.Years = append(result.Years, dataset.StatisticsYear)
	}
	if len(datasets) == 0 {
		return result, nil
	}
	selected := datasets[0]
	if q.Year != 0 {
		found := false
		for _, dataset := range datasets {
			if dataset.StatisticsYear == q.Year {
				selected, found = dataset, true
				break
			}
		}
		if !found {
			return nil, gorm.ErrRecordNotFound
		}
	}
	result.StatisticsYear, result.ComparisonYear = selected.StatisticsYear, selected.ComparisonYear
	result.Title, result.ImportedAt = selected.Title, selected.ImportedAt.Format("2006-01-02")
	var rows []models.MemberStatisticsRow
	if err := s.db.Select("enterprise_name, province, region, sub_field, is_belt, is_axis, revenue_year_wan, revenue_prior_wan").
		Where("dataset_id = ?", selected.ID).Find(&rows).Error; err != nil {
		return nil, err
	}
	result.Filters = memberStatisticsFilterOptions(rows)
	filtered := make([]models.MemberStatisticsRow, 0, len(rows))
	for _, row := range rows {
		if !memberStatisticsMatchesDashboardQuery(row, q) {
			continue
		}
		filtered = append(filtered, row)
	}
	var buildErr error
	result.Summary, buildErr = memberStatisticsDashboardSummary(filtered)
	if buildErr != nil {
		return nil, buildErr
	}
	for _, group := range []struct {
		field  string
		target *[]MemberStatisticsGroup
	}{{"province", &result.Provinces}, {"region", &result.Regions}, {"subField", &result.SubFields}} {
		*group.target, buildErr = memberStatisticsDashboardGroups(filtered, group.field)
		if buildErr != nil {
			return nil, buildErr
		}
	}
	return result, nil
}

func memberStatisticsMatchesDashboardQuery(row models.MemberStatisticsRow, q MemberStatisticsDashboardQuery) bool {
	return (q.Region == "" || row.Region == q.Region) &&
		(q.Province == "" || row.Province == q.Province) &&
		(q.SubField == "" || row.SubField == q.SubField) &&
		(q.IsBelt == nil || row.IsBelt == *q.IsBelt) &&
		(q.IsAxis == nil || row.IsAxis == *q.IsAxis)
}

func memberStatisticsFilterOptions(rows []models.MemberStatisticsRow) MemberStatisticsFilters {
	p, r, f := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, row := range rows {
		if row.Province != "" {
			p[row.Province] = true
		}
		if row.Region != "" {
			r[row.Region] = true
		}
		if row.SubField != "" {
			f[row.SubField] = true
		}
	}
	return MemberStatisticsFilters{Provinces: memberStatisticsSortedKeys(p), Regions: memberStatisticsSortedKeys(r), SubFields: memberStatisticsSortedKeys(f)}
}

func memberStatisticsDashboardSummary(rows []models.MemberStatisticsRow) (MemberStatisticsSummary, error) {
	summary := MemberStatisticsSummary{RecordCount: len(rows)}
	companies, provinces := map[string]bool{}, map[string]bool{}
	year, prior := new(big.Rat), new(big.Rat)
	for _, row := range rows {
		companies[row.EnterpriseName] = true
		provinces[row.Province] = true
		current, okCurrent := new(big.Rat).SetString(row.RevenueYearWan)
		previous, okPrevious := new(big.Rat).SetString(row.RevenuePriorWan)
		if !okCurrent || !okPrevious {
			return summary, strconv.ErrSyntax
		}
		year.Add(year, current)
		prior.Add(prior, previous)
	}
	summary.EnterpriseCount, summary.ProvinceCount = len(companies), len(provinces)
	summary.RevenueYearWan, summary.RevenuePriorWan = trimDecimalZeros(year.FloatString(10)), trimDecimalZeros(prior.FloatString(10))
	if prior.Sign() != 0 {
		yoy := new(big.Rat).Sub(new(big.Rat).Quo(year, prior), big.NewRat(1, 1))
		summary.TotalYoYPercent = new(big.Rat).Mul(yoy, big.NewRat(100, 1)).FloatString(4)
	}
	return summary, nil
}

func memberStatisticsDashboardGroups(rows []models.MemberStatisticsRow, field string) ([]MemberStatisticsGroup, error) {
	buckets := map[string][]models.MemberStatisticsRow{}
	for _, row := range rows {
		name := row.Province
		if field == "region" {
			name = row.Region
		}
		if field == "subField" {
			name = row.SubField
		}
		if strings.TrimSpace(name) != "" {
			buckets[name] = append(buckets[name], row)
		}
	}
	groups := make([]MemberStatisticsGroup, 0, len(buckets))
	for name, groupRows := range buckets {
		summary, err := memberStatisticsDashboardSummary(groupRows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, MemberStatisticsGroup{
			Name: name, RecordCount: summary.RecordCount, EnterpriseCount: summary.EnterpriseCount,
			RevenueYearWan: summary.RevenueYearWan, RevenuePriorWan: summary.RevenuePriorWan, YoYPercent: summary.TotalYoYPercent,
		})
	}
	sort.Slice(groups, func(i, j int) bool {
		iAmount, _ := strconv.ParseFloat(groups[i].RevenueYearWan, 64)
		jAmount, _ := strconv.ParseFloat(groups[j].RevenueYearWan, 64)
		if iAmount == jAmount {
			return groups[i].Name < groups[j].Name
		}
		return iAmount > jAmount
	})
	return groups, nil
}

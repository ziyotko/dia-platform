package services

import (
	"errors"
	"math/big"
	"sort"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"portal/internal/models"
)

type MemberStatisticsService struct{ db *gorm.DB }

func NewMemberStatisticsService(db *gorm.DB) *MemberStatisticsService {
	return &MemberStatisticsService{db: db}
}

type MemberStatisticsSummary struct {
	RecordCount     int    `json:"recordCount"`
	EnterpriseCount int    `json:"enterpriseCount"`
	ProvinceCount   int    `json:"provinceCount"`
	RevenueYearWan  string `json:"revenueYearWan"`
	RevenuePriorWan string `json:"revenuePriorWan"`
	TotalYoYPercent string `json:"totalYoyPercent"`
	MissingRDYear   int    `json:"missingRdYear"`
	MissingRDPrior  int    `json:"missingRdPrior"`
}

type MemberStatisticsFilters struct {
	Provinces []string `json:"provinces"`
	Regions   []string `json:"regions"`
	SubFields []string `json:"subFields"`
}

type MemberStatisticsDetail struct {
	Dataset        models.MemberStatisticsDataset `json:"dataset"`
	ColumnName     string                         `json:"columnName"`
	Summary        MemberStatisticsSummary        `json:"summary"`
	DuplicateNames []MemberStatisticsDuplicate    `json:"duplicateNames"`
	Filters        MemberStatisticsFilters        `json:"filters"`
}

type MemberStatisticsRowQuery struct {
	EnterpriseName string
	Province       string
	Region         string
	SubField       string
	MissingRD      bool
	Page           int
	PageSize       int
}

func (s *MemberStatisticsService) ListDatasets() ([]models.MemberStatisticsDataset, error) {
	var datasets []models.MemberStatisticsDataset
	err := s.db.Order("statistics_year DESC, id DESC").Find(&datasets).Error
	return datasets, err
}

// GetDatasetDetail 用原始明细计算汇总；不信任 Excel 的数据透视表或预存汇总。
func (s *MemberStatisticsService) GetDatasetDetail(id uint) (*MemberStatisticsDetail, error) {
	var dataset models.MemberStatisticsDataset
	if err := s.db.First(&dataset, id).Error; err != nil {
		return nil, err
	}
	var column models.MemberColumn
	if err := s.db.Select("name").First(&column, dataset.MemberColumnID).Error; err != nil {
		return nil, err
	}
	var rows []models.MemberStatisticsRow
	if err := s.db.Select("enterprise_name, source_row, province, region, sub_field, revenue_year_wan, revenue_prior_wan, rd_year_wan, rd_prior_wan").
		Where("dataset_id = ?", id).Order("source_row ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return buildMemberStatisticsDetail(dataset, column.Name, rows)
}

func buildMemberStatisticsDetail(dataset models.MemberStatisticsDataset, columnName string, rows []models.MemberStatisticsRow) (*MemberStatisticsDetail, error) {
	companyRows := map[string][]int{}
	provinces := map[string]bool{}
	regions := map[string]bool{}
	subFields := map[string]bool{}
	sumYear, sumPrior := new(big.Rat), new(big.Rat)
	summary := MemberStatisticsSummary{RecordCount: len(rows)}
	for _, row := range rows {
		companyRows[row.EnterpriseName] = append(companyRows[row.EnterpriseName], row.SourceRow)
		provinces[row.Province] = true
		regions[row.Region] = true
		subFields[row.SubField] = true
		current, okCurrent := new(big.Rat).SetString(row.RevenueYearWan)
		prior, okPrior := new(big.Rat).SetString(row.RevenuePriorWan)
		if !okCurrent || !okPrior {
			return nil, errors.New("统计明细中存在无效收入数值")
		}
		sumYear.Add(sumYear, current)
		sumPrior.Add(sumPrior, prior)
		if row.RDYearWan == nil {
			summary.MissingRDYear++
		}
		if row.RDPriorWan == nil {
			summary.MissingRDPrior++
		}
	}
	summary.EnterpriseCount = len(companyRows)
	summary.ProvinceCount = len(provinces)
	summary.RevenueYearWan = trimDecimalZeros(sumYear.FloatString(10))
	summary.RevenuePriorWan = trimDecimalZeros(sumPrior.FloatString(10))
	if sumPrior.Sign() > 0 {
		yoy := new(big.Rat).Quo(sumYear, sumPrior)
		yoy.Sub(yoy, big.NewRat(1, 1))
		yoy.Mul(yoy, big.NewRat(100, 1))
		summary.TotalYoYPercent = yoy.FloatString(4)
	}
	duplicates := make([]MemberStatisticsDuplicate, 0)
	for name, sourceRows := range companyRows {
		if len(sourceRows) > 1 {
			duplicates = append(duplicates, MemberStatisticsDuplicate{EnterpriseName: name, SourceRows: sourceRows})
		}
	}
	sort.Slice(duplicates, func(i, j int) bool {
		return duplicates[i].SourceRows[0] < duplicates[j].SourceRows[0]
	})
	return &MemberStatisticsDetail{
		Dataset: dataset, ColumnName: columnName, Summary: summary,
		DuplicateNames: duplicates,
		Filters: MemberStatisticsFilters{
			Provinces: memberStatisticsSortedKeys(provinces),
			Regions:   memberStatisticsSortedKeys(regions),
			SubFields: memberStatisticsSortedKeys(subFields),
		},
	}, nil
}

func memberStatisticsSortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (s *MemberStatisticsService) GetRows(datasetID uint, query MemberStatisticsRowQuery) ([]models.MemberStatisticsRow, int64, error) {
	var exists int64
	if err := s.db.Model(&models.MemberStatisticsDataset{}).Where("id = ?", datasetID).Count(&exists).Error; err != nil {
		return nil, 0, err
	}
	if exists == 0 {
		return nil, 0, gorm.ErrRecordNotFound
	}
	q := s.db.Model(&models.MemberStatisticsRow{}).Where("dataset_id = ?", datasetID)
	if name := strings.TrimSpace(query.EnterpriseName); name != "" {
		if len([]rune(name)) > 100 {
			return nil, 0, errors.New("企业名称筛选过长")
		}
		q = q.Where("enterprise_name LIKE ?", "%"+name+"%")
	}
	if query.Province != "" {
		q = q.Where("province = ?", query.Province)
	}
	if query.Region != "" {
		q = q.Where("region = ?", query.Region)
	}
	if query.SubField != "" {
		q = q.Where("sub_field = ?", query.SubField)
	}
	if query.MissingRD {
		q = q.Where("rd_year_wan IS NULL OR rd_prior_wan IS NULL")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.MemberStatisticsRow
	err := q.Order("source_row ASC").Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).Find(&rows).Error
	return rows, total, err
}

func (s *MemberStatisticsService) UpdateStatus(id uint, status int) error {
	if status != models.MemberStatisticsStatusPublished && status != models.MemberStatisticsStatusOffline {
		return errors.New("状态只能是已发布或已下线")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var dataset models.MemberStatisticsDataset
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&dataset, id).Error; err != nil {
			return err
		}
		if status == models.MemberStatisticsStatusPublished {
			var column models.MemberColumn
			if err := tx.First(&column, dataset.MemberColumnID).Error; err != nil {
				return err
			}
			if column.Status != models.MemberColumnStatusEnabled {
				return errors.New("所属会员栏目已禁用，不能发布")
			}
			var count int64
			if err := tx.Model(&models.MemberStatisticsRow{}).Where("dataset_id = ?", id).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return errors.New("数据集没有明细，不能发布")
			}
		}
		return tx.Model(&dataset).Update("status", status).Error
	})
}

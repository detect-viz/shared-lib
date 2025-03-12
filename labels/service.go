package labels

import (
	"encoding/csv"
	"mime/multipart"
	"slices"
	"sort"

	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/storage/mysql"
	"github.com/google/wire"
)

var LabelSet = wire.NewSet(
	NewService,
	wire.Bind(new(Service), new(*serviceImpl)),
)

// Service 標籤服務
type serviceImpl struct {
	mysql *mysql.Client
}

// NewService 創建標籤服務
func NewService(mysql *mysql.Client) *serviceImpl {
	return &serviceImpl{mysql: mysql}
}

// Create 創建標籤
func (s *serviceImpl) Create(realm string, input *models.LabelDTO) (*models.LabelDTO, error) {
	label := models.LabelKey{RealmName: realm, KeyName: input.Key}
	result, err := s.mysql.CreateLabel(&label, input.Values)
	if err != nil {
		return nil, err
	}
	values := make([]string, len(result.Values))
	for i, value := range result.Values {
		values[i] = value.Value
	}
	return &models.LabelDTO{ID: result.ID, Key: result.KeyName, Values: values}, nil
}

// Get 查詢單個標籤
func (s *serviceImpl) Get(id int64) (*models.LabelDTO, error) {
	label, err := s.mysql.GetLabel(id)
	if err != nil {
		return nil, err
	}
	values := make([]string, len(label.Values))
	for i, value := range label.Values {
		values[i] = value.Value
	}
	return &models.LabelDTO{ID: label.ID, Key: label.KeyName, Values: values}, nil
}

// List 列出所有標籤（支援分頁）
func (s *serviceImpl) List(realm string, cursor int64, limit int) ([]models.LabelDTO, int64, error) {

	labels, total, err := s.mysql.ListLabels(realm, cursor, limit)
	if err != nil {
		return nil, 0, err
	}

	labelDTOs := make([]models.LabelDTO, len(labels))
	for i, label := range labels {
		values := make([]string, len(label.Values))
		for j, value := range label.Values {
			values[j] = value.Value
		}
		labelDTOs[i] = models.LabelDTO{ID: label.ID, Key: label.KeyName, Values: values}
	}
	return labelDTOs, total, nil
}

// Update 更新標籤
func (s *serviceImpl) Update(realm string, input *models.LabelDTO) (*models.LabelDTO, error) {
	label := models.LabelKey{ID: input.ID, KeyName: input.Key}
	updates := input.Values
	newLabel, err := s.mysql.UpdateLabel(label, updates)
	if err != nil {
		return nil, err
	}
	values := make([]string, len(newLabel.Values))
	for i, value := range newLabel.Values {
		values[i] = value.Value
	}
	return &models.LabelDTO{ID: newLabel.ID, Key: newLabel.KeyName, Values: values}, nil
}

// UpdateKey 更新標籤 Key
func (s *serviceImpl) UpdateKeyName(realm, oldKey, newKey string) (*models.LabelDTO, error) {
	label, err := s.mysql.UpdateLabelKeyName(realm, oldKey, newKey)
	if err != nil {
		return nil, err
	}
	values := make([]string, len(label.Values))
	for i, value := range label.Values {
		values[i] = value.Value
	}
	return &models.LabelDTO{ID: label.ID, Key: label.KeyName, Values: values}, nil
}

// Delete 刪除標籤
func (s *serviceImpl) Delete(id int64) error {
	return s.mysql.DeleteLabel(id)
}

// GetKeyOptions 查詢標籤的可選值
func (s *serviceImpl) GetKeyOptions(realm string) ([]models.OptionResponse, error) {
	options := []models.OptionResponse{}
	labels, _, err := s.List(realm, 1000, 0)
	if err != nil {
		return nil, err
	}

	for _, label := range labels {
		options = append(options, models.OptionResponse{
			Text:  label.Key,
			Value: label.Key,
		})
	}
	return options, nil
}

// ExportCSV 下載標籤 CSV
func (s *serviceImpl) ExportCSV(realm string) ([][]string, error) {

	// 獲取數據
	labels, _, err := s.List(realm, 0, 1000)
	if err != nil {
		return nil, err
	}

	if len(labels) == 0 {
		return [][]string{{"Key", "Value"}}, nil
	}

	// 寫入標題
	res := [][]string{{"Key", "Value"}}

	// 排序
	sort.Slice(labels, func(i, j int) bool {
		return labels[i].Key < labels[j].Key
	})

	// 寫入數據
	for _, label := range labels {
		for _, value := range label.Values {
			res = append(res, []string{label.Key, value})
		}
	}
	return res, nil
}

// ImportCSV 匯入標籤 CSV
func (s *serviceImpl) ImportCSV(realm string, f *multipart.FileHeader) error {
	// 開啟檔案
	src, err := f.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// 讀取 CSV 內容
	reader := csv.NewReader(src)
	rows, err := reader.ReadAll()
	if err != nil {
		return err
	}

	// 解析標籤
	keys := []string{}
	values := [][]string{}
	for i, row := range rows {
		if i == 0 {
			continue // 跳過標題列
		}
		if !slices.Contains(keys, row[0]) {
			keys = append(keys, row[0])
			values = append(values, []string{})
		}
		idx := slices.Index(keys, row[0])
		values[idx] = append(values[idx], row[1])
	}

	labels := make([]models.LabelDTO, len(keys))
	for i, key := range keys {
		labels[i] = models.LabelDTO{
			Key:    key,
			Values: values[i],
		}
	}
	// 批量寫入 DB
	err = s.mysql.BulkCreateOrUpdateLabel(realm, labels)
	if err != nil {
		return err
	}
	return nil
}

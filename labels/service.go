package labels

import (
	"encoding/csv"
	"net/http"

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
func (s *serviceImpl) Create(realm string, label *models.Label) (*models.Label, error) {
	label.RealmName = realm
	return s.mysql.CreateLabel(label)
}

// Get 查詢單個標籤
func (s *serviceImpl) Get(realm, key string) (*models.Label, error) {
	return s.mysql.GetLabel(realm, key)
}

// List 列出所有標籤（支援分頁）
func (s *serviceImpl) List(realm string, limit, offset int) ([]models.Label, error) {
	return s.mysql.ListLabels(realm, limit, offset)
}

// Update 更新標籤
func (s *serviceImpl) Update(realm, key string, updates map[string]interface{}) error {
	return s.mysql.UpdateLabel(realm, key, updates)
}

// UpdateKey 更新標籤 Key
func (s *serviceImpl) UpdateKey(realm, oldKey, newKey string) error {
	return s.mysql.UpdateKey(realm, oldKey, newKey)
}

// Delete 刪除標籤
func (s *serviceImpl) Delete(realm, key string) error {
	return s.mysql.DeleteLabel(realm, key)
}

// Exists 檢查標籤是否存在
func (s *serviceImpl) Exists(realm, key string) (bool, error) {
	return s.mysql.ExistsLabel(realm, key)
}

// GetKeyOptions 查詢標籤的可選值
func (s *serviceImpl) GetKeyOptions(realm, key string) ([]models.OptionResponse, error) {
	labels, err := s.List(realm, 1000, 0)
	if err != nil {
		return nil, err
	}

	for _, label := range labels {
		if label.KeyName == key {
			return []models.OptionResponse{
				{
					Text:  label.Value.String(),
					Value: label.Value.String(),
				},
			}, nil
		}
	}
	return nil, nil
}

// BulkCreateOrUpdate 批量新增或更新標籤
func (s *serviceImpl) BulkCreateOrUpdate(realm string, labels []models.Label) ([]models.Label, error) {
	return s.mysql.BulkCreateOrUpdateLabel(realm, labels)
}

// ExportCSV 下載標籤 CSV
func (s *serviceImpl) ExportCSV(w http.ResponseWriter, r *http.Request) {
	realm := r.URL.Query().Get("realm")
	labels, err := s.List(realm, 1000, 0)
	if err != nil {
		http.Error(w, "Failed to fetch labels", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=labels.csv")
	w.Header().Set("Content-Type", "text/csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// 寫入標題
	writer.Write([]string{"Key", "Value"})

	// 寫入數據
	for _, label := range labels {
		writer.Write([]string{
			label.KeyName,
			label.Value.String(),
		})
	}
}

package mysql

import (
	"github.com/detect-viz/shared-lib/models"
)

// 創建
func (c *Client) CreateTarget(target *models.Target) (*models.Target, error) {

	if err := c.db.Create(&target).Error; err != nil {
		return nil, ParseDBError(err)
	}

	return target, nil
}

// 是否為新的監控對象
func (c *Client) CheckTargetExists(realm, dataSource, resourceName, partitionName string) (bool, error) {
	var count int64
	err := c.db.Debug().Table("targets").Where("data_source = ? AND resource_name = ? AND partition_name = ? AND realm_name = ?", dataSource, resourceName, partitionName, realm).Count(&count).Error
	if err != nil {
		return false, ParseDBError(err)
	}
	return count > 0, nil
}

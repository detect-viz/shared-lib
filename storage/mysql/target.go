package mysql

import (
	"github.com/detect-viz/shared-lib/models"
)

// 創建
func (c *Client) CreateTarget(target *models.Target) (*models.Target, error) {
	target.ID = GenerateUUID16()
	if err := c.db.Create(&target).Error; err != nil {
		return nil, ParseDBError(err)
	}

	return target, nil
}

// 是否為新的監控對象
func (c *Client) CheckTargetExists(realm, dataSourceType, resourceName, partitionName string) (bool, error) {
	var count int64
	err := c.db.Debug().Table("targets").Where("datasource_name = ? AND resource_name = ? AND partition_name = ? AND realm_name = ?", dataSourceType, resourceName, partitionName, realm).Count(&count).Error
	if err != nil {
		return false, ParseDBError(err)
	}
	return count > 0, nil
}

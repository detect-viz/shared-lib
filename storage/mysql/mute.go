package mysql

import (
	"fmt"

	"github.com/detect-viz/shared-lib/models"
)

// CreateMute 創建抑制規則
func (c *Client) CreateMute(mute *models.Mute) error {
	return c.db.Create(mute).Error
}

// GetMute 獲取抑制規則
func (c *Client) GetMute(id int64) (*models.Mute, error) {
	var mute models.Mute
	if err := c.db.Preload("ResourceGroups").First(&mute, id).Error; err != nil {
		return nil, fmt.Errorf("獲取抑制規則失敗: %w", err)
	}
	return &mute, nil
}

// ListMutes 獲取抑制規則列表
func (c *Client) ListMutes(realm string) ([]models.Mute, error) {
	var mutes []models.Mute
	if err := c.db.Preload("ResourceGroups").Where("realm_name = ?", realm).Find(&mutes).Error; err != nil {
		return nil, fmt.Errorf("獲取抑制規則列表失敗: %w", err)
	}
	return mutes, nil
}

// UpdateMute 更新抑制規則
func (c *Client) UpdateMute(mute *models.Mute) error {
	return c.db.Model(mute).Updates(mute).Error
}

// DeleteMute 刪除抑制規則
func (c *Client) DeleteMute(id int64) error {
	return c.db.Delete(&models.Mute{}, id).Error
}

// GetMuteResourceGroups 獲取規則關聯
func (c *Client) GetMuteResourceGroups(muteID int64) ([]models.ResourceGroup, error) {
	var mute models.Mute
	if err := c.db.Preload("ResourceGroups").First(&mute, muteID).Error; err != nil {
		return nil, fmt.Errorf("獲取抑制規則失敗: %w", err)
	}
	return mute.ResourceGroups, nil
}

// GetMutesByResourceGroup 獲取資源組的靜音規則
func (c *Client) GetMutesByResourceGroup(resourceGroupID int64) ([]models.Mute, error) {
	var mutes []models.Mute
	if err := c.db.Preload("ResourceGroups").
		Joins("JOIN mute_resource_groups ON mutes.id = mute_resource_groups.mute_id").
		Where("mute_resource_groups.resource_group_id = ?", resourceGroupID).
		Find(&mutes).Error; err != nil {
		return nil, fmt.Errorf("獲取靜音規則失敗: %w", err)
	}
	return mutes, nil
}

// CheckMuteName 檢查名稱是否重複
func (c *Client) CheckMuteName(mute models.Mute) (bool, string) {
	var count int64
	c.db.Model(&models.Mute{}).
		Where("realm_name = ? AND name = ? AND id != ?", mute.RealmName, mute.Name, mute.ID).
		Count(&count)

	if count > 0 {
		return false, "name existed in other mute rule"
	}
	return true, ""
}

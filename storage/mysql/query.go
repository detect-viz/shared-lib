package mysql

// Exists 檢查是否有相同 name
func (c *Client) Exists(realm, table, column, value string) (bool, error) {
	var count int64
	err := c.db.Debug().Table(table).Where("realm_name = ? AND "+column+" = ? AND deleted_at IS NULL", realm, value).Count(&count).Error
	if err != nil {
		return false, ParseDBError(err)
	}
	return count > 0, nil
}

// 檢查是否有相同 name，但排除自身 ID
func (c *Client) ExistsExcludeSelf(realm, table, column, value string, excludeID int64) (bool, error) {
	var count int64
	err := c.db.Table(table).
		Where("realm_name = ? AND id <> ? AND deleted_at IS NULL", realm, excludeID).
		Where(column+" = ?", value). // 避免 SQL 拼接風險
		Count(&count).Error

	if err != nil {
		return false, ParseDBError(err)
	}
	return count > 0, nil
}

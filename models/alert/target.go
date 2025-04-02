package alert

import "github.com/detect-viz/shared-lib/models/common"

// * 監控對象
// * 監控對象是監控數據的來源，可以是主機、容器、數據庫等
// * 監控對象可以有不同的數據源，比如 nmon、njmon、logman、sysstat 等
// * 監控對象可以有不同的指標類別，比如 cpu、memory、disk、network、system 等
// * 監控對象可以有不同的採集頻率，比如 10 秒、1 分鐘、5 分鐘等
// * 監控對象可以有不同的發送間隔，比如 10 秒、1 分鐘、5 分鐘等

type Target struct {
	ID        []byte `json:"id"`
	RealmName string `json:"realm_name"`

	DatasourceName     string `json:"datasource_name"`
	Category           string `json:"category"`
	CollectionInterval int    `json:"collection_interval"`
	ReportingInterval  int    `json:"reporting_interval"`
	IsHidden           bool   `json:"is_hidden"`
	ResourceName       string `json:"resource_name"`
	PartitionName      string `json:"partition_name"`
	common.AuditUserModel
	common.AuditTimeModel
}

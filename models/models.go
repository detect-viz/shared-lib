package models

import (
	"github.com/detect-viz/shared-lib/models/alert"
	"github.com/detect-viz/shared-lib/models/common"
	"github.com/detect-viz/shared-lib/models/config"
	"github.com/detect-viz/shared-lib/models/dto"
	"github.com/detect-viz/shared-lib/models/label"
	"github.com/detect-viz/shared-lib/models/mute"

	"github.com/detect-viz/shared-lib/models/notifier"
	"github.com/detect-viz/shared-lib/models/parser"
	"github.com/detect-viz/shared-lib/models/resource"
	"github.com/detect-viz/shared-lib/models/scheduler"
	"github.com/detect-viz/shared-lib/models/template"
)

// 統一對外暴露所有 model
type (
	// DTO 相關
	LabelDTO = dto.LabelDTO

	// Response 相關
	MetricRuleOverview  = alert.MetricRuleOverview
	RuleOverview        = alert.RuleOverview
	RuleResponse        = alert.RuleResponse
	RuleContactResponse = alert.RuleContactResponse
	ContactResponse     = alert.ContactResponse

	// Alert 相關
	Rule               = alert.Rule
	Target             = alert.Target
	MetricRule         = alert.MetricRule
	Template           = alert.Template
	RuleState          = alert.RuleState
	Contact            = alert.Contact
	RuleContact        = alert.RuleContact
	TriggeredLog       = alert.TriggeredLog
	NotifyLog          = alert.NotifyLog
	TriggeredLogIDsMap = alert.TriggeredLogIDsMap

	//* Alert Input Schema
	AlertPayload = alert.AlertPayload
	Metadata     = alert.Metadata
	MetricValue  = alert.MetricValue

	// Notify 相關
	EmailSetting  = notifier.EmailSetting
	TeamsSetting  = notifier.TeamsSetting
	LineConfig    = notifier.LineConfig
	WebhookConfig = notifier.WebhookConfig

	// Common 相關
	AuditUserModel = common.AuditUserModel
	AuditTimeModel = common.AuditTimeModel
	JSONMap        = common.JSONMap
	SeveritySet    = common.SeveritySet
	Response       = common.Response
	OptionResponse = common.OptionResponse
	MetricResponse = common.MetricResponse
	SSOUser        = common.SSOUser
	RotateSetting  = common.RotateSetting
	RotateTask     = common.RotateTask
	NotifySetting  = common.NotifySetting
	Task           = common.Task
	TaskInfo       = common.TaskInfo

	// Resource 相關
	ResourceGroup = resource.ResourceGroup
	Resource      = resource.Resource

	// Config 相關
	Config         = config.Config
	GlobalConfig   = config.GlobalConfig
	ServerConfig   = config.ServerConfig
	ParserConfig   = config.ParserConfig
	LoggerConfig   = config.LoggerConfig
	DatabaseConfig = config.DatabaseConfig
	InfluxDBConfig = config.InfluxDBConfig
	MySQLConfig    = config.MySQLConfig
	AlertConfig    = config.AlertConfig
	Code           = config.Code
	KeycloakConfig = config.KeycloakConfig

	// Template 相關
	TemplateData      = template.TemplateData
	ResourceGroupInfo = template.ResourceGroupInfo
	TriggeredInfo     = template.TriggeredInfo
	DefaultTemplate   = template.DefaultTemplate

	// Scheduler 相關
	JobStatus = scheduler.JobStatus

	// Logger 相關

	// Parser 相關
	FileInfo = parser.FileInfo
	//MetricValue    = parser.MetricValue
	Tags           = parser.Tags
	MetricField    = parser.MetricField
	MetricLibrarys = parser.MetricLibrarys

	// Label 相關
	LabelKey   = label.LabelKey
	LabelValue = label.LabelValue

	// Mute 相關
	Mute              = mute.Mute
	TimeRange         = mute.TimeRange
	MuteResourceGroup = mute.MuteResourceGroup

	//RuleLabelValue       = alert.RuleLabelValue
)

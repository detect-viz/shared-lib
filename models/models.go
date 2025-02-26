package models

import (
	"github.com/detect-viz/shared-lib/models/alert"
	"github.com/detect-viz/shared-lib/models/common"
	"github.com/detect-viz/shared-lib/models/config"
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
	// Alert 相關
	Rule                 = alert.Rule
	AlertRuleDetail      = alert.AlertRuleDetail
	AlertState           = alert.AlertState
	Contact              = alert.Contact
	Template             = alert.Template
	AlertMessage         = alert.AlertMessage
	AlertMessageTemplate = alert.AlertMessageTemplate
	CheckRule            = alert.CheckRule
	TriggerLog           = alert.TriggerLog
	NotifyLog            = alert.NotifyLog
	NotifyTriggerLog     = alert.NotifyTriggerLog
	AlertRuleContact     = alert.AlertRuleContact
	JSONMap              = alert.JSONMap
	CurrentAlert         = alert.CurrentAlert
	HistoryAlert         = alert.HistoryAlert
	RuleConfig           = alert.RuleConfig
	AlertContactSeverity = alert.AlertContactSeverity
	ContactDefinition    = alert.ContactDefinition
	ContactConfig        = alert.ContactConfig
	RuleDefinition       = alert.RuleDefinition
	ThresholdConfig      = alert.ThresholdConfig
	AlertRuleLabel       = alert.AlertRuleLabel

	// Mute 相關
	Mute              = mute.Mute
	TimeRange         = mute.TimeRange
	MuteResourceGroup = mute.MuteResourceGroup

	// Notify 相關
	EmailSetting  = notifier.EmailSetting
	TeamsSetting  = notifier.TeamsSetting
	LineConfig    = notifier.LineConfig
	WebhookConfig = notifier.WebhookConfig

	// Label 相關
	Label = label.Label

	// Common 相關
	Common         = common.Common
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

	// Parser 相關
	FileInfo       = parser.FileInfo
	MetricValue    = parser.MetricValue
	Tags           = parser.Tags
	MetricField    = parser.MetricField
	MetricLibrarys = parser.MetricLibrarys

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
	MetricRule     = config.MetricRule
	KeycloakConfig = config.KeycloakConfig
	// Template 相關
	TemplateData      = template.TemplateData
	ResourceGroupInfo = template.ResourceGroupInfo
	TriggerInfo       = template.TriggerInfo
	DefaultTemplate   = template.DefaultTemplate

	// Scheduler 相關
	JobStatus = scheduler.JobStatus

	// Logger 相關

)

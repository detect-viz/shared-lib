package models

import (
	"shared-lib/models/alert"
	"shared-lib/models/common"
	"shared-lib/models/config"
	"shared-lib/models/label"
	"shared-lib/models/notify"
	"shared-lib/models/parser"
	"shared-lib/models/scheduler"
	"shared-lib/models/template"
)

// 統一對外暴露所有 model
type (
	// Alert 相關
	AlertRule              = alert.AlertRule
	AlertRuleDetail        = alert.AlertRuleDetail
	AlertState             = alert.AlertState
	AlertMute              = alert.AlertMute
	AlertContact           = alert.AlertContact
	AlertTemplate          = alert.AlertTemplate
	AlertMessage           = alert.AlertMessage
	AlertMessageTemplate   = alert.AlertMessageTemplate
	CheckRule              = alert.CheckRule
	MetricRule             = alert.MetricRule
	AlertMuteRule          = alert.AlertMuteRule
	TriggerLog             = alert.TriggerLog
	NotificationLog        = alert.NotificationLog
	NotificationTriggerLog = alert.NotificationTriggerLog
	AlertRuleContact       = alert.AlertRuleContact
	RepeatTime             = alert.RepeatTime
	JSONMap                = alert.JSONMap

	// Notify 相關
	ChannelType   = notify.ChannelType
	NotifyStatus  = notify.NotifyStatus
	EmailSetting  = notify.EmailSetting
	TeamsSetting  = notify.TeamsSetting
	LineConfig    = notify.LineConfig
	WebhookConfig = notify.WebhookConfig

	// Label 相關
	Label = label.Label

	// Common 相關
	Common         = common.Common
	Response       = common.Response
	OptionResponse = common.OptionResponse
	SSOUser        = common.SSOUser
	RotateSetting  = common.RotateSetting
	RotateTask     = common.RotateTask
	// Parser 相關
	FileInfo       = parser.FileInfo
	MetricValue    = parser.MetricValue
	Tags           = parser.Tags
	MetricField    = parser.MetricField
	MetricLibrarys = parser.MetricLibrarys

	// Config 相關
	Config           = config.Config
	ServerConfig     = config.ServerConfig
	ParserConfig     = config.ParserConfig
	LoggerConfig     = config.LoggerConfig
	DatabaseConfig   = config.DatabaseConfig
	AlertConfig      = config.AlertConfig
	Code             = config.Code
	ChannelConfig    = config.ChannelConfig
	NotifyConfig     = config.NotifyConfig
	MetricSpecConfig = config.MetricSpecConfig

	// Template 相關
	TemplateData      = template.TemplateData
	ResourceGroupInfo = template.ResourceGroupInfo
	TriggerInfo       = template.TriggerInfo
	DefaultTemplate   = template.DefaultTemplate

	// Scheduler 相關
	SchedulerConfig = scheduler.SchedulerConfig
	SchedulerJob    = scheduler.SchedulerJob
	JobStatus       = scheduler.JobStatus

	// Logger 相關

)

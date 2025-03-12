package config

import (
	"github.com/detect-viz/shared-lib/models/alert"
)

// Config 全局配置結構
type Config struct {
	Global   *GlobalConfig  `mapstructure:"global" yaml:"global"`
	Logger   LoggerConfig   `mapstructure:"logger" yaml:"logger"`
	Database DatabaseConfig `mapstructure:"database" yaml:"database"`
	Alert    AlertConfig    `mapstructure:"alert" yaml:"alert"`
	Server   ServerConfig   `mapstructure:"server" yaml:"server"`
	Keycloak KeycloakConfig `mapstructure:"keycloak" yaml:"keycloak"`
}

// 標準狀態值
type GlobalConfig struct {
	Code        Codes                       `yaml:"code" json:"code"`
	Metric      Metrics                     `yaml:"metric" json:"metric"`
	Tag         Tags                        `yaml:"tag" json:"tag"`
	MetricRules map[string]alert.MetricRule `yaml:"metric_rules" json:"metric_rules"`
	Templates   []alert.Template            `yaml:"templates" json:"templates"`
}
type Codes struct {
	Severity struct {
		Crit Code `yaml:"crit" json:"crit"`
		Warn Code `yaml:"warn" json:"warn"`
		Info Code `yaml:"info" json:"info"`
	} `yaml:"severity" json:"severity"`
	State struct {
		// 追蹤規則狀態
		Rule struct {
			Alerting Code `yaml:"alerting" json:"alerting"`
			Resolved Code `yaml:"resolved" json:"resolved"`
			Normal   Code `yaml:"normal" json:"normal"`
			Disabled Code `yaml:"disabled" json:"disabled"`
		} `yaml:"rule" json:"rule"`
		// 追蹤通知狀態
		Contact struct {
			Normal   Code `yaml:"normal" json:"normal"`
			Muting   Code `yaml:"muting" json:"muting"`
			Silence  Code `yaml:"silence" json:"silence"`
			Disabled Code `yaml:"disabled" json:"disabled"`
		} `yaml:"contact" json:"contact"`
		// 追蹤通知結果
		Notify struct {
			Sent    Code `yaml:"sent" json:"sent"`
			Solved  Code `yaml:"solved" json:"solved"`
			Pending Code `yaml:"pending" json:"pending"`
			Delayed Code `yaml:"delayed" json:"delayed"`
			Failed  Code `yaml:"failed" json:"failed"`
		} `yaml:"notify" json:"notify"`
	}
	ChannelType struct {
		Discord Code `yaml:"discord" json:"discord"`
		Line    Code `yaml:"line" json:"line"`
		Mail    Code `yaml:"mail" json:"mail"`
		Slack   Code `yaml:"slack" json:"slack"`
		Teams   Code `yaml:"teams" json:"teams"`
		Webex   Code `yaml:"webex" json:"webex"`
		Webhook Code `yaml:"webhook" json:"webhook"`
	} `yaml:"channel_type" json:"channel_type"`
	HealthState struct {
		Disconnected Code `yaml:"disconnected" json:"disconnected"`
		Error        Code `yaml:"error" json:"error"`
		Healthy      Code `yaml:"healthy" json:"healthy"`
		Warning      Code `yaml:"warning" json:"warning"`
	} `yaml:"health_state" json:"health_state"`
	MuteState struct {
		Active    Code `yaml:"active"`
		Disable   Code `yaml:"disable"`
		Ended     Code `yaml:"ended"`
		Scheduled Code `yaml:"scheduled" json:"scheduled"`
	} `yaml:"mute_state" json:"mute_state"`
}
type Metrics struct {
	CPU struct {
		Usage  MetricSpec `yaml:"usage" json:"usage"`
		Idle   MetricSpec `yaml:"idle" json:"idle"`
		System MetricSpec `yaml:"system" json:"system"`
		User   MetricSpec `yaml:"user" json:"user"`
		IOWait MetricSpec `yaml:"iowait" json:"iowait"`
		Nice   MetricSpec `yaml:"nice" json:"nice"`
		Steal  MetricSpec `yaml:"steal" json:"steal"`
	} `yaml:"cpu" json:"cpu"`

	Memory struct {
		TotalBytes  MetricSpec `yaml:"total_bytes" json:"total_bytes"`
		UsedBytes   MetricSpec `yaml:"used_bytes" json:"used_bytes"`
		BufferBytes MetricSpec `yaml:"buffer_bytes" json:"buffer_bytes"`
		CacheBytes  MetricSpec `yaml:"cache_bytes" json:"cache_bytes"`
		BufferUsage MetricSpec `yaml:"buffer_usage" json:"buffer_usage"`
		CacheUsage  MetricSpec `yaml:"cache_usage" json:"cache_usage"`
		Usage       MetricSpec `yaml:"usage" json:"usage"`
		FreeUsage   MetricSpec `yaml:"free_usage" json:"free_usage"`
	} `yaml:"memory" json:"memory"`

	Swap struct {
		TotalBytes MetricSpec `yaml:"total_bytes" json:"total_bytes"`
		UsedBytes  MetricSpec `yaml:"used_bytes" json:"used_bytes"`
		Usage      MetricSpec `yaml:"usage" json:"usage"`
	} `yaml:"swap" json:"swap"`

	Network struct {
		SentBytes   MetricSpec `yaml:"sent_bytes" json:"sent_bytes"`
		RecvBytes   MetricSpec `yaml:"recv_bytes" json:"recv_bytes"`
		SentPackets MetricSpec `yaml:"sent_packets" json:"sent_packets"`
		RecvPackets MetricSpec `yaml:"recv_packets" json:"recv_packets"`
		SentErrors  MetricSpec `yaml:"sent_errs" json:"sent_errs"`
		RecvErrors  MetricSpec `yaml:"recv_errs" json:"recv_errs"`
	} `yaml:"network" json:"network"`

	Disk struct {
		Busy        MetricSpec `yaml:"busy" json:"busy"`
		ReadBytes   MetricSpec `yaml:"read_bytes" json:"read_bytes"`
		WriteBytes  MetricSpec `yaml:"write_bytes" json:"write_bytes"`
		Reads       MetricSpec `yaml:"reads" json:"reads"`
		Writes      MetricSpec `yaml:"writes" json:"writes"`
		Xfers       MetricSpec `yaml:"xfers" json:"xfers"`
		Rqueue      MetricSpec `yaml:"rqueue" json:"rqueue"`
		Wqueue      MetricSpec `yaml:"wqueue" json:"wqueue"`
		QueueLength MetricSpec `yaml:"queue_length" json:"queue_length"`
		IOTime      MetricSpec `yaml:"io_time" json:"io_time"`
		ReadTime    MetricSpec `yaml:"read_time" json:"read_time"`
		WriteTime   MetricSpec `yaml:"write_time" json:"write_time"`
	} `yaml:"disk" json:"disk"`

	Filesystem struct {
		TotalBytes MetricSpec `yaml:"total_bytes" json:"total_bytes"`
		FreeBytes  MetricSpec `yaml:"free_bytes" json:"free_bytes"`
		UsedBytes  MetricSpec `yaml:"used_bytes" json:"used_bytes"`
		Usage      MetricSpec `yaml:"usage" json:"usage"`
		FreeUsage  MetricSpec `yaml:"free_usage" json:"free_usage"`
	} `yaml:"filesystem" json:"filesystem"`

	Process struct {
		PID        MetricSpec `yaml:"pid" json:"pid"`
		PsCPUUsage MetricSpec `yaml:"ps_cpu_usage" json:"ps_cpu_usage"`
		PsMemUsage MetricSpec `yaml:"ps_mem_usage" json:"ps_mem_usage"`
	} `yaml:"process" json:"process"`

	System struct {
		Uptime MetricSpec `yaml:"uptime" json:"uptime"`
	} `yaml:"system" json:"system"`

	Connection struct {
		Refused MetricSpec `yaml:"refused" json:"refused"`
		Status  MetricSpec `yaml:"status" json:"status"`
	} `yaml:"connection" json:"connection"`

	Tablespace struct {
		Autoextensible  MetricSpec `yaml:"autoextensible" json:"autoextensible"`
		FilesCount      MetricSpec `yaml:"files_count" json:"files_count"`
		TotalSpaceBytes MetricSpec `yaml:"total_space_bytes" json:"total_space_bytes"`
		UsedSpaceBytes  MetricSpec `yaml:"used_space_bytes" json:"used_space_bytes"`
		FreeSpaceBytes  MetricSpec `yaml:"free_space_bytes" json:"free_space_bytes"`
		UsedUsage       MetricSpec `yaml:"used_usage" json:"used_usage"`
		FreeUsage       MetricSpec `yaml:"free_usage" json:"free_usage"`
		MaxSizeBytes    MetricSpec `yaml:"max_size_bytes" json:"max_size_bytes"`
		MaxFreeBytes    MetricSpec `yaml:"max_free_bytes" json:"max_free_bytes"`
		AutoUsedUsage   MetricSpec `yaml:"auto_used_usage" json:"auto_used_usage"`
		AutoFreeUsage   MetricSpec `yaml:"auto_free_usage" json:"auto_free_usage"`
	} `yaml:"tablespace" json:"tablespace"`
}
type Tags struct {
	Base struct {
		Host      string `yaml:"host" json:"host"`
		Database  string `yaml:"database" json:"database"`
		Source    string `yaml:"source" json:"source"`
		Os        string `yaml:"os" json:"os"`
		IP        string `yaml:"ip" json:"ip"`
		Timestamp string `yaml:"timestamp" json:"timestamp"`
		Value     string `yaml:"value" json:"value"`
		Metric    string `yaml:"metric" json:"metric"`
	} `yaml:"base" json:"base"`
	Os struct {
		Aix     string `yaml:"aix" json:"aix"`
		Linux   string `yaml:"linux" json:"linux"`
		Windows string `yaml:"windows" json:"windows"`
	} `yaml:"os" json:"os"`
	Source struct {
		Nmon             string `yaml:"nmon" json:"nmon"`
		Njmon            string `yaml:"njmon" json:"njmon"`
		Logman           string `yaml:"logman" json:"logman"`
		Sysstat          string `yaml:"sysstat" json:"sysstat"`
		OracleAwrrpt     string `yaml:"oracle_awrrpt" json:"oracle_awrrpt"`
		OracleTablespace string `yaml:"oracle_tablespace" json:"oracle_tablespace"`
		OracleConnection string `yaml:"oracle_connection" json:"oracle_connection"`
	} `yaml:"source" json:"source"`
	Partition struct {
		TablespaceName string `yaml:"tablespace_name" json:"tablespace_name"`
		CPUName        string `yaml:"cpu_name" json:"cpu_name"`
		DiskName       string `yaml:"disk_name" json:"disk_name"`
		FilesystemName string `yaml:"filesystem_name" json:"filesystem_name"`
		SwapName       string `yaml:"swap_name" json:"swap_name"`
		ProcessName    string `yaml:"process_name" json:"process_name"`
		NetworkName    string `yaml:"network_name" json:"network_name"`
	} `yaml:"partition" json:"partition"`
	Category struct {
		CPU        string `yaml:"cpu" json:"cpu"`
		Memory     string `yaml:"memory" json:"memory"`
		Swap       string `yaml:"swap" json:"swap"`
		Network    string `yaml:"network" json:"network"`
		Disk       string `yaml:"disk" json:"disk"`
		Filesystem string `yaml:"filesystem" json:"filesystem"`
		Process    string `yaml:"process" json:"process"`
		System     string `yaml:"system" json:"system"`
		Connection string `yaml:"connection" json:"connection"`
		Tablespace string `yaml:"tablespace" json:"tablespace"`
	} `yaml:"category" json:"category"`
	Sql struct {
		ID     string `yaml:"id" json:"id"`
		Module string `yaml:"module" json:"module"`
		Text   string `yaml:"text" json:"text"`
	} `yaml:"sql" json:"sql"`
}
type Code struct {
	Name  string `yaml:"name" json:"name"`
	Alias string `yaml:"alias" json:"alias"`
	Desc  string `yaml:"desc" json:"desc"`
}

package config

type MappingConfig struct {
	Code   Codes   `yaml:"code"`
	Metric Metrics `yaml:"metric"`
	Tag    Tags    `yaml:"tag"`
}
type Codes struct {
	Severity struct {
		Crit Code `yaml:"crit"`
		Warn Code `yaml:"warn"`
		Info Code `yaml:"info"`
	} `yaml:"severity"`
	State struct {
		// 追蹤規則狀態
		Rule struct {
			Alerting Code `yaml:"alerting"`
			Resolved Code `yaml:"resolved"`
			Normal   Code `yaml:"normal"`
			Disable  Code `yaml:"disable"`
		} `yaml:"rule"`
		// 追蹤通知狀態
		Contact struct {
			Normal  Code `yaml:"normal"`
			Muting  Code `yaml:"muting"`
			Silence Code `yaml:"silence"`
		} `yaml:"contact"`
		// 追蹤通知結果
		Notify struct {
			Solved Code `yaml:"solved"`
			Failed Code `yaml:"failed"`
		} `yaml:"notify"`
		Trigger struct {
			Unresolved Code `yaml:"unresolved"`
			Resolved   Code `yaml:"resolved"`
		} `yaml:"trigger"`
	}
	ChannelType struct {
		Discord Code `yaml:"discord"`
		Line    Code `yaml:"line"`
		Mail    Code `yaml:"mail"`
		Slack   Code `yaml:"slack"`
		Teams   Code `yaml:"teams"`
		Webex   Code `yaml:"webex"`
		Webhook Code `yaml:"webhook"`
	} `yaml:"channel_type"`
	HealthState struct {
		Disconnected Code `yaml:"disconnected"`
		Error        Code `yaml:"error"`
		Healthy      Code `yaml:"healthy"`
		Warning      Code `yaml:"warning"`
	} `yaml:"health_state"`
	MuteState struct {
		Active    Code `yaml:"active"`
		Disable   Code `yaml:"disable"`
		Ended     Code `yaml:"ended"`
		Scheduled Code `yaml:"scheduled"`
	} `yaml:"mute_state"`
}

type Metrics struct {
	CPU struct {
		Usage  MetricSpec `yaml:"usage"`
		Idle   MetricSpec `yaml:"idle"`
		System MetricSpec `yaml:"system"`
		User   MetricSpec `yaml:"user"`
		IOWait MetricSpec `yaml:"iowait"`
		Nice   MetricSpec `yaml:"nice"`
		Steal  MetricSpec `yaml:"steal"`
	} `yaml:"cpu"`

	Memory struct {
		TotalBytes  MetricSpec `yaml:"total_bytes"`
		UsedBytes   MetricSpec `yaml:"used_bytes"`
		BufferBytes MetricSpec `yaml:"buffer_bytes"`
		CacheBytes  MetricSpec `yaml:"cache_bytes"`
		BufferUsage MetricSpec `yaml:"buffer_usage"`
		CacheUsage  MetricSpec `yaml:"cache_usage"`
		Usage       MetricSpec `yaml:"usage"`
		FreeUsage   MetricSpec `yaml:"free_usage"`
	} `yaml:"memory"`

	Swap struct {
		TotalBytes MetricSpec `yaml:"total_bytes"`
		UsedBytes  MetricSpec `yaml:"used_bytes"`
		Usage      MetricSpec `yaml:"usage"`
	} `yaml:"swap"`

	Network struct {
		SentBytes   MetricSpec `yaml:"sent_bytes"`
		RecvBytes   MetricSpec `yaml:"recv_bytes"`
		SentPackets MetricSpec `yaml:"sent_packets"`
		RecvPackets MetricSpec `yaml:"recv_packets"`
		SentErrors  MetricSpec `yaml:"sent_errs"`
		RecvErrors  MetricSpec `yaml:"recv_errs"`
	} `yaml:"network"`

	Disk struct {
		Busy        MetricSpec `yaml:"busy"`
		ReadBytes   MetricSpec `yaml:"read_bytes"`
		WriteBytes  MetricSpec `yaml:"write_bytes"`
		Reads       MetricSpec `yaml:"reads"`
		Writes      MetricSpec `yaml:"writes"`
		Xfers       MetricSpec `yaml:"xfers"`
		Rqueue      MetricSpec `yaml:"rqueue"`
		Wqueue      MetricSpec `yaml:"wqueue"`
		QueueLength MetricSpec `yaml:"queue_length"`
		IOTime      MetricSpec `yaml:"io_time"`
		ReadTime    MetricSpec `yaml:"read_time"`
		WriteTime   MetricSpec `yaml:"write_time"`
	} `yaml:"disk"`

	Filesystem struct {
		TotalBytes MetricSpec `yaml:"total_bytes"`
		FreeBytes  MetricSpec `yaml:"free_bytes"`
		UsedBytes  MetricSpec `yaml:"used_bytes"`
		Usage      MetricSpec `yaml:"usage"`
		FreeUsage  MetricSpec `yaml:"free_usage"`
	} `yaml:"filesystem"`

	Process struct {
		PID        MetricSpec `yaml:"pid"`
		PsCPUUsage MetricSpec `yaml:"ps_cpu_usage"`
		PsMemUsage MetricSpec `yaml:"ps_mem_usage"`
	} `yaml:"process"`

	System struct {
		Uptime MetricSpec `yaml:"uptime"`
	} `yaml:"system"`

	Connection struct {
		Refused MetricSpec `yaml:"refused"`
		Status  MetricSpec `yaml:"status"`
	} `yaml:"connection"`

	Tablespace struct {
		Autoextensible MetricSpec `yaml:"autoextensible"`
		FilesCount     MetricSpec `yaml:"files_count"`
		TotalSpaceGB   MetricSpec `yaml:"total_space_gb"`
		UsedSpaceGB    MetricSpec `yaml:"used_space_gb"`
		FreeSpaceGB    MetricSpec `yaml:"free_space_gb"`
		UsedPct        MetricSpec `yaml:"used_pct"`
		FreePct        MetricSpec `yaml:"free_pct"`
		MaxSizeGB      MetricSpec `yaml:"max_size_gb"`
		MaxFreeGB      MetricSpec `yaml:"max_free_gb"`
		AutoUsedPct    MetricSpec `yaml:"auto_used_pct"`
		AutoFreePct    MetricSpec `yaml:"auto_free_pct"`
	} `yaml:"tablespace"`
}
type Tags struct {
	Base struct {
		Host      string `json:"host"`
		Database  string `json:"database"`
		Source    string `json:"source"`
		Os        string `json:"os"`
		IP        string `json:"ip"`
		Timestamp string `json:"timestamp"`
		Value     string `json:"value"`
		Metric    string `json:"metric"`
	} `json:"base"`
	Os struct {
		Aix     string `json:"aix"`
		Linux   string `json:"linux"`
		Windows string `json:"windows"`
	} `json:"os"`
	Source struct {
		Nmon             string `json:"nmon"`
		Njmon            string `json:"njmon"`
		Logman           string `json:"logman"`
		Sysstat          string `json:"sysstat"`
		OracleAwrrpt     string `json:"oracle_awrrpt"`
		OracleTablespace string `json:"oracle_tablespace"`
		OracleConnection string `json:"oracle_connection"`
	} `json:"source"`
	Partition struct {
		TablespaceName string `json:"tablespace_name"`
		CPUName        string `json:"cpu_name"`
		DiskName       string `json:"disk_name"`
		FilesystemName string `json:"filesystem_name"`
		SwapName       string `json:"swap_name"`
		ProcessName    string `json:"process_name"`
		NetworkName    string `json:"network_name"`
	} `json:"partition"`
	Category struct {
		CPU        string `json:"cpu"`
		Memory     string `json:"memory"`
		Swap       string `json:"swap"`
		Network    string `json:"network"`
		Disk       string `json:"disk"`
		Filesystem string `json:"filesystem"`
		Process    string `json:"process"`
		System     string `json:"system"`
		Connection string `json:"connection"`
		Tablespace string `json:"tablespace"`
	} `json:"category"`
	Sql struct {
		ID     string `json:"id"`
		Module string `json:"module"`
		Text   string `json:"text"`
	} `json:"sql"`
}
type Code struct {
	Name  string `yaml:"name"`
	Alias string `yaml:"alias"`
	Desc  string `yaml:"desc"`
}

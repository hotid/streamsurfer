package structures

import (
	"crypto/sha256"
	"time"
)

// custom values for HTML-templates and reports
type ConfigStub struct {
	Name string `yaml:"name,omitempty"`
}

type ConfigZabbix struct {
	DiscoveryPath   string   `yaml:"discovery-path,omitempty"`
	DiscoveryGroups []string `yaml:"discovery-groups,omitempty"`
	NameTemplate    string   `yaml:"name-template,omitempty"`
	TitleTemplate   string   `yaml:"title-template,omitempty"`
}

type Config struct {
	GroupParams      map[Key]*ConfigGroup
	GroupStreams     map[Key]map[Key]Stream // map[groupname]stream
	Stubs            ConfigStub
	Zabbix           ConfigZabbix
	Samples          []string
	ListenHTTP       string
	User             string
	Pass             string
	UserAgents       []string
	ErrorLog         string
	ExpireDurationDB time.Duration // measured in hours
	IsReady          chan bool     // config parsed and ready to use
}

func (cfg *Config) Params(groupName string) ConfigGroup {
	if data, ok := cfg.GroupParams[sha256.Sum256([]byte(groupName))]; ok {
		return *data
	} else {
		return ConfigGroup{}
	}
}

type Params struct {
	ProbersHTTP            uint          `yaml:"http-probers,omitempty"`              // num of
	ProbersHLS             uint          `yaml:"hls-probers,omitempty"`               // num of
	ProbersHDS             uint          `yaml:"hds-probers,omitempty"`               // num of
	ProbersWV              uint          `yaml:"wv-probers,omitempty"`                // num of
	MediaProbers           uint          `yaml:"media-probers,omitempty"`             // num of
	CheckBrokenTime        uint          `yaml:"check-broken-time"`                   // ms
	ConnectTimeout         time.Duration `yaml:"connect-timeout,omitempty"`           // sec
	RWTimeout              time.Duration `yaml:"rw-timeout,omitempty"`                // sec
	SlowWarningTimeout     time.Duration `yaml:"slow-warning-timeout,omitempty"`      // sec
	VerySlowWarningTimeout time.Duration `yaml:"very-slow-warning-timeout,omitempty"` // sec
	TimeBetweenTasks       time.Duration `yaml:"time-between-tasks,omitempty"`        // sec
	TaskTTL                time.Duration `yaml:"task-ttl,omitempty"`                  // sec
	TryOneSegment          bool          `yaml:"one-segment,omitempty"`
	MethodHTTP             string        `yaml:"http-method,omitempty"` // GET, HEAD
	ListenHTTP             string        `yaml:"http-api-listen,omitempty"`
	ErrorLog               string        `yaml:"error-log,omitempty"`
	Zabbix                 Zabbix        `yaml:"zabbix,omitempty"`
	ParseName              string        `yaml:"parse-name,omitempty"` // regexp for alternative method of title/name parsing from the URL
	User                   string        `yaml:"user,omitempty"`
	Pass                   string        `yaml:"pass,omitempty"`
}

type Zabbix struct {
	DiscoveryPath   string   `yaml:"discovery-path,omitempty"`
	DiscoveryGroups []string `yaml:"discovery-groups,omitempty"`
	StreamTemplate  string   `yaml:"stream-template,omitempty"`
}

// parsed grup config
type ConfigGroup struct {
	Name                   string
	Type                   StreamType
	Probers                int
	MediaProbers           int
	CheckBrokenTime        int
	ConnectTimeout         time.Duration
	RWTimeout              time.Duration
	SlowWarningTimeout     time.Duration
	VerySlowWarningTimeout time.Duration
	TimeBetweenTasks       time.Duration
	TaskTTL                time.Duration
	TryOneSegment          bool
	MethodHTTP             string
	ParseMethod            string
	User                   string
	Pass                   string
}

type HTTPConfig struct {
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
}

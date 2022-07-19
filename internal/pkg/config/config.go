package config

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/hotid/streamsurfer/internal/pkg/helpers"
	. "github.com/hotid/streamsurfer/internal/pkg/structures"
	"io/ioutil"
	"launchpad.net/goyaml"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func InitAnotherConfig(confile string) *Config {
	config := new(Config)
	rawcfg := rawConfig(confile, config)
	config.IsReady = make(chan bool, 1)
	parseOptionsConfig(rawcfg, config)
	parseGroupsConfig(rawcfg, config)
	config.IsReady <- true
	return config
}

// raw config data
type configYAML struct {
	ListenHTTP       string                     `yaml:"http-api-listen,omitempty"`
	User             string                     `yaml:"http-api-user,omitempty"`
	Pass             string                     `yaml:"http-api-pass,omitempty"`
	Stubs            ConfigStub                 `yaml:"stubs,omitempty"`
	Zabbix           ConfigZabbix               `yaml:"zabbix,omitempty"`
	Samples          []string                   `yaml:"unmortal,omitempty"`
	UserAgents       []string                   `yaml:"user-agents,omitempty"`
	Defaults         configGroupYAML            `yaml:"defaults,omitempty"`
	Groups           map[string]configGroupYAML `yaml:"groups,omitempty"`
	ExpireDurationDB time.Duration              `yaml:"db-expired"` // measured in hours
}

// rawconfig group data
type configGroupYAML struct {
	Type                   string        `yaml:"type,omitempty"`
	URI                    string        `yaml:"streams-uri,omitempty"`               // external link list
	Streams                []string      `yaml:"streams,omitempty"`                   // link list
	Probers                int           `yaml:"probers,omitempty"`                   // num of
	MediaProbers           int           `yaml:"media-probers,omitempty"`             // num of
	CheckBrokenTime        int           `yaml:"check-broken-time"`                   // ms
	ConnectTimeout         time.Duration `yaml:"connect-timeout,omitempty"`           // sec
	RWTimeout              time.Duration `yaml:"rw-timeout,omitempty"`                // sec
	SlowWarningTimeout     time.Duration `yaml:"slow-warning-timeout,omitempty"`      // sec
	VerySlowWarningTimeout time.Duration `yaml:"very-slow-warning-timeout,omitempty"` // sec
	TimeBetweenTasks       time.Duration `yaml:"time-between-tasks,omitempty"`        // sec
	TaskTTL                time.Duration `yaml:"task-ttl,omitempty"`                  // sec
	TryOneSegment          bool          `yaml:"one-segment,omitempty"`
	MethodHTTP             string        `yaml:"http-method,omitempty"` // GET, HEAD
	ErrorLog               string        `yaml:"error-log,omitempty"`
	ParseMethod            string        `yaml:"parse-method,omitempty"` // regexp for alternative method of title/name parsing from the URL
	User                   string        `yaml:"user,omitempty"`
	Pass                   string        `yaml:"pass,omitempty"`
}

// Read raw config with YAML validation
func rawConfig(confile string, config *Config) *configYAML {
	cfg := new(configYAML)

	// Hardcoded defaults:
	config.Stubs = ConfigStub{Name: "Stream Surfer"}

	data, err := ioutil.ReadFile(helpers.FullPath(confile))
	if err == nil {
		err = goyaml.Unmarshal(data, &cfg)
		if err != nil {
			print("Config file parsing failed. Hardcoded defaults used.\n")
		}
	} else {
		panic(err)
	}

	return cfg
}

//
func parseOptionsConfig(rawconfig *configYAML, config *Config) {
	config.ListenHTTP = rawconfig.ListenHTTP
	config.User = rawconfig.User
	config.Pass = rawconfig.Pass
	config.Stubs = rawconfig.Stubs
	config.Zabbix = rawconfig.Zabbix
	config.Samples = rawconfig.Samples
	config.UserAgents = rawconfig.UserAgents
	config.ExpireDurationDB = rawconfig.ExpireDurationDB * time.Hour
}

//
func parseGroupsConfig(rawconfig *configYAML, config *Config) {
	config.GroupParams = make(map[Key]*ConfigGroup)
	config.GroupStreams = make(map[Key]map[Key]Stream)

	for groupName, groupData := range rawconfig.Groups {
		key := sha256.Sum256([]byte(groupName))
		stype := String2StreamType(groupData.Type)

		config.GroupParams[key] = &ConfigGroup{
			Name:                   groupName,
			Type:                   stype,
			Probers:                groupData.Probers,
			MediaProbers:           groupData.MediaProbers,
			CheckBrokenTime:        groupData.CheckBrokenTime,
			ParseMethod:            groupData.ParseMethod,
			TimeBetweenTasks:       groupData.TimeBetweenTasks,
			ConnectTimeout:         groupData.ConnectTimeout,
			RWTimeout:              groupData.RWTimeout,
			SlowWarningTimeout:     groupData.SlowWarningTimeout,
			VerySlowWarningTimeout: groupData.VerySlowWarningTimeout,
			TaskTTL:                groupData.TaskTTL,
			TryOneSegment:          groupData.TryOneSegment,
			MethodHTTP:             strings.ToUpper(groupData.MethodHTTP),
			User:                   groupData.User,
			Pass:                   groupData.Pass,
		}

		if groupData.URI != "" {
			config.GroupStreams[key] = make(map[Key]Stream)
			addRemoteConfig(config.GroupStreams[key], config.GroupParams[key], groupName, groupData.URI, groupData.User, groupData.Pass)
		} else {
			config.GroupStreams[key] = make(map[Key]Stream)
			addLocalConfig(config.GroupStreams[key], config.GroupParams[key], groupName, groupData.Streams)
		}
	}
	// println("----------------------------------------")
	// for key, data := range config.GroupParams {
	// 	fmt.Printf("%s %#v\n", key, data)
	// }
	// println("----------------------------------------")
	// for key, data := range config.GroupStreams {
	// 	fmt.Printf("%s %#v\n", key, data)
	// }
}

func addLocalConfig(dest map[Key]Stream, params *ConfigGroup, group string, sources []string) {
	for _, source := range sources {
		uri, name, title := splitName(params.ParseMethod, source)
		key := sha256.Sum256([]byte(uri))
		dest[key] = Stream{StreamKey: key, URI: uri, Type: params.Type, Name: name, Title: title, Group: group}
	}
}

func addRemoteConfig(dest map[Key]Stream, params *ConfigGroup, group string, uri, remoteUser, remotePass string) error {
	defer func() error {
		if r := recover(); r != nil {
			return errors.New(fmt.Sprintf("Can't get remote config for (%s) %s %s", params.Type, group, uri))
		}
		return nil
	}()

	client := helpers.NewTimeoutClient(10*time.Second, 10*time.Second)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return err
	}
	if remoteUser != "" {
		req.SetBasicAuth(remoteUser, remotePass)
	}
	result, err := client.Do(req)
	if err == nil {
		body := bufio.NewReader(result.Body)
		for {
			line, err := body.ReadString('\n')
			if err != nil {
				break
			}
			uri, name, title := splitName(params.ParseMethod, line)
			key := sha256.Sum256([]byte(uri))
			dest[key] = Stream{StreamKey: key, URI: uri, Type: params.Type, Name: name, Title: title, Group: group}
		}
	}
	return err
}

func String2StreamType(s string) StreamType {
	switch strings.ToLower(s) {
	case "sample":
		return SAMPLE
	case "hls":
		return HLS
	case "hds":
		return HDS
	case "wv":
		return WV
	case "http":
		return HTTP
	default:
		return UNKSTREAM
	}
}

// Helper. Split stream link to URI and Name parts.
// Supported both cases: title<space>uri and uri<space>title
// If `re` presents then name parsed from uri by regular expression.
// URI must be prepended by http:// or https://
func splitName(re, source string) (uri, name, title string) {
	source = strings.TrimSpace(source)
	sep := regexp.MustCompile("htt(p|ps)://")
	loc := sep.FindStringIndex(source)
	if loc != nil {
		if loc[0] == 0 { // uri title
			splitted := strings.SplitN(source, " ", 2)
			if len(splitted) > 1 {
				title = strings.TrimSpace(splitted[1])
			}
			uri = strings.TrimSpace(splitted[0])
		} else { // title uri
			title = strings.TrimSpace(source[0:loc[0]])
			uri = source[loc[0]:]
		}
		if title == "" {
			title = uri
		}
	}
	if re != "" { // get name by regexp
		compiledRe := regexp.MustCompile(re)
		vals := compiledRe.FindStringSubmatch(uri)
		if len(vals) > 1 {
			name = vals[1]
		} else {
			name = title
		}
	} else {
		name = title
	}
	return
}

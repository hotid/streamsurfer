package structures

import (
	"bytes"
	"fmt"
	"github.com/grafov/m3u8"
	"net/http"
	"time"
)

// Data structures related to whole program

const (
	SERVER = "Stream Surfer"
)

// Kinds of streams
// Must be in consistence with StreamType2String() and String2StreamType()
const (
	UNKSTREAM StreamType = iota // хрень какая-то
	SAMPLE                      // internet resources for monitor self checks
	HTTP                        // "plain" HTTP
	HLS                         // Apple HTTP Live Streaming
	HDS                         // Adobe HTTP Dynamic Streaming
	WV                          // Widevine VOD
)

const (
	INFO Severity = iota
	WARNING
	ERROR
	CRITICAL
)

// Error codes (ordered by errors importance).
// If several errors detected then only one with the heaviest weight reported.
// Must be in consistence with String2StreamErr() and StreamErr2String()
const (
	SUCCESS        ErrType = iota
	DEBUG_LEVEL            // Internal debug messages follow below:
	TTLEXPIRED             // Task was not executed because task TTL expired. StreamSurfer too busy.
	HLSPARSER              // HLS parser error (debug)
	BADREQUEST             // Request failed (internal client error)
	WARNING_LEVEL          // Warnings follow below:
	SLOW                   // SlowWarning threshold on reading server response
	VERYSLOW               // VerySlowWarning threshold on reading server response
	ERROR_LEVEL            // Errors follow below:
	CTIMEOUT               // Timeout on connect
	RTIMEOUT               // Timeout on read
	BADLENGTH              // ContentLength value not equal real content length
	BODYREAD               // Response body read error
	CRITICAL_LEVEL         // Permanent errors level
	REFUSED                // Connection refused
	BADSTATUS              // HTTP Status >= 400
	BADURI                 // Incorret URI format
	LISTEMPTY              // HLS specific (by m3u8 lib)
	BADFORMAT              // HLS specific (by m3u8 lib)
	UNKERR                 // хрень какая-то
)

// Commands for probers.
const (
	STOP_MON Command = iota
	START_MON
	RELOAD_CONFIG // TODO made dynamic stream loading and unloading
	LOAD_GROUP
	LOAD_STREAM
	DROP_GROUP
	DROP_STREAM
)

type Severity uint
type StreamType uint // Type of checked streams
type ErrType uint
type Command uint

type Group struct {
	Type StreamType
	Name string
}

type Stream struct {
	StreamKey Key
	URI       string
	Type      StreamType
	Name      string
	Title     string // опциональный заголовок = name по умолчанию
	Group     string
}

// Stream checking task
type Task struct {
	Stream
	ReadBody bool
	ReplyTo  chan *Result
	TTL      time.Time // valid until the time
	Tid      int64     // task id (unique for each stream box)
}

type VariantTask struct {
	Task
}

type ChunkTask struct {
	Task
}

// Stream group
type GroupTask struct {
	Type    StreamType
	Name    string
	Tasks   *Task
	ReplyTo chan Result
}

// Result of task of check streaming
type Result struct {
	Task              *Task
	ErrType           ErrType
	HTTPCode          int    // HTTP status code
	HTTPStatus        string // HTTP status string
	ContentLength     int64
	RealContentLength int64
	Headers           http.Header
	Body              bytes.Buffer
	Started           time.Time     // начало исполнения проверки
	Elapsed           time.Duration // понадобилось времени на задачу
	TotalErrs         uint
	//Meta              interface{} // Reference to metainformation about result data (playlist type etc.)
	Pid        *Result   // link to parent check (is nil for top level URLs)
	SubResults []*Result // Результаты вложенных проверок (i.e. media playlists for different bitrate of master playlists)
}

// Results persistently keeped in Redis
type KeepedResult struct {
	Tid int64 // all subresults have same task id
	Stream
	Master            bool // is master result?
	ErrType           ErrType
	HTTPCode          int    // HTTP status code
	HTTPStatus        string // HTTP status string
	ContentLength     int64
	RealContentLength int64
	Headers           http.Header
	Body              []byte
	Started           time.Time     // начало исполнения проверки
	Elapsed           time.Duration // понадобилось времени на задачу
	TotalErrs         uint
}

// StreamBox statistics
type Stats struct {
	Checks       int64     // checks made
	Errors       int64     // total errors found
	Errors6min   int       // errors in last 3 min
	Errors6hours int       //
	LastCheck    time.Time // last check was at the time
	NextCheck    time.Time // next check will be at this time
	// TODO результаты анализа потока в streambox (анализ перезапускать регулярно по таймеру)
	// TODO вынести инфу о потоке потом в отдельную структуру
}

type MetaHLS struct {
	ListType  m3u8.ListType // type of analyzed playlist
	DeepLinks []string      // sublists for analysis
}

type MetaHDS struct {
	ListType  m3u8.ListType // XXX type of analyzed playlist
	DeepLinks []string      // sublists for analysis
}

// ключ для статистики
type Key [32]byte

func (k *Key) String() string {
	return fmt.Sprintf("%x", k[:])
}

// запросppы на сохранение статистики
type StatInQuery struct {
	Stream Stream
	Last   Stats
}

// запросы на получение статистики
type StatOutQuery struct {
	Key     Key
	ReplyTo chan Stats
}

// запросы на сохранение статистики
type ResultInQuery struct {
	Stream Stream
	Last   Result
}

// запросы на получение статистики
type ResultOutQuery struct {
	Key     Key
	ReplyTo chan []KeepedResult
}

// common type for queries
type OutQuery struct {
	Key     Key
	From    time.Time
	To      time.Time
	ReplyTo chan interface{}
}

type ErrHistoryKey struct {
	Curhour string
	ErrType ErrType
	Group   string
	Name    string
	URI     string
}

type ErrTotalHistoryKey struct {
	Curhour string
	Group   string
	Name    string
}

// Values for the webpage
type PageValues struct {
	Title string
	Data  interface{}
}

// Report about the occured errors
type Report struct {
	Error          ErrType
	Severity       Severity
	Title          string
	Body           string
	RelatedGroups  []*Group
	RelatedStreams []*Stream
	RelatedTasks   []*Task
	Generated      time.Time
}

// Notes last checked point for the stream.
type CheckPoint struct {
	Tid         int64
	Occured     time.Time
	OpenedRange *ErrRange // last non closed error range
}

// Range of errors in the stream. For usage in reports.
type ErrRange struct {
	FromTid      int64
	ToTid        int64
	Occured      time.Time
	Discontinued time.Time
	Err          ErrType
}

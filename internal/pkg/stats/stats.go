package stats

import (
	"errors"
	"github.com/hotid/streamsurfer/internal/pkg/storage"
	. "github.com/hotid/streamsurfer/internal/pkg/structures"
	"sync"
	"time"
)

var StatsGlobals = struct {
	TotalMonitoringPoints     int
	TotalHTTPMonitoringPoints int
	TotalWVMonitoringPoints   int
	TotalHLSMonitoringPoints  int
	TotalHDSMonitoringPoints  int
	MonitoringState           bool // is inet available?
}{}

// Обмен данными с хранителем статистики.
var (
	statIn    chan StatInQuery
	statOut   chan StatOutQuery
	resultIn  chan ResultInQuery
	resultOut chan ResultOutQuery
	errorsOut chan OutQuery
)

// // Streams statistics
// var ReportedStreams = struct {
// 	sync.RWMutex
// 	data map[string]map[string]StreamStats // map [group][stream]stream-state
// }{data: make(map[string]map[string]StreamStats)}

var ErrHistory = struct {
	sync.RWMutex
	count map[ErrHistoryKey]uint
}{count: make(map[ErrHistoryKey]uint)}

var ErrTotalHistory = struct {
	sync.RWMutex
	count map[ErrTotalHistoryKey]uint
}{count: make(map[ErrTotalHistoryKey]uint)}

type Incident struct {
	Id uint64
	Key
	Error *Result
}

// Elder
// Gather results history and statistics from all streamboxes
func StatKeeper(cfg *Config) {
	//var results map[Key][]Result = make(map[Key][]Result)
	var stats map[Key]Stats = make(map[Key]Stats)
	lastCleanUpTime := time.Now()
	//	var errors map[Key]map[time.Time]ErrType = make(map[Key]map[time.Time]ErrType)

	statIn = make(chan StatInQuery, 8192) // receive stats
	statOut = make(chan StatOutQuery, 8)  // send stats
	resultIn = make(chan ResultInQuery, 4096)
	resultOut = make(chan ResultOutQuery, 8)
	errorsOut = make(chan OutQuery, 8)

	// storage maintainance period
	///timer := time.Tick(12 * time.Second)

	for {
		select {
		case state := <-statIn: // receive new statitics data for saving
			stats[state.Stream.StreamKey] = state.Last

		case key := <-statOut:
			if val, ok := stats[key.Key]; ok {
				key.ReplyTo <- val
			} else {
				key.ReplyTo <- Stats{}
			}

		case state := <-resultIn: // incoming results from streamboxes
			//results[Key{state.Stream.Group, state.Stream.Name}] = append(results[Key{state.Stream.Group, state.Stream.Name}], state.Last)
			storage.RedKeepResult(state.Stream.StreamKey, state.Last.Started, state.Last)
			if state.Last.ErrType > WARNING_LEVEL {
				storage.RedKeepError(state.Stream.StreamKey, state.Last.Started, state.Last.ErrType)
			}
			//		delete(errors, Key{state.Stream.Group, state.Stream.Name})

		case key := <-resultOut:
			data, err := storage.RedLoadResults(key.Key, time.Now().Add(-6*time.Hour), time.Now())
			if err != nil {
				key.ReplyTo <- nil
			} else {
				key.ReplyTo <- data
			}

		case key := <-errorsOut: // get error list by streams
			//			result := make(map[time.Time]ErrType)

			data, err := storage.RedLoadErrors(key.Key, key.From, key.To)
			if err != nil {
				key.ReplyTo <- nil
			} else {
				key.ReplyTo <- data
			}

		default: // expired keys cleanup
			if time.Since(lastCleanUpTime) > 30*time.Second {
				storage.RemoveExpiredErrors(cfg.ExpireDurationDB, cfg)
				storage.RemoveExpiredResults(cfg.ExpireDurationDB, cfg)
				lastCleanUpTime = time.Now()
			}
		}
	}
}

// Put result of probe task to statistics.
func SaveStats(stream Stream, last Stats) {
	statIn <- StatInQuery{Stream: stream, Last: last}
}

// Получить состояние по последней проверке.
func LoadStats(key Key) Stats {
	result := make(chan Stats)
	statOut <- StatOutQuery{Key: key, ReplyTo: result}
	data := <-result
	return data
}

// Put result of probe task to statistics.
func SaveResult(stream Stream, last Result) {
	resultIn <- ResultInQuery{Stream: stream, Last: last}
}

// Получить состояние по последней проверке.
func LoadLastResult(key Key) (KeepedResult, error) {
	result := make(chan []KeepedResult)
	resultOut <- ResultOutQuery{Key: key, ReplyTo: result}
	data := <-result
	if data != nil {
		return data[len(data)-1], nil
	} else {
		return KeepedResult{}, errors.New("result not found")
	}
}

// Получить статистику по каналу за всё время наблюдения.
// TODO добавить фильтр по времени (начиная с)
func LoadHistoryResults(key Key) ([]KeepedResult, error) {
	result := make(chan []KeepedResult)
	resultOut <- ResultOutQuery{Key: key, ReplyTo: result}
	data := <-result
	if data != nil {
		return data, nil
	} else {
		return nil, errors.New("result not found")
	}
}

func LoadHistoryErrors(key Key, from time.Duration) (map[time.Time]ErrType, error) {
	result := make(chan interface{})
	errorsOut <- OutQuery{Key: key, From: time.Now().Add(-from), To: time.Now(), ReplyTo: result}
	data := <-result
	if data != nil {
		return data.(map[time.Time]ErrType), nil
	} else {
		return nil, errors.New("result not found")
	}
}

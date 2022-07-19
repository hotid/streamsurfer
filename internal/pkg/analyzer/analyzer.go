package analyzer

import (
	"fmt"
	"github.com/hotid/streamsurfer/internal/pkg/stats"
	. "github.com/hotid/streamsurfer/internal/pkg/structures"
	"time"
)

func ProblemAnalyzer(cfg *Config) {
	var reports []Report
	lastAnalyzed := make(map[Key]CheckPoint)

	for {
		time.Sleep(10 * time.Second)
		for groupKey, groupData := range cfg.GroupParams {
			for streamKey, _ := range cfg.GroupStreams[groupKey] {
				if _, ok := lastAnalyzed[streamKey]; !ok {
					startpoint := time.Now().Add(-2 * time.Hour)
					lastAnalyzed[streamKey] = CheckPoint{0, startpoint, nil}
				}
				checkPoint := lastAnalyzed[streamKey]
				hist, err := stats.LoadHistoryResults(streamKey) // TODO загружать только результаты после времени lastAnalyzed
				if err != nil {
					continue
				}
				switch groupData.Type {
				case HLS:
					reports = analyzeHLS(streamKey, hist, &checkPoint)
				case HDS:
					reports = analyzeHDS(hist)
				case HTTP:
					reports = analyzeHTTP(hist)
				}
				lastAnalyzed[streamKey] = checkPoint
				// TODO save new checkpoint to DB
				if len(reports) > 0 {
					fmt.Printf("reports for %s %v\n", streamKey.String(), reports)
				}
			}
		}
		// TODO RemoveExpiredReports(cfg.ExpireDurationDB)
	}
}

// Report found problems from problem analyzer
// Maybe several methods to report problems of different prirority
func ProblemReporter() {

}

func LoadReports() []Report {
	return nil
}

/*
 * Private realization
 */

// Analyze HLS and form report with error states.
// We check for an each check in an each task in results set. Task interpreted as problem if error occured in
// the any single check. Then we aggregate problem tasks into error ranges ([]ErrRange). Report builder then
// analyze error ranges and make reports on them.
func analyzeHLS(key Key, hist []KeepedResult, lastCheck *CheckPoint) (reports []Report) {
	var (
		isRangeOpened           bool       // problem under analyzator cursor
		isTaskOK                = true     // statuses for current check and task
		start, stop             time.Time  // start and stop timestamps of error period
		prevTid, fromTid, toTid int64      // task id
		errlevel                ErrType    // error level
		errorRanges             []ErrRange // ranges with error states of the stream
		forSave                 *ErrRange  // continious range of failed tasks
	)

	for _, hitem := range hist {
		if hitem.Started.Before(lastCheck.Occured) {
			continue
		}

		if prevTid > 0 && prevTid != hitem.Tid { // переход задач
			if isTaskOK && forSave != nil { // период ошибок кончился
				errorRanges = append(errorRanges, *forSave)
				forSave = nil
				isRangeOpened = false
			} else {
				isTaskOK = true
			}
			prevTid = hitem.Tid
		}

		if prevTid == 0 {
			prevTid = hitem.Tid
		}

		if hitem.ErrType > ERROR_LEVEL {
			isTaskOK = false
			if hitem.ErrType > errlevel {
				errlevel = hitem.ErrType
			}
			if !isRangeOpened {
				isRangeOpened = true
				fromTid = hitem.Tid
				start = hitem.Started
				toTid = fromTid
				stop = start
			} else {
				toTid = hitem.Tid
				stop = hitem.Started.Add(hitem.Elapsed)
			}
			forSave = &ErrRange{fromTid, hitem.Tid, start, stop, errlevel}
		}
	}

	lastCheck = &CheckPoint{toTid, stop, nil}
	if isRangeOpened && errlevel > 0 { // период остался незакрыт
		errorRanges = append(errorRanges, ErrRange{fromTid, toTid, start, stop, errlevel})
	}
	if len(errorRanges) > 0 {
		fmt.Printf("err range for %s %#v\n", key.String(), errorRanges)
	}
	// Permanent errors report. Error is permanent if it continued more than 10 minute.
	for _, val := range errorRanges {
		if val.Discontinued.Sub(val.Occured) > 10*time.Minute {
			reports = append(reports, generatePermanentErrorsReport(key, errorRanges, isRangeOpened))
		}
	}
	return reports
}

func analyzeHDS(hist []KeepedResult) []Report {
	return []Report{}
}

func analyzeHTTP(hist []KeepedResult) []Report {
	return []Report{}
}

/*
 * Report generators
 */

// Permanent errors report generator
func generatePermanentErrorsReport(key Key, ranges []ErrRange, errorPersists bool) Report {
	return Report{Title: "Sample report"}
}

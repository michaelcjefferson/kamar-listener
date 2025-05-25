package main

import (
	"sync"
	"time"

	"github.com/michaelcjefferson/kamar-listener/internal/data"
)

// Mutex to prevent race conditions, in case two routines try to write to appMetrics at the same time
type appMetrics struct {
	lastCheckTime  time.Time
	lastInsertTime time.Time
	mu             sync.RWMutex
	recordsToday   int
	totalRecords   int
	countByType    map[string]int
}

// With mutex active read and return the values for the listener service's last check time, last insert time, records inserted today, and total number of records
func (a *appMetrics) Snapshot() (time.Time, time.Time, int, int, map[string]int) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.lastCheckTime, a.lastInsertTime, a.recordsToday, a.totalRecords, a.countByType
}

func (a *appMetrics) SetLastCheckTime(t time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastCheckTime = t
}

func (a *appMetrics) SetLastInsertTime(t time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastInsertTime = t
}

func (a *appMetrics) SetRecords(today, total int, countByType map[string]int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.recordsToday = today
	a.totalRecords = total
	a.countByType = countByType
}

func (a *appMetrics) IncreaseRecordCount(count int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.recordsToday += count
	a.totalRecords += count
}

// Get the total number of KAMAR records that exist in the KAMAR database, as well as the total number that have been added/updated in the last 24 hours, and update the app metrics numbers to reflect these
func (app *application) UpdateRecordCountsFromDB() error {
	today, total := 0, 0
	countByType := make(map[string]int)

	tod, tot, err := app.models.Assessments.GetAssessmentCount()
	if err != nil {
		app.logger.PrintError(err, map[string]any{
			"message": "error getting assessment count",
		})
		return err
	}
	today += tod
	total += tot
	countByType["assessments"] = tot

	tod, tot, err = app.models.Attendance.GetAttendanceCount()
	if err != nil {
		app.logger.PrintError(err, map[string]any{
			"message": "error getting attendance count",
		})
		return err
	}
	today += tod
	total += tot
	countByType["attendance"] = tot

	tod, tot, err = app.models.Pastoral.GetPastoralCount()
	if err != nil {
		app.logger.PrintError(err, map[string]any{
			"message": "error getting pastoral count",
		})
		return err
	}
	today += tod
	total += tot
	countByType["pastoral"] = tot

	tod, tot, err = app.models.Results.GetResultsCount()
	if err != nil {
		app.logger.PrintError(err, map[string]any{
			"message": "error getting results count",
		})
		return err
	}
	today += tod
	total += tot
	countByType["results"] = tot

	tod, tot, err = app.models.Staff.GetStaffCount()
	if err != nil {
		app.logger.PrintError(err, map[string]any{
			"message": "error getting staff count",
		})
		return err
	}
	today += tod
	total += tot
	countByType["staff"] = tot

	tod, tot, err = app.models.Students.GetStudentsCount()
	if err != nil {
		app.logger.PrintError(err, map[string]any{
			"message": "error getting students count",
		})
		return err
	}
	today += tod
	total += tot
	countByType["students"] = tot

	tod, tot, err = app.models.Subjects.GetSubjectsCount()
	if err != nil {
		app.logger.PrintError(err, map[string]any{
			"message": "error getting students count",
		})
		return err
	}
	today += tod
	total += tot
	countByType["subjects"] = tot

	tod, tot, err = app.models.Timetables.GetTimetablesCount()
	if err != nil {
		app.logger.PrintError(err, map[string]any{
			"message": "error getting students count",
		})
		return err
	}
	today += tod
	total += tot
	countByType["timetables"] = tot

	app.appMetrics.SetRecords(today, total, countByType)

	return nil
}

func (app *application) UpdateCheckInsertTimesFromDB() error {
	var events []data.ListenerEvent

	events, err := app.models.ListenerEvents.GetMostRecentCheckAndInsert()
	if err != nil {
		return err
	}

	for _, e := range events {
		if e.ReqType == "check" {
			app.appMetrics.SetLastCheckTime(e.Time)
		}
		if e.ReqType == "insert" {
			app.appMetrics.SetLastInsertTime(e.Time)
		}
	}

	return nil
}

// Once per hour, call UpdateRecordCountsFromDB, to ensure app metrics are close to accurate
func (app *application) initiateRecordCountUpdateCycle() {
	app.background(func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				err := app.UpdateRecordCountsFromDB()
				if err != nil {
					app.logger.PrintError(err, nil)
				}
			case <-app.isShuttingDown:
				app.logger.PrintInfo("record count update cycle ending - shut down signal received", nil)
				return
			}
		}
	})
}

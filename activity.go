// Package activity provides a simple in memory implementation for recording
// activities and the time spent on each activity via JSON requests. The
// actions and times added only exists in memory and are not persisted.
package activity

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
)

// Work represents the underlying json object sent to AddAction()
type Work struct {
	Action string `json:"action"` // Allowed action types are not defined
	Time   int    `json:"time"`   // Time period type is not defined
}

// Average is the underlying json array object returned via GetStat()
type Average struct {
	Action string `json:"action"` // Action added via AddAction()
	Avg    int    `json:"avg"`    // Average time spent on Action
}

// Used to hold metrics for a single activity
type activityHistory struct {
	totalExecutionTime int
	actionRepeated     int
}

var activitySummary = struct {
	mu sync.RWMutex               // Protects the map
	m  map[string]activityHistory // key = Action, value = activitySummary
}{m: make(map[string]activityHistory)}

// addAction takes the passed in work activity and updates the time spent on the specified Action and
// the # of times the Action has been performed
func (activity *Work) addAction() {
	activitySummary.mu.Lock()
	a, exists := activitySummary.m[activity.Action]
	if exists {
		// Update the current action record=
		a.totalExecutionTime += activity.Time
		a.actionRepeated++
	} else {
		// No activity history for the specified action exists create one
		a = activityHistory{
			totalExecutionTime: activity.Time,
			actionRepeated:     1,
		}
	}
	activitySummary.m[activity.Action] = a
	activitySummary.mu.Unlock()
}

// AddAction this function accepts a json serialized string in the form "{ action: string, time: int}" and
// maintains an average time for each action that can be retrieved using GetStat(). The time period is not defined
// and it's the callers responsiblity to track the type of time period and ensure all entered time periods are
// consistent
func AddAction(jsonActivity string) error {
	bs := []byte(jsonActivity)
	if !json.Valid(bs) {
		return fmt.Errorf(`Invalid JSON request %v must be in the form { action: string, time: int}`, jsonActivity)
	}

	var activity Work
	err := json.Unmarshal(bs, &activity)
	if err != nil {
		log.Println(jsonActivity, err)
		return fmt.Errorf("JSON Unmarshal error on %v: %v", jsonActivity, err)
	}
	activity.addAction()
	return nil
}

// getStatus calculates the average stats for each action that has been recorded by AddAction
func getStats() []Average {
	activitySummary.mu.RLock()
	stats := make([]Average, 0, len(activitySummary.m))
	for action, activityHistory := range activitySummary.m {
		// Per requested API round to int
		average := int(math.Round(float64(activityHistory.totalExecutionTime) / float64(activityHistory.actionRepeated)))
		activityAvg := Average{
			Action: action,
			Avg:    average,
		}
		stats = append(stats, activityAvg)
	}
	activitySummary.mu.RUnlock()
	return stats
}

// GetStats returns a serialized json array of the action and the average time for each action that has been
// provided to the addAction function.
func GetStats() string {
	stats := getStats()

	// Marshall the data to return to callee
	bs, err := json.Marshal(stats)
	if err != nil {
		// This shouldn't ever happen.
		log.Printf("JSON Marshal error on %v: %v", stats, err)
		return ""
	}
	return string(bs)
}

// clearStats clears all saved activity that has been recorded by addAction
func clearStats() {
	activitySummary.mu.Lock()
	activitySummary.m = make(map[string]activityHistory)
	activitySummary.mu.Unlock()
}

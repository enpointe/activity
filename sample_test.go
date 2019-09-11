package activity_test

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/enpointe/activity"
)

func Example() {

	// record  the specified action and time spent on action
	record := func(action string, time int) error {
		// Create a work record
		record := activity.Work{
			Action: action,
			Time:   time}

		// Send the data as a JSON object to AddAction
		bs, err := json.Marshal(record)
		if err != nil {
			return err
		}
		return activity.AddAction(string(bs))
	}

	// retrieve fetch the average of the recorded stats
	retrieve := func() ([]activity.Average, error) {
		// Retrieve the JSON stats via GetStats
		data := activity.GetStats()

		// Convert the JSON object to a struct
		bs := []byte(data)
		var actionStats []activity.Average
		err := json.Unmarshal(bs, &actionStats)
		return actionStats, err
	}

	// Optional: Clear any cached statistics
	activity.ClearStats()

	// Record activity
	record("jump", 100)
	record("run", 75)
	record("jump", 200)

	// Retrieve average of recorded activities
	actionStats, err := retrieve()
	if err != nil {
		log.Fatalf("Failed to retrieve stats: %v", err)
		return
	}

	// Output the average for each recorded activity
	for _, stat := range actionStats {
		fmt.Printf("activity: %v  avg: %v\n", stat.Action, stat.Avg)
	}
	// Output:
	// activity: jump  avg: 150
	// activity: run  avg: 75
}

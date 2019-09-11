package activity_test

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/enpointe/activity"
)

// testRecord add the specified action and time spent on action to
// recorded actions
func testRecord(t *testing.T, action string, time int) {
	record := activity.Work{
		Action: action,
		Time:   time}

	// Send the data as a JSON object to AddAction
	bs, err := json.Marshal(record)
	if err != nil {
		t.Errorf("JSON Marshal failed: %v", err)
	}
	err = activity.AddAction(string(bs))
	if err != nil {
		t.Error(err)
	}
}

// retrieve fetch the average of the recorded stats
func testRetrieve(t *testing.T) (string, []activity.Average) {
	// Retrieve the JSON stats via GetStats
	data := activity.GetStats()
	// Convert the JSON object to a struct
	bs := []byte(data)
	var actionStats []activity.Average
	err := json.Unmarshal(bs, &actionStats)
	if err != nil {
		t.Errorf("Failed to unmarshal GetStats data %v: %v", data, err)
	}
	return data, actionStats
}

// Test to ensure clearStats zeros out all captured activites
func TestClearStats(t *testing.T) {
	activity.ClearStats()
	data, actionStats := testRetrieve(t)
	if len(actionStats) != 0 {
		t.Errorf("ClearStats() did not clear stats: %v", data)
	}
}

// Test sending a empty Json string to ensure error is generated
func TestEmptyAction(t *testing.T) {
	t.Parallel()
	var emptyJSONString string
	err := activity.AddAction(emptyJSONString)
	if err == nil {
		t.Error("Expected error to be generated for invalid json input")
	}
}

// Test addAction when invalid JSON string passed to ensure error is generated
func TestInvalidJSONStr(t *testing.T) {
	t.Parallel()
	err := activity.AddAction("Not a JSON string")
	if err == nil {
		t.Error(err)
	}
}

// Test addAction to ensure that an valid JSON string, but incorrect for our
// implementation generates and error
func TestMalformedJSONInput(t *testing.T) {
	t.Parallel()
	malformed := string([]byte(`
    {
        "somedata": "jump",
    }
	`))
	err := activity.AddAction(malformed)
	if err == nil {
		t.Error("Excepted failure for malformed JSON data")
	}
}

// Test addAction using expected marshalled data form
func TestSingleDataAction(t *testing.T) {
	t.Parallel()
	jsonActivity := string([]byte(`
    {
        "action": "run",
        "time": 100
    }
	`))
	err := activity.AddAction(jsonActivity)
	if err != nil {
		t.Errorf("Unexpected failure: %v", err)
	}
}

// Perform a raw json transfer
func TestSimpleAction(t *testing.T) {
	t.Parallel()
	testRecord(t, "swim", 1000)
}

// Test that average for action jump is computed correctly when returned via GetStats()
func TestGetStatsAvg(t *testing.T) {
	const Jump = "jump"
	activity.ClearStats()
	testRecord(t, Jump, 100)
	testRecord(t, Jump, 25)
	testRecord(t, Jump, 25)
	testRecord(t, Jump, 100)
	testRecord(t, Jump, 100)
	testRecord(t, Jump, 75)
	testRecord(t, Jump, 25)
	expectedAverage := 64
	data, actionStats := testRetrieve(t)
	if len(actionStats) > 1 {
		t.Errorf("Expected 1 statistic got %v: %v", len(actionStats), data)
	}

	// Check the resulting stats to ensure it matches the expected average value
	if Jump != actionStats[0].Action {
		t.Errorf("Expected action %v got %v", Jump, actionStats[0].Action)
	}
	if expectedAverage != actionStats[0].Avg {
		t.Errorf("Expected Average of %v got %v", expectedAverage, actionStats[0].Avg)
	}
}

// Test multiple concurrent actions to test for deadlock condition
func TestMultipleConcurrentActions(t *testing.T) {
	var wg sync.WaitGroup
	const operations = 100
	wg.Add(operations)
	activity.ClearStats()
	for i := 0; i < operations/2; i++ {
		go func() {
			defer wg.Done()
			testRetrieve(t)
			testRecord(t, "run", 100)
			testRecord(t, "walk", 10)
			testRecord(t, "run", 25)
		}()
		go func() {
			defer wg.Done()
			testRecord(t, "walk", 300)
			testRecord(t, "run", 300)
			testRecord(t, "walk", 10)
			testRetrieve(t)
		}()
	}
	wg.Wait()
}

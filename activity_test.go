package activity

import (
	"encoding/json"
	"sync"
	"testing"
)

// Helper function that retieves the stats data and returns
// both the umarshalled and resulting []Averag data
func testGetStats(t *testing.T) (string, []Average) {
	data := GetStats()
	bs := []byte(data)
	// Unmarshall the results
	var actionStats []Average
	err := json.Unmarshal(bs, &actionStats)
	if err != nil {
		t.Errorf("Failed to unmarshal GetStats data %v: %v", data, err)
	}
	return data, actionStats
}

// Test to ensure clearStats zeros out all captured activites
func TestClearStats(t *testing.T) {
	clearStats()
	data, actionStats := testGetStats(t)
	if len(actionStats) != 0 {
		t.Errorf("clearStats() did not clear stats: %v", data)
	}
}

// Test sending a empty Json string to ensure error is generated
func TestEmptyAction(t *testing.T) {
	t.Parallel()
	var emptyJSONString string
	err := AddAction(emptyJSONString)
	if err == nil {
		t.Error("Expected error to be generated for invalid json input")
	}
}

// Test addAction when invalid JSON string passed to ensure error is generated
func TestInvalidJSONStr(t *testing.T) {
	t.Parallel()
	err := AddAction("Not a JSON string")
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
	err := AddAction(malformed)
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
	err := AddAction(jsonActivity)
	if err != nil {
		t.Errorf("Unexpected failure: %v", err)
	}
}

// Perform a raw json transfer
func TestSimpleAction(t *testing.T) {
	t.Parallel()
	workout := Work{
		Action: "swim",
		Time:   1000}

	bs, err := json.Marshal(workout)
	if err != nil {
		// This shouldn't ever happen.
		t.Errorf("Unexpected JSON Marshal error on %v: %v", workout, err)
	}
	err = AddAction(string(bs))
	if err != nil {
		t.Errorf("Failed to add action: %v", err)
	}
}

// Test that average for action jump is computed correctly when returned via GetStats()
func TestGetStatsAvg(t *testing.T) {
	const Jump = "jump"
	jumps := []Work{
		Work{Jump, 100},
		Work{Jump, 25},
		Work{Jump, 25},
		Work{Jump, 100},
		Work{Jump, 100},
		Work{Jump, 75},
		Work{Jump, 25},
	}
	// Expected average 64
	expectedAverage := 64
	clearStats()
	for _, activity := range jumps {
		activity.addAction()
	}
	data, actionStats := testGetStats(t)
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
	jumpAction := string([]byte(`
    {
        "action": "run",
        "time": 100
    }
	`))
	walkAction := string([]byte(`
    {
        "action": "walk",
        "time": 100
    }
	`))
	clearStats()
	AddAction(jumpAction)
	for i := 0; i < operations/2; i++ {
		go func() {
			defer wg.Done()
			stats := GetStats()
			if len(stats) == 0 {
				t.Errorf("Failed to retrieve any stats")
			}
			err := AddAction(jumpAction)
			if err != nil {
				t.Error(err)
			}
			AddAction(walkAction)
			if err != nil {
				t.Error(err)
			}
		}()
		go func() {
			defer wg.Done()
			err := AddAction(jumpAction)
			if err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}

package activity

import (
	"encoding/json"
	"fmt"
	"log"
)

func ExampleAddAction() {
	activity := Activity{
		Action: "swim",
		Time:   300}

	bs, err := json.Marshal(activity)
	if err != nil {
		log.Fatal(err)
	}
	err = AddAction(string(bs))
	if err != nil {
		log.Fatal(err)
	}
	err = AddAction(`{"action":"swim", "time":170}`)
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleGetStats() {
	data := GetStats()
	bs := []byte(data)
	var actionStats []Average
	err := json.Unmarshal(bs, &actionStats)
	if err != nil {
		log.Fatal(err)
	}
	for _, stat := range actionStats {
		fmt.Printf("activity: %v\tavg: %v\n", stat.Action, stat.Avg)
	}
}

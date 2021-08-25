package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/apognu/gocal"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Requires 3 arguments:\n\t path-to-primary-ics path-to-secondary-ics start-date(yyyy-mm-dd)")
		return
	}

	primaryFilePath := os.Args[1]
	secondaryFilePath := os.Args[2]

	start, err := time.Parse("2006-01-02", os.Args[3])
	if err != nil {
		fmt.Println("error: " + err.Error())
		return
	}

	end := start.AddDate(0, 7, 0)

	primaryOnCallWeeks, primaryKeys := parseOnCallSchedule(start, end, primaryFilePath)
	secondaryOnCallWeeks, _ := parseOnCallSchedule(start, end, secondaryFilePath)

	fmt.Println("Week,Primary NASA,Primary EMEA,Secondary NASA,Secondary EMEA")

	for _, k := range primaryKeys {
		peopleOnCallPrimary := primaryOnCallWeeks[k]
		peopleOnCallSecondary := secondaryOnCallWeeks[k]
		rowStr := fmt.Sprintf("%v,%v,%v", weekStart(k).Format("2006-01-02"), prettyPrintArray(peopleOnCallPrimary), prettyPrintArray(peopleOnCallSecondary))
		// For some reason OpsGenie only picks up Joram's first name
		if !strings.Contains(rowStr, "Joram Wilander") {
			rowStr = strings.Replace(rowStr, "Joram", "Joram Wilander", -1)
		}
		fmt.Println(rowStr)
	}
}

func prettyPrintArray(array []string) string {
	if len(array) == 0 {
		return ","
	}

	finalStr := array[0]
	for i, str := range array {
		if i == 0 {
			continue
		}
		finalStr += "," + str
	}

	return finalStr
}

func parseOnCallSchedule(start, end time.Time, filePath string) (map[string][]string, []string) {
	f, _ := os.Open(filePath)
	defer f.Close()

	c := gocal.NewParser(f)
	c.Start, c.End = &start, &end
	c.Parse()

	onCallWeeks := map[string][]string{}

	for _, e := range c.Events {
		year, week := e.Start.ISOWeek()
		key := fmt.Sprintf("%v-%v", year, week)

		person := personFromEventSummary(e.Summary)

		peopleOnCall := onCallWeeks[key]
		if peopleOnCall == nil {
			peopleOnCall = []string{person}
		} else {
			found := false
			for _, p := range peopleOnCall {
				if p == person {
					found = true
					break
				}
			}
			if !found {
				peopleOnCall = append(peopleOnCall, person)
			}
		}

		onCallWeeks[key] = peopleOnCall
	}

	keys := make([]string, len(onCallWeeks))
	i := 0
	for k := range onCallWeeks {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	return onCallWeeks, keys
}

func personFromEventSummary(summary string) string {
	return strings.Split(summary, " (user)")[0]
}

func weekStart(yearWeek string) time.Time {
	splitStr := strings.Split(yearWeek, "-")
	year, _ := strconv.ParseInt(splitStr[0], 10, 64)
	week, _ := strconv.ParseInt(splitStr[1], 10, 64)

	// Start from the middle of the year:
	t := time.Date(int(year), 7, 1, 0, 0, 0, 0, time.UTC)

	// Roll back to Monday:
	if wd := t.Weekday(); wd == time.Sunday {
		t = t.AddDate(0, 0, -6)
	} else {
		t = t.AddDate(0, 0, -int(wd)+1)
	}

	// Difference in weeks:
	_, w := t.ISOWeek()
	t = t.AddDate(0, 0, (int(week)-w)*7)

	return t
}

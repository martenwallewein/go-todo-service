package markdown

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type TimeTrackingList struct {
	Months []*TimeTrackingMonth
}

type TimeTrackingMonth struct {
	Items []*TimeTrackingItem
	Date  time.Time
}

type TimeTrackingItem struct {
	InProgress bool
	Task       string
	Start      time.Time
	End        time.Time
}

func (tl *TimeTrackingList) GetCurrentMonth() *TimeTrackingMonth {
	today := time.Now()
	for _, m := range tl.Months {
		if today.Month() == m.Date.Month() {
			return m
		}
	}

	return nil
}

func (tm *TimeTrackingMonth) GetTodaysTasks() []*TimeTrackingItem {
	todaysTasks := make([]*TimeTrackingItem, 0)
	now := time.Now()
	for _, item := range tm.Items {
		if DayEqual(item.Start, now) {
			todaysTasks = append(todaysTasks, item)
		}
	}

	return todaysTasks
}

func (tm *TimeTrackingMonth) StartTodayTask(task string) {
	todaysTasks := tm.GetTodaysTasks()
	newItem := &TimeTrackingItem{
		Task:       task,
		Start:      time.Now(),
		InProgress: true,
	}
	if len(todaysTasks) > 0 {
		index := 0
		target := todaysTasks[len(todaysTasks)-1]
		for i, v := range tm.Items {
			if target.Task == v.Task && DayEqual(v.Start, target.Start) {
				index = i
				break
			}
		}
		logrus.Info("Inserting at index ", index+1)
		tm.Items = InsertIntoSliceAtIndex(tm.Items, newItem, index+1)
	} else {
		tm.Items = append(tm.Items, newItem)
	}

}

func (tm *TimeTrackingMonth) CompleteTodayTask(task string) bool {
	today := time.Now()
	for _, item := range tm.Items {
		if !DayEqual(item.Start, today) {
			continue
		}

		logrus.Warn("Checking ", item.Task, " to match ", task, " with result ", strings.Index(item.Task, task))
		if strings.Index(item.Task, task) >= 0 {
			item.InProgress = false
			item.End = time.Now()
			// Since these are not pointers we need to update the element
			// tm.Items[index] = item
			return false
		}
	}
	return true
}

func (tl *TimeTrackingList) WriteToFile(file string) error {
	str := ""
	curDay := 0
	for _, month := range tl.Months {
		// Write month line
		str += fmt.Sprintf("## %s/%d\n", appendZeroIfMissing(int(month.Date.Month())), month.Date.Year())
		// write -TimeTrackings
		str += "- times:\n"
		// Render all TimeTrackings
		for _, item := range month.Items {
			if item.Start.Day() != curDay { // Add new day
				str += fmt.Sprintf("    - %s.%s:\n", appendZeroIfMissing(item.Start.Day()), appendZeroIfMissing(int(item.Start.Month())))
				curDay = item.Start.Day()
			}

			if item.InProgress {
				str += fmt.Sprintf("        - [%s:%s-] %s\n", appendZeroIfMissing(item.Start.Hour()), appendZeroIfMissing(item.Start.Minute()), item.Task)
			} else {
				str += fmt.Sprintf("        - [%s:%s-%s:%s] %s\n", appendZeroIfMissing(item.Start.Hour()), appendZeroIfMissing(item.Start.Minute()), appendZeroIfMissing(item.End.Hour()), appendZeroIfMissing(item.End.Minute()), item.Task)
			}
		}
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}

	_, err = f.WriteString(str)
	if err != nil {
		return err
	}

	return nil
}

func ParseTimeTrackingMarkdown(file string) (*TimeTrackingList, error) {

	tl := &TimeTrackingList{}
	readFile, err := os.Open(file)

	if err != nil {
		return nil, err
	}
	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)
	// TimeTracking: Year
	curDay := time.Now()
	curYear := time.Now()
	curYearInt := 0
	var timeTrackingMonth *TimeTrackingMonth = nil

	for fileScanner.Scan() {
		line := fileScanner.Text()
		if line == "" {
			continue
		}
		if strings.Index(line, "##") == 0 { // New Month
			if timeTrackingMonth != nil {
				tl.Months = append(tl.Months, timeTrackingMonth)
			}

			re, _ := regexp.Compile("##.*\\s(.*)\\/(.*)")
			matches := re.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				fmt.Printf("key=%s, value=%s\n", match[1], match[2])
				curYearInt, _ = strconv.Atoi(match[2])
				if strings.Index(match[1], "0") == 0 {
					match[1] = strings.ReplaceAll(match[1], "0", "")
				}
				month, _ := strconv.Atoi(match[1])
				curYear = time.Date(curYearInt, time.Month(month), 1, 1, 1, 0, 0, time.Local)
				fmt.Println("Setting new year to ", curYearInt, " and month ", month)
				break
			}

			timeTrackingMonth = &TimeTrackingMonth{
				Items: []*TimeTrackingItem{},
				Date:  curYear,
			}
		}

		if strings.Index(line, "- TimeTrackings") >= 0 { // Add TimeTrackings to month
			continue
		}

		if strings.Index(line, "- [") == -1 { // New date
			re, _ := regexp.Compile(".*\\s(.*)\\.(.*)\\:.*")
			matches := re.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				fmt.Printf("key=%s, value=%s\n", match[1], match[2])
				if strings.Index(match[1], "0") == 0 {
					match[1] = strings.ReplaceAll(match[1], "0", "")
				}
				month, _ := strconv.Atoi(match[2])
				if strings.Index(match[2], "0") == 0 {
					match[2] = strings.ReplaceAll(match[2], "0", "")
				}
				day, _ := strconv.Atoi(match[1])
				curDay = time.Date(2023, time.Month(month), day, 0, 0, 0, 0, time.Local)
				break
			}
		} else {
			/*
				## 01/2023
				- times:
					- 09.01:
						- [09:00-10:30] 8) Kollegefrechdachs meeting
						- [10:30-11:00] Duschen
						- [11:00-12:30] Telegram TimeTracking integration
						- [13:00-14:00] Complete undeployed telegram TimeTracking integration
					- 10.01:
						- [09:00-10:00] 7) Complete Hercules slides
						- [10:00-] 6) Tofino Meeting
			*/
			res := strings.Trim(line, "")
			inProgress := strings.Index(line, "-]") >= 0
			parts := strings.Split(res, "] ")
			times := strings.ReplaceAll(parts[0], "- [", "")
			logrus.Warn(times)
			timeParts := strings.Split(times, "-")
			start, end := startEndTimeFromString(curYearInt, int(timeTrackingMonth.Date.Month()), curDay.Day(), timeParts)

			task := parts[1]
			td := &TimeTrackingItem{
				InProgress: inProgress,
				Task:       task,
				Start:      start,
			}
			if end != nil {
				td.End = *end
			}
			timeTrackingMonth.Items = append(timeTrackingMonth.Items, td)

		}
	}

	tl.Months = append(tl.Months, timeTrackingMonth)

	readFile.Close()
	return tl, nil

}

func startEndTimeFromString(year int, month int, day int, timeParts []string) (time.Time, *time.Time) {
	startParts := strings.Split(strings.Trim(timeParts[0], " "), ":")
	endParts := strings.Split(strings.Trim(timeParts[1], " "), ":")

	logrus.Warn(startParts)
	logrus.Warn(endParts)
	logrus.Warn(timeParts[1])
	startHour, _ := strconv.Atoi(startParts[0])
	startMinute, _ := strconv.Atoi(startParts[1])
	start := time.Date(year, time.Month(month), day, startHour, startMinute, 0, 0, time.Local)
	if len(endParts) < 2 {
		return start, nil
	}
	endHour, _ := strconv.Atoi(endParts[0])
	endMinute, _ := strconv.Atoi(endParts[1])

	end := time.Date(year, time.Month(month), day, endHour, endMinute, 0, 0, time.Local)

	return start, &end
}

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

type TodoList struct {
	Goals  []*TodoItem
	Months []*TodoMonth
}

type TodoMonth struct {
	Goals []*TodoItem
	Items []*TodoItem
	Date  time.Time
}

type TodoItem struct {
	Done       bool
	InProgress bool
	Task       string
	Day        time.Time
}

func (tm *TodoMonth) GetTodaysTasks() []*TodoItem {
	todaysTasks := make([]*TodoItem, 0)
	now := time.Now()
	for _, item := range tm.Items {
		if DayEqual(item.Day, now) {
			todaysTasks = append(todaysTasks, item)
		}
	}

	return todaysTasks
}

func (tm *TodoMonth) GetFullTask(task string) *TodoItem {
	today := time.Now()
	for _, item := range tm.Items {
		if !DayEqual(item.Day, today) {
			continue
		}

		logrus.Warn("Checking ", item.Task, " to match ", task, " with result ", strings.Index(item.Task, task))
		if strings.Index(item.Task, task) >= 0 {
			return item
		}
	}
	return nil
}

func (tm *TodoMonth) StartTodayTask(task string) bool {
	today := time.Now()
	for _, item := range tm.Items {
		if !DayEqual(item.Day, today) {
			continue
		}

		logrus.Warn("Checking ", item.Task, " to match ", task, " with result ", strings.Index(item.Task, task))
		if strings.Index(item.Task, task) >= 0 {
			item.InProgress = true
			// Since these are not pointers we need to update the element
			// tm.Items[index] = item
			return false
		}
	}

	// Just add and complete it
	tm.AddTodayTask(task, false, true)
	return true
}

func (tm *TodoMonth) CompleteTodayTask(task string) bool {
	today := time.Now()
	for _, item := range tm.Items {
		if !DayEqual(item.Day, today) {
			continue
		}

		logrus.Warn("Checking ", item.Task, " to match ", task, " with result ", strings.Index(item.Task, task))
		if strings.Index(item.Task, task) >= 0 {
			item.Done = true
			item.InProgress = false
			// Since these are not pointers we need to update the element
			// tm.Items[index] = item
			return false
		}
	}

	// Just add and complete it
	tm.AddTodayTask(task, true, false)
	return true
}

func DayEqual(v, today time.Time) bool {
	return v.Day() == today.Day() && v.Month() == today.Month() && v.Year() == today.Year()
}

/*func insertTodoItem(list []*TodoItem, c *TodoItem, i int) []*TodoItem {
	if i == len(list)-1 {
		return append(list, c)
	}
	return append(list[:(i+1)], append([]*TodoItem{c}, list[(i+1):]...)...)
}*/

func InsertIntoSliceAtIndex[T any](destination []T, element T, index int) []T {
	if len(destination) == index {
		return append(destination, element)
	}

	destination = append(destination[:index+1], destination[index:]...) // index < len(a)
	destination[index] = element

	return destination
}

func (tm *TodoMonth) AddTodayTask(task string, completed bool, inProgress bool) {

	// TODO: Not numbered tasks

	todaysTasks := tm.GetTodaysTasks()
	num := len(todaysTasks) + 1
	newItem := &TodoItem{
		Done:       completed,
		Task:       fmt.Sprintf("%d) %s", num, task),
		Day:        time.Now(),
		InProgress: inProgress,
	}
	if len(todaysTasks) > 0 {
		index := 0
		target := todaysTasks[len(todaysTasks)-1]
		for i, v := range tm.Items {
			if target.Task == v.Task && DayEqual(v.Day, target.Day) {
				index = i
				break
			}
		}
		logrus.Info("Inserting at index ", index+1)
		tm.Items = InsertIntoSliceAtIndex(tm.Items, newItem, index+1)
	} else {
		tm.Items = append(tm.Items, newItem)
	}

	// tm.Items = append(tm.Items)
}

func (tl *TodoList) GetCurrentMonth() *TodoMonth {
	today := time.Now()
	for _, m := range tl.Months {
		if today.Month() == m.Date.Month() {
			return m
		}
	}

	return nil
}

func appendZeroIfMissing(val int) string {
	str := fmt.Sprintf("%d", val)
	if len(str) == 1 {
		return "0" + str
	}

	return str
}

func (tl *TodoList) WriteToFile(file string) error {
	str := ""
	curDay := 0
	for _, month := range tl.Months {
		// Write month line
		str += fmt.Sprintf("## %s/%d\n", appendZeroIfMissing(int(month.Date.Month())), month.Date.Year())
		// write -goals
		str += "- goals:\n"
		// Render all goals
		for _, goal := range month.Goals {
			if goal.Done {
				str += fmt.Sprintf("    - [x] %s\n", goal.Task)
			} else if goal.InProgress {
				str += fmt.Sprintf("    - [0] %s\n", goal.Task)
			} else {
				str += fmt.Sprintf("    - [ ] %s\n", goal.Task)
			}
		}
		// write -todos
		str += "- todos:\n"
		// Render all todos
		for _, todo := range month.Items {
			if todo.Day.Day() != curDay { // Add new day
				str += fmt.Sprintf("    - %s.%s:\n", appendZeroIfMissing(todo.Day.Day()), appendZeroIfMissing(int(todo.Day.Month())))
				curDay = todo.Day.Day()
			}

			if todo.Done {
				str += fmt.Sprintf("        - [x] %s\n", todo.Task)
			} else if todo.InProgress {
				str += fmt.Sprintf("        - [0] %s\n", todo.Task)
			} else {
				str += fmt.Sprintf("        - [ ] %s\n", todo.Task)
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

func ParseMarkdown(file string) (*TodoList, error) {

	tl := &TodoList{}
	readFile, err := os.Open(file)

	if err != nil {
		return nil, err
	}
	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)
	// TODO: Year
	curDay := time.Now()
	curYear := time.Now()
	curYearInt := 0
	var todoMonth *TodoMonth = nil
	goalsMode := true

	for fileScanner.Scan() {
		line := fileScanner.Text()
		if line == "" {
			continue
		}
		if strings.Index(line, "##") == 0 { // New Month
			if todoMonth != nil {
				tl.Months = append(tl.Months, todoMonth)
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

			todoMonth = &TodoMonth{
				Goals: []*TodoItem{},
				Items: []*TodoItem{},
				Date:  curYear,
			}
		}

		if strings.Index(line, "- goals") >= 0 { // Add goals to month
			goalsMode = true
			continue
		}

		if strings.Index(line, "- todos") >= 0 { // Add todos to month
			goalsMode = false
			fmt.Println("Moving to todos")
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
			res := strings.Trim(line, "")
			done := strings.Index(line, "[x]") >= 0
			inProgress := strings.Index(line, "[0]") >= 0
			res = strings.ReplaceAll(res, "- [ ] ", "")
			res = strings.ReplaceAll(res, "- [x] ", "")
			res = strings.ReplaceAll(res, "- [0] ", "")
			res = strings.Trim(res, " ")
			td := &TodoItem{
				Done:       done,
				InProgress: inProgress,
				Task:       res,
				Day:        curDay,
			}
			logrus.Warn(*td)
			if goalsMode {
				todoMonth.Goals = append(todoMonth.Goals, td)
			} else {
				todoMonth.Items = append(todoMonth.Items, td)
			}
		}
	}

	tl.Months = append(tl.Months, todoMonth)

	readFile.Close()
	return tl, nil

}

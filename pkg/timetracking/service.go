package timetracking

import (
	"fmt"
	"path/filepath"

	"github.com/martenwallewein/todo-service/pkg/git"
	"github.com/martenwallewein/todo-service/pkg/markdown"
	"github.com/sirupsen/logrus"
)

type TimeTrackingService struct {
	repoPath string
}

func NewTimeTrackingService(repoPath string) *TimeTrackingService {
	return &TimeTrackingService{
		repoPath,
	}
}

func (ts *TimeTrackingService) PrepareRepo() (*git.GitRepo, error) {
	logrus.Warn(ts.repoPath)
	repo, err := git.Load(ts.repoPath)
	if err != nil {
		return nil, err
	}

	err = repo.FetchAndRebase()
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (ts *TimeTrackingService) CommitAndPushRepo(repo *git.GitRepo, message string) error {
	err := repo.CommitAll(message)
	if err != nil {
		return err
	}

	err = repo.Push()
	if err != nil {
		return err
	}

	return nil
}

func (ts *TimeTrackingService) LoadTimeTrackingList() (*markdown.TimeTrackingList, error) {
	return markdown.ParseTimeTrackingMarkdown(filepath.Join(ts.repoPath, "timetracking.md"))
}

func (ts *TimeTrackingService) SaveTimeTrackingList(tl *markdown.TimeTrackingList) error {
	return tl.WriteToFile(filepath.Join(ts.repoPath, "timetracking.md"))
}

func (ts *TimeTrackingService) CompleteTodayTimeTracking(task string) error {
	repo, err := ts.PrepareRepo()
	if err != nil {
		return err
	}
	tl, err := ts.LoadTimeTrackingList()
	if err != nil {
		return err
	}

	month := tl.GetCurrentMonth()
	month.CompleteTodayTask(task)

	err = ts.SaveTimeTrackingList(tl)
	if err != nil {
		return err
	}

	err = ts.CommitAndPushRepo(repo, fmt.Sprintf("Complete task %s", task))
	if err != nil {
		return err
	}

	return nil
}

func (ts *TimeTrackingService) StartTodayTimeTracking(task string) error {
	repo, err := ts.PrepareRepo()
	if err != nil {
		return err
	}
	tl, err := ts.LoadTimeTrackingList()
	if err != nil {
		return err
	}

	month := tl.GetCurrentMonth()
	month.StartTodayTask(task)

	err = ts.SaveTimeTrackingList(tl)
	if err != nil {
		return err
	}

	err = ts.CommitAndPushRepo(repo, fmt.Sprintf("Complete task %s", task))
	if err != nil {
		return err
	}

	return nil
}

func (ts *TimeTrackingService) GetTodaysTimeTrackings() ([]*markdown.TimeTrackingItem, error) {

	_, err := ts.PrepareRepo()
	if err != nil {
		return nil, err
	}

	tl, err := ts.LoadTimeTrackingList()
	if err != nil {
		return nil, err
	}

	month := tl.GetCurrentMonth()
	return month.GetTodaysTasks(), nil
}

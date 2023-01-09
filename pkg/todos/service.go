package todos

import (
	"fmt"
	"path/filepath"

	"github.com/martenwallewein/todo-service/pkg/git"
	"github.com/martenwallewein/todo-service/pkg/markdown"
	"github.com/sirupsen/logrus"
)

type TodoService struct {
	repoPath string
}

func NewTodoService(repoPath string) *TodoService {
	return &TodoService{
		repoPath,
	}
}

func (ts *TodoService) PrepareRepo() (*git.GitRepo, error) {
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

func (ts *TodoService) CommitAndPushRepo(repo *git.GitRepo, message string) error {
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

func (ts *TodoService) LoadTodoList() (*markdown.TodoList, error) {
	return markdown.ParseMarkdown(filepath.Join(ts.repoPath, "todos.md"))
}

func (ts *TodoService) SaveTodoList(tl *markdown.TodoList) error {
	return tl.WriteToFile(filepath.Join(ts.repoPath, "todos.md"))
}

func (ts *TodoService) AddTodayTodo(task string) error {

	repo, err := ts.PrepareRepo()
	if err != nil {
		return err
	}
	tl, err := ts.LoadTodoList()
	if err != nil {
		return err
	}

	month := tl.GetCurrentMonth()
	month.AddTodayTask(task, false)

	err = ts.SaveTodoList(tl)
	if err != nil {
		return err
	}

	err = ts.CommitAndPushRepo(repo, fmt.Sprintf("Add task %s to todos", task))
	if err != nil {
		return err
	}

	return nil
}

func (ts *TodoService) CompleteTodayTodo(task string) error {
	repo, err := ts.PrepareRepo()
	if err != nil {
		return err
	}
	tl, err := ts.LoadTodoList()
	if err != nil {
		return err
	}

	month := tl.GetCurrentMonth()
	month.CompleteTodayTask(task)

	err = ts.SaveTodoList(tl)
	if err != nil {
		return err
	}

	err = ts.CommitAndPushRepo(repo, fmt.Sprintf("Complete task %s", task))
	if err != nil {
		return err
	}

	return nil
}

func (ts *TodoService) GetTodaysTodos() ([]*markdown.TodoItem, error) {

	_, err := ts.PrepareRepo()
	if err != nil {
		return nil, err
	}

	tl, err := ts.LoadTodoList()
	if err != nil {
		return nil, err
	}

	month := tl.GetCurrentMonth()
	return month.GetTodaysTasks(), nil
}

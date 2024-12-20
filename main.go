package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/xyzbit/minitaskx/core/model"
)

const (
	host = "http://localhost:8080/"
)

var rootCmd = &cobra.Command{
	Use:   "minitaskx-example",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application.`,
}

var createTaskCmd = &cobra.Command{
	Use:   "create",
	Short: "Run a new task",
	RunE: func(cmd *cobra.Command, args []string) error {
		bizID := fmt.Sprintf("test_flow_run_%d", rand.Int31())
		if err := createTask(bizID); err != nil {
			return err
		}
		fmt.Println("create task success, will be finshed after 30s, bizID: ", bizID)
		return nil
	},
}

var watchTaskCmd = &cobra.Command{
	Use:   "watch",
	Short: "watch task status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("please input bizID")
		}
		bizID := args[0]
		return watchTaskStatus(bizID)
	},
}

var pauseTaskCmd = &cobra.Command{
	Use:   "pause",
	Short: "pause task",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("please input bizID")
		}
		bizID := args[0]
		return operateTask(bizID, model.TaskStatusPaused)
	},
}

var resumeTaskCmd = &cobra.Command{
	Use:   "resume",
	Short: "resume task",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("please input bizID")
		}
		bizID := args[0]
		return operateTask(bizID, model.TaskStatusRunning)
	},
}

var stopTaskCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop task",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("please input bizID")
		}
		bizID := args[0]
		return operateTask(bizID, model.TaskStatusStop)
	},
}

func init() {
	rootCmd.AddCommand(createTaskCmd)
	rootCmd.AddCommand(pauseTaskCmd)
	rootCmd.AddCommand(resumeTaskCmd)
	rootCmd.AddCommand(stopTaskCmd)
	rootCmd.AddCommand(watchTaskCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func createTask(bizID string) error {
	var req bytes.Buffer

	d := fmt.Sprintf(`
		{
			"biz_id": "%s",
			"biz_type": "test",
			"type": "goroutine",
			"payload": "test task run"
		}
	`, bizID)
	req.WriteString(d)
	resp, err := http.Post(host+"api/v1/scheduler/tasks/create", "application/json", &req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		return fmt.Errorf("got %d want %d", status, http.StatusOK)
	}

	return nil
}

func operateTask(bizID string, status model.TaskStatus) error {
	req := bytes.NewBufferString(fmt.Sprintf(`
		{
			"biz_id": "%s",
			"status": "%s"
		}
	`, bizID, status.String()))
	resp, err := http.Post(host+"api/v1/scheduler/tasks/operate", "application/json", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		return fmt.Errorf("got %d want %d", status, http.StatusOK)
	}
	return nil
}

func watchTaskStatus(bizID string) error {
	for {
		time.Sleep(time.Second)

		task, err := listTasks(bizID)
		if err != nil {
			return err
		}

		if task == nil {
			continue
		}
		fmt.Printf("task status: %s\n", task.Status)
	}
}

func listTasks(bizID string) (*model.Task, error) {
	resp, err := http.Get(host + "api/v1/scheduler/tasks/list?biz_ids=" + bizID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		Data []*model.Task `json:"data"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	if len(data.Data) == 0 {
		return nil, nil
	}
	return data.Data[0], nil
}

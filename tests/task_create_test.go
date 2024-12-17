package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/xyzbit/minitaskx/core/model"
)

const (
	host = "http://localhost:8080/"
)

func TestTaskRun(t *testing.T) {
	bizID := fmt.Sprintf("test_flow_run_%d", rand.Int31())
	if err := createTask(bizID); err != nil {
		t.Fatal(err)
	}
	t.Log("create task success and wait for task finshed(<10s)")

	if err := watchTaskStatus(bizID, model.TaskStatusSuccess); err != nil {
		t.Fatal(err)
	}
	t.Log("task had finshed")
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

func watchTaskStatus(bizID string, status model.TaskStatus) error {
	for {
		time.Sleep(time.Second)

		task, err := listTasks(bizID)
		if err != nil {
			return err
		}

		if task == nil {
			continue
		}
		if task.Status == status {
			return nil
		}
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

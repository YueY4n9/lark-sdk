package lark_sdk

import (
	"context"
	"github.com/YueY4n9/gotools/echo"
	"testing"
)

// user-key = 7288177041931763713

func newProjectClient() ProjectClient {
	return NewProjectClient("MII_669A08DA390B4004", "A7A4E3E8F69E11679F1B830CEBBEA12D", "MII_669A08DA390B4004", "A7A4E3E8F69E11679F1B830CEBBEA12D")
}

func TestProjectClient_ListTask(t *testing.T) {
	client := newProjectClient()
	taskList, err := client.ListTask(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	echo.Json(taskList)
}

func TestProjectClient_ListProject(t *testing.T) {
	client := newProjectClient()
	project, err := client.ListProject(context.Background(), "7288177041931763713")
	if err != nil {
		t.Fatal(err)
	}
	echo.Json(project)
}

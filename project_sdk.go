package lark_sdk

import (
	"context"
	projectsdk "github.com/larksuite/project-oapi-sdk-golang"
	"github.com/larksuite/project-oapi-sdk-golang/service/project"
	"github.com/larksuite/project-oapi-sdk-golang/service/task"
	"github.com/larksuite/project-oapi-sdk-golang/service/workitem"
	"net/http"
)

// user-key = 7288177041931763713

type ProjectClient interface {
	Client() *projectsdk.Client

	ListProject(ctx context.Context, userKey string) ([]string, error)
	ListWorkItem(ctx context.Context, projectKey string) ([]*workitem.WorkItemInfo, error)
	ListTask(ctx context.Context) ([]*task.SubDetail, error)
}

type projectClient struct {
	appId       string
	appSecret   string
	debugId     string
	debugSecret string
	client      *projectsdk.Client
}

func NewProjectClient(appId, appSecret string, debug ...string) ProjectClient {
	header := http.Header{}
	header.Add("X-USER-KEY", "7288177041931763713")
	c := &projectClient{
		appId:     appId,
		appSecret: appSecret,
		client:    projectsdk.NewClient(appId, appSecret, projectsdk.WithHeaders(header)),
	}
	if len(debug) == 2 {
		c.debugId = debug[0]
		c.debugSecret = debug[1]
	}
	return c
}

func (c *projectClient) Client() *projectsdk.Client {
	return c.client
}

func (c *projectClient) ListProject(ctx context.Context, userKey string) ([]string, error) {
	req := project.NewListProjectReqBuilder().
		UserKey(userKey).
		Build()
	resp, err := c.client.Project.ListProject(ctx, req)
	if err != nil {
		return nil, err
	}
	if !resp.Success() {
		return nil, resp
	}
	return resp.Data, nil
}

func (c *projectClient) ListWorkItem(ctx context.Context, projectKey string) ([]*workitem.WorkItemInfo, error) {
	req := workitem.NewFilterReqBuilder().
		ProjectKey(projectKey).
		Build()
	resp, err := c.client.WorkItem.Filter(ctx, req)
	if err != nil {
		return nil, err
	}
	if !resp.Success() {
		return nil, resp
	}
	return resp.Data, nil
}

func (c *projectClient) ListTask(ctx context.Context) ([]*task.SubDetail, error) {
	req := task.NewSearchSubtaskReqBuilder().Name("奇波升级").
		Build()
	resp, err := c.client.Task.SearchSubtask(ctx, req)
	if err != nil {
		return nil, err
	}
	if !resp.Success() {
		return nil, resp
	}
	return resp.Data, nil
}

package lark_sdk

import (
	"context"
	"lark-sdk/common/slice"
	"testing"

	print2 "lark-sdk/common/print"
	"lark-sdk/common/sendmsg"
)

func newClient() *LarkClient {
	return NewClient("cli_a5db505ec72b500c", "9MBpBhQmU2pmEch5BVukPr5EiAH2l5ct")
}

func TestNewClient(t *testing.T) {
	newClient()
}

func TestLarkClient_GetUserByUserId(t *testing.T) {
	c := newClient()
	userInfo, err := c.GetUserByUserId(context.Background(), "3291738c")
	if err != nil {
		t.Fatal(err)
	}
	print2.Json(userInfo)
}

func TestLarkClient_GetEmployeeByUserId(t *testing.T) {
	c := newClient()
	employee, err := c.GetEmployeeByUserId(context.Background(), "3291738c")
	if err != nil {
		t.Fatal(err)
	}
	print2.Json(employee)
}

func TestLarkClient_GetAttachment(t *testing.T) {
	c := newClient()
	err := c.GetAttachment(context.Background(), "WuMQb619foEUlVxgQNQcxzd8nAe")
	if err != nil {
		t.Fatal(err)
	}
}

func TestLarkClient_GetDepartmentManagerByDfs(t *testing.T) {
	c := newClient()
	managerByDfs, err := c.GetDepartmentManagerByDfs(context.Background(), "3291738c")
	if err != nil {
		t.Fatal(err)
	}
	print2.Json(managerByDfs)
}

func TestLarkClient_ListUserByDepartmentId(t *testing.T) {
	c := newClient()
	deptIds := []string{"f584acfcee939abf", "19f86abeg64cd7c3", "5bgcf6577f87522f", "d3fgd2ca1418a4eb", "b365a5d1cb5b69a2", "a58bbgg9g63agcg1"}
	res := make([]string, 0)
	for _, dept := range deptIds {
		userIds, err := c.ListUserByDepartmentId(context.Background(), dept)
		if err != nil {
			t.Fatal(err)
		}
		res = append(res, userIds...)
	}
	res = slice.RmDupl(res)
	print2.Json(res)
}

func TestSendCxMsg(t *testing.T) {
	sendmsg.SendCxMsg()
}

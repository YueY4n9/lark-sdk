package lark_sdk

import (
	"context"
	"testing"
)

func newClient() *LarkClient {
	return NewClient("cli_a5db505ec72b500c", "9MBpBhQmU2pmEch5BVukPr5EiAH2l5ct")
}

func TestNewClient(t *testing.T) {
	newClient()
}

func TestLarkClient_GetEmployeeByUserId(t *testing.T) {
	c := newClient()
	employee, err := c.GetEmployeeByUserId(context.Background(), "b4f48ae1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(*employee.UserId)
}

func TestLarkClient_GetAttachment(t *testing.T) {
	c := newClient()
	err := c.GetAttachment(context.Background(), "C1RPbwn86o6nVTxJG9LcTWyen2e")
	if err != nil {
		t.Fatal(err)
	}
}

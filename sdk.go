package lark_sdk

import (
	"context"
	"fmt"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	larkehr "github.com/larksuite/oapi-sdk-go/v3/service/ehr/v1"
	"github.com/pkg/errors"
	"io"
	print2 "lark-sdk/common/print"
	"os"
)

type LarkClient struct {
	client *lark.Client
}

func NewClient(appId, appSecret string) *LarkClient {
	return &LarkClient{client: lark.NewClient(appId, appSecret, lark.WithEnableTokenCache(true))}
}

func (c *LarkClient) GetUserByUserId(ctx context.Context, userId string) (*larkcontact.User, error) {
	resp, err := c.client.Contact.User.Get(ctx, larkcontact.NewGetUserReqBuilder().
		UserId(userId).
		UserIdType("user_id").
		Build())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return nil, errors.New(fmt.Sprintf("%v %v %v", resp.Code, resp.Msg, resp.RequestId()))
	}
	return resp.Data.User, nil
}

func (c *LarkClient) GetEmployeeByUserId(ctx context.Context, userId string) (*larkehr.Employee, error) {
	req := larkehr.NewListEmployeeReqBuilder().
		View("full").
		UserIdType("user_id").
		UserIds([]string{userId}).
		Build()
	resp, err := c.client.Ehr.Employee.List(ctx, req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return nil, errors.New(fmt.Sprintf("%v %v %v", resp.Code, resp.Msg, resp.RequestId()))
	}
	print2.Json(resp.Data.Items[0])
	return resp.Data.Items[0], nil
}

func (c *LarkClient) GetAttachment(ctx context.Context, token string) error {
	resp, err := c.client.Ehr.Attachment.Get(ctx, larkehr.NewGetAttachmentReqBuilder().
		Token(token).
		Build())
	if err != nil {
		fmt.Println(err)
		return err
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return errors.New(fmt.Sprintf("%v %v %v", resp.Code, resp.Msg, resp.RequestId()))
	}
	data, err := io.ReadAll(resp.File)
	if err != nil {
		return err
	}
	if err = os.WriteFile("./temp/"+resp.FileName, data, 0644); err != nil {
		return err
	}
	return nil
}

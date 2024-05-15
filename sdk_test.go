package lark_sdk

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/YueY4n9/gotools/echo"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	"github.com/pkg/errors"
)

func newClient() *LarkClient {
	return NewClient("cli_a6b488d797f8500c", "d91mgv3S0ZBVu0IKpX5vGfkiHWMACAkV")
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
	echo.Json(userInfo)
}

func TestLarkClient_GetEmployeeByUserId(t *testing.T) {
	c := newClient()
	employee, err := c.GetEmpByUserId(context.Background(), "ae5e29b9")
	if err != nil {
		t.Fatal(err)
	}
	echo.Json(employee)
}

func TestLarkClient_GetAttachment(t *testing.T) {
	c := newClient()
	err := c.GetAttachment(context.Background(), "WuMQb619foEUlVxgQNQcxzd8nAe")
	if err != nil {
		t.Fatal(err)
	}
}

func TestLarkClient_SendCardMsg(t *testing.T) {
	c := newClient()
	obj := struct {
		UserName string `json:"user_name"`
		YearNum  string `json:"year_num"`
	}{
		UserName: "岳杨",
		YearNum:  "2",
	}
	err := c.SendCardMsg(context.Background(), "user_id", "3291738c", "ctp_AAiXVQ8gl9eZ", obj)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLarkClient_AllUser(t *testing.T) {
	ctx := context.Background()
	c := newClient()
	allUser, err := c.AllUser(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, user := range allUser {
		for _, field := range user.CustomAttrs {
			if *field.Id == "C-7271191843688349697" && *field.Value.Text == "A049" {
				echo.Json(*user.UserId)
			}
			if *field.Id == "C-7271191843688349697" {
				workstation := *field.Value.Text
				if !strings.Contains(workstation, "-") && !containsChinese(workstation) {
					workstation = replaceWorkstation(workstation)
					err = updateWorkstation(ctx, *user.UserId, workstation)
					if err != nil {
						t.Fatal(err)
					}
					echo.Json(workstation)
				} else if strings.Contains(workstation, "--") {
					workstation = strings.ReplaceAll(workstation, "--", "-")
					err = updateWorkstation(ctx, *user.UserId, workstation)
					if err != nil {
						t.Fatal(err)
					}
					echo.Json(workstation)
				}
				break
			}
		}
	}
}

func containsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Scripts["Han"], r) {
			return true
		}
	}
	return false
}

func replaceWorkstation(workstation string) string {
	if len(workstation) == 5 {
		workstation = strings.Join([]string{workstation[:2], "-", workstation[2:]}, "")
		echo.Json(workstation)
	} else if len(workstation) == 6 {
		workstation = strings.Join([]string{workstation[:3], "-", workstation[3:]}, "")
		echo.Json(workstation)
	}
	return workstation
}

func updateWorkstation(ctx context.Context, userId, workstation string) error {
	c := newClient()
	// 修改用户信息
	req := larkcontact.NewPatchUserReqBuilder().
		UserId(userId).
		UserIdType("user_id").
		User(larkcontact.NewUserBuilder().
			CustomAttrs([]*larkcontact.UserCustomAttr{
				larkcontact.NewUserCustomAttrBuilder().
					Id("C-7271191843688349697"). // 座位号字段的 Id
					Type("HREF").
					Value(larkcontact.NewUserCustomAttrValueBuilder().
						Text(workstation).
						Url("https://ooia5293gn.feishu.cn/wiki/G1A6wmxuxiNsW0kMRx5cr4bpnqe").
						Build()).
					Build(),
			}).
			Build()).
		Build()
	resp, err := c.Client.Contact.User.Patch(ctx, req)
	if err != nil {
		return err
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return errors.New(resp.Msg)
	}
	return nil
}

func TestLarkClient_GetApprovalDefineByCode(t *testing.T) {
	ctx := context.Background()
	c := newClient()
	approval, err := c.GetApprovalDefineByCode(ctx, "89077CAC-C940-490C-B4AC-1C58731B03D5")
	if err != nil {
		t.Fatal(err)
	}
	echo.Json(approval)
}

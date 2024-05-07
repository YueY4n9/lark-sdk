package lark_sdk

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	larkehr "github.com/larksuite/oapi-sdk-go/v3/service/ehr/v1"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
	"io"
	"lark-sdk/common/slice"
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
		DepartmentIdType("department_id").
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

func (c *LarkClient) ListEmployee(ctx context.Context, userIds []string) ([]*larkehr.Employee, error) {
	res := make([]*larkehr.Employee, 0)
	for _, chunk := range slice.ChunkSlice(userIds, 100) {
		req := larkehr.NewListEmployeeReqBuilder().
			View("full").
			UserIdType("user_id").
			UserIds(chunk).
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
		res = append(res, resp.Data.Items...)
	}
	return res, nil
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

func (c *LarkClient) GetUserByDepartmentId(ctx context.Context, departmentId string) ([]*larkcontact.User, error) {
	deptIds := make(map[string]struct{})
	for hasMore, pageToken := true, ""; hasMore; {
		getChildrenDeptReq := larkcontact.NewChildrenDepartmentReqBuilder().
			DepartmentId(departmentId).
			UserIdType("user_id").
			DepartmentIdType("department_id").
			FetchChild(true).
			PageToken(pageToken).
			PageSize(50).
			Build()
		getChildrenDeptResp, err := c.client.Contact.Department.Children(ctx, getChildrenDeptReq)
		if err != nil {
			return nil, err
		}
		if !getChildrenDeptResp.Success() {
			fmt.Println(getChildrenDeptResp.Code, getChildrenDeptResp.Msg, getChildrenDeptResp.RequestId())
			return nil, errors.New(getChildrenDeptResp.Msg)
		}
		hasMore = *getChildrenDeptResp.Data.HasMore
		if hasMore {
			pageToken = *getChildrenDeptResp.Data.PageToken
		}
		for _, department := range getChildrenDeptResp.Data.Items {
			deptIds[*department.DepartmentId] = struct{}{}
		}
	}
	res := make([]*larkcontact.User, 0)
	for deptId := range deptIds {
		for hasMore, pageToken := true, ""; hasMore; {
			getUserReq := larkcontact.NewFindByDepartmentUserReqBuilder().
				UserIdType("user_id").
				DepartmentIdType("department_id").
				DepartmentId(deptId).
				PageToken(pageToken).
				PageSize(50).
				Build()
			getUserResp, err := c.client.Contact.User.FindByDepartment(ctx, getUserReq)
			if err != nil {
				return nil, err
			}
			if !getUserResp.Success() {
				fmt.Println(getUserResp.Code, getUserResp.Msg, getUserResp.RequestId())
				return nil, errors.New(getUserResp.Msg)
			}
			hasMore = *getUserResp.Data.HasMore
			if hasMore {
				pageToken = *getUserResp.Data.PageToken
			}
			for i := range getUserResp.Data.Items {
				res = append(res, getUserResp.Data.Items[i])
			}
		}
	}
	return res, nil
}

func (c *LarkClient) SendMessage(ctx context.Context, userId, msgType, msg string) error {
	msgCreateReq := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType("user_id").
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(userId).
			MsgType(msgType).
			Content(msg).
			Uuid(uuid.New().String()).
			Build()).
		Build()
	msgCreateResp, err := c.client.Im.Message.Create(ctx, msgCreateReq)
	if err != nil {
		return err
	}
	if !msgCreateResp.Success() {
		fmt.Println(msgCreateResp.Code, msgCreateResp.Msg, msgCreateResp.RequestId())
		return errors.New(msgCreateResp.Msg)
	}
	return nil
}

func (c *LarkClient) GetDepartmentManagerByDfs(ctx context.Context, userId string) ([]string, error) {
	res := make([]string, 0)
	userInfo, err := c.GetEmployeeByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	if userInfo.SystemFields.Manager != nil {
		res = append(res, *userInfo.SystemFields.Manager.UserId)
		managers, err := c.GetDepartmentManagerByDfs(ctx, *userInfo.SystemFields.Manager.UserId)
		if err != nil {
			return nil, err
		}
		res = append(res, managers...)
	}
	return res, nil
}

func (c *LarkClient) ListUserByDepartmentId(ctx context.Context, deptId string) ([]string, error) {
	res := make([]string, 0)
	deptIds := make(map[string]struct{})
	deptIds[deptId] = struct{}{}
	for hasMore, pageToken := true, ""; hasMore; {
		getChildrenDeptReq := larkcontact.NewChildrenDepartmentReqBuilder().
			DepartmentId(deptId).
			UserIdType("user_id").
			DepartmentIdType("department_id").
			FetchChild(true).
			PageToken(pageToken).
			PageSize(50).
			Build()
		getChildrenDeptResp, err := c.client.Contact.Department.Children(ctx, getChildrenDeptReq)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		if !getChildrenDeptResp.Success() {
			fmt.Println(getChildrenDeptResp.Code, getChildrenDeptResp.Msg, getChildrenDeptResp.RequestId())
			return nil, err
		}
		hasMore = *getChildrenDeptResp.Data.HasMore
		if hasMore {
			pageToken = *getChildrenDeptResp.Data.PageToken
		}
		for _, department := range getChildrenDeptResp.Data.Items {
			deptIds[*department.DepartmentId] = struct{}{}
		}
	}
	for deptId := range deptIds {
		for hasMore, pageToken := true, ""; hasMore; {
			getUserReq := larkcontact.NewFindByDepartmentUserReqBuilder().
				UserIdType("user_id").
				DepartmentIdType("department_id").
				DepartmentId(deptId).
				PageToken(pageToken).
				PageSize(50).
				Build()
			getUserResp, err := c.client.Contact.User.FindByDepartment(ctx, getUserReq)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			if !getUserResp.Success() {
				fmt.Println(getUserResp.Code, getUserResp.Msg, getUserResp.RequestId())
				return nil, errors.New(getUserResp.Msg)
			}
			hasMore = *getUserResp.Data.HasMore
			if hasMore {
				pageToken = *getUserResp.Data.PageToken
			}
			for _, user := range getUserResp.Data.Items {
				res = append(res, *user.UserId)
			}
		}
	}
	res = slice.RmDupl(res)
	fmt.Printf("lark user count: %v\n", len(res))
	return res, nil
}

func (c *LarkClient) ListRoleMember(ctx context.Context, roleId string) ([]*larkcontact.FunctionalRoleMember, error) {
	req := larkcontact.NewListFunctionalRoleMemberReqBuilder().
		RoleId(roleId).
		UserIdType(`user_id`).
		DepartmentIdType(`department_id`).
		Build()
	resp, err := c.client.Contact.FunctionalRoleMember.List(ctx, req)
	if err != nil {
		return nil, err
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return nil, errors.New(resp.Msg)
	}
	return resp.Data.Members, nil
}

// GetPMRoleByUserId request: contact:functional_role,contact:user.employee_id:readonly,
func (c *LarkClient) GetPMRoleByUserId(ctx context.Context, userId string) ([]string, error) {
	// pm_role_id: 7vb5do17annj7mr
	res := make([]string, 0)
	// 1. 获取 pm 角色下所有成员管理的 user_id 和管理范围的 department_ids
	roleMembers, err := c.ListRoleMember(ctx, "7vb5do17annj7mr")
	if err != nil {
		return nil, err
	}
	// 2. 获取 pm 用户的管理部门下的所有人员
	for _, roleMember := range roleMembers {
		for _, dept := range roleMember.DepartmentIds {
			userIds, err := c.ListUserByDepartmentId(context.Background(), dept)
			if err != nil {
				return nil, err
			}
			res = append(res, userIds...)
		}
		res = slice.RmDupl(res)
		// 3. 每天定时保存到数据库中
	}
	// 3. 将用户的部门列表和每个 pm 角色的部门列表做交集，判断用户是否属于该 pm 管理
	return res, nil
}

func (c *LarkClient) GetDepartmentById(ctx context.Context, departmentId string) (*larkcontact.Department, error) {
	req := larkcontact.NewGetDepartmentReqBuilder().
		DepartmentId(departmentId).
		UserIdType(`user_id`).
		DepartmentIdType(`department_id`).
		Build()
	resp, err := c.client.Contact.Department.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return nil, errors.New(resp.Msg)
	}
	return resp.Data.Department, nil
}

func (c *LarkClient) GetChildDepartment(ctx context.Context, departmentId string) ([]string, error) {
	deptIds := make([]string, 0)
	for hasMore, pageToken := true, ""; hasMore; {
		getChildrenDeptReq := larkcontact.NewChildrenDepartmentReqBuilder().
			DepartmentId(departmentId).
			UserIdType("user_id").
			DepartmentIdType("department_id").
			FetchChild(true).
			PageToken(pageToken).
			PageSize(50).
			Build()
		getChildrenDeptResp, err := c.client.Contact.Department.Children(ctx, getChildrenDeptReq)
		if err != nil {
			return nil, err
		}
		if !getChildrenDeptResp.Success() {
			fmt.Println(getChildrenDeptResp.Code, getChildrenDeptResp.Msg, getChildrenDeptResp.RequestId())
			return nil, errors.New(getChildrenDeptResp.Msg)
		}
		hasMore = *getChildrenDeptResp.Data.HasMore
		if hasMore {
			pageToken = *getChildrenDeptResp.Data.PageToken
		}
		for _, department := range getChildrenDeptResp.Data.Items {
			deptIds[] = *department.DepartmentId
		}
	}
	deptIds = slice.RmDupl(deptIds)
	return deptIds, nil
}

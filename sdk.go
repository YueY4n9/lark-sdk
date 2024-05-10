package lark_sdk

import (
	"context"
	"encoding/json"
	"fmt"
	larkapproval "github.com/larksuite/oapi-sdk-go/v3/service/approval/v4"
	"io"
	"os"

	_slice "github.com/YueY4n9/gotools/slice"
	"github.com/google/uuid"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	larkehr "github.com/larksuite/oapi-sdk-go/v3/service/ehr/v1"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
)

type LarkClient struct {
	Client *lark.Client
}

func NewClient(appId, appSecret string) *LarkClient {
	return &LarkClient{Client: lark.NewClient(appId, appSecret, lark.WithEnableTokenCache(true))}
}

// GetUserByUserId finish
func (c *LarkClient) GetUserByUserId(ctx context.Context, userId string) (*larkcontact.User, error) {
	resp, err := c.Client.Contact.User.Get(ctx, larkcontact.NewGetUserReqBuilder().
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

// GetEmpByUserId finish
func (c *LarkClient) GetEmpByUserId(ctx context.Context, userId string) (*larkehr.Employee, error) {
	req := larkehr.NewListEmployeeReqBuilder().
		View("full").
		UserIdType("user_id").
		UserIds([]string{userId}).
		Build()
	resp, err := c.Client.Ehr.Employee.List(ctx, req)
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

func (c *LarkClient) ListEmp(ctx context.Context, userIds []string) ([]*larkehr.Employee, error) {
	res := make([]*larkehr.Employee, 0)
	for _, chunk := range _slice.ChunkSlice(userIds, 100) {
		req := larkehr.NewListEmployeeReqBuilder().
			View("full").
			UserIdType("user_id").
			UserIds(chunk).
			Build()
		resp, err := c.Client.Ehr.Employee.List(ctx, req)
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

// GetAttachment finish
func (c *LarkClient) GetAttachment(ctx context.Context, token string) error {
	resp, err := c.Client.Ehr.Attachment.Get(ctx, larkehr.NewGetAttachmentReqBuilder().
		Token(token).
		Build())
	if err != nil {
		fmt.Println(err)
		return err
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return errors.New(resp.Msg)
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

// SendMsg finish
func (c *LarkClient) SendMsg(ctx context.Context, receiveIdType, receivedId, msgType, content string) error {
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIdType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(receivedId).
			MsgType(msgType).
			Content(content).
			Uuid(uuid.New().String()).
			Build()).
		Build()
	resp, err := c.Client.Im.Message.Create(ctx, req)
	if err != nil {
		return err
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return errors.New(resp.Msg)
	}
	return nil
}

// SendCardMsg TODO test
func (c *LarkClient) SendCardMsg(ctx context.Context, receiveIdType, receivedId, cardId string, templateVar interface{}) error {
	type msgData struct {
		TemplateId       string      `json:"template_id"`
		TemplateVariable interface{} `json:"template_variable"`
	}
	type message struct {
		Type string  `json:"type"`
		Data msgData `json:"data"`
	}
	m := message{
		Type: "template",
		Data: msgData{
			TemplateId:       cardId,
			TemplateVariable: templateVar,
		},
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return c.SendMsg(ctx, receiveIdType, receivedId, "interactive", string(bytes))
}

// ListUserByDeptId finish
func (c *LarkClient) ListUserByDeptId(ctx context.Context, deptId string) ([]string, error) {
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
		getChildrenDeptResp, err := c.Client.Contact.Department.Children(ctx, getChildrenDeptReq)
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
	for deptId := range deptIds {
		for hasMore, pageToken := true, ""; hasMore; {
			getUserReq := larkcontact.NewFindByDepartmentUserReqBuilder().
				UserIdType("user_id").
				DepartmentIdType("department_id").
				DepartmentId(deptId).
				PageToken(pageToken).
				PageSize(50).
				Build()
			getUserResp, err := c.Client.Contact.User.FindByDepartment(ctx, getUserReq)
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
			for _, user := range getUserResp.Data.Items {
				res = append(res, *user.UserId)
			}
		}
	}
	res = _slice.RemoveDuplication(res)
	fmt.Printf("lark user count: %v\n", len(res))
	return res, nil
}

//func (c *LarkClient) ListRoleMember(ctx context.Context, roleId string) ([]*larkcontact.FunctionalRoleMember, error) {
//	req := larkcontact.NewListFunctionalRoleMemberReqBuilder().
//		RoleId(roleId).
//		UserIdType(`user_id`).
//		DepartmentIdType(`department_id`).
//		Build()
//	resp, err := c.Client.Contact.FunctionalRoleMember.List(ctx, req)
//	if err != nil {
//		return nil, err
//	}
//	if !resp.Success() {
//		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
//		return nil, errors.New(resp.Msg)
//	}
//	return resp.Data.Members, nil
//}

//func (c *LarkClient) GetPMRoleByUserId(ctx context.Context, userId string) ([]string, error) {
//	// pm_role_id: 7vb5do17annj7mr
//	res := make([]string, 0)
//	// 1. 获取 pm 角色下所有成员管理的 user_id 和管理范围的 department_ids
//	roleMembers, err := c.ListRoleMember(ctx, "7vb5do17annj7mr")
//	if err != nil {
//		return nil, err
//	}
//	// 2. 获取 pm 用户的管理部门下的所有人员
//	for _, roleMember := range roleMembers {
//		for _, dept := range roleMember.DepartmentIds {
//			userIds, err := c.ListUserByDeptId(context.Background(), dept)
//			if err != nil {
//				return nil, err
//			}
//			res = append(res, userIds...)
//		}
//		res = _slice.RemoveDuplication(res)
//		// 3. 每天定时保存到数据库中
//	}
//	// 3. 将用户的部门列表和每个 pm 角色的部门列表做交集，判断用户是否属于该 pm 管理
//	return res, nil
//}

// GetDeptById finish
func (c *LarkClient) GetDeptById(ctx context.Context, departmentId string) (*larkcontact.Department, error) {
	req := larkcontact.NewGetDepartmentReqBuilder().
		DepartmentId(departmentId).
		UserIdType(`user_id`).
		DepartmentIdType(`department_id`).
		Build()
	resp, err := c.Client.Contact.Department.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return nil, errors.New(resp.Msg)
	}
	return resp.Data.Department, nil
}

// ListChildDeptByDeptId finish
func (c *LarkClient) ListChildDeptByDeptId(ctx context.Context, deptId string) ([]*larkcontact.Department, error) {
	res := make([]*larkcontact.Department, 0)
	deptInfo, err := c.GetDeptById(ctx, deptId)
	if err != nil {
		return nil, err
	}
	res = append(res, deptInfo)
	for hasMore, pageToken := true, ""; hasMore; {
		req := larkcontact.NewChildrenDepartmentReqBuilder().
			DepartmentId(deptId).
			UserIdType("user_id").
			DepartmentIdType("department_id").
			FetchChild(true).
			PageToken(pageToken).
			PageSize(50).
			Build()
		getChildrenDeptResp, err := c.Client.Contact.Department.Children(ctx, req)
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
		for _, item := range getChildrenDeptResp.Data.Items {
			res = append(res, item)
		}
	}
	return res, nil
}

// ListChildDeptIdByDeptId finish
func (c *LarkClient) ListChildDeptIdByDeptId(ctx context.Context, deptId string) ([]string, error) {
	res := make([]string, 0)
	depts, err := c.ListChildDeptByDeptId(ctx, deptId)
	if err != nil {
		return nil, err
	}
	for _, dept := range depts {
		res = append(res, *dept.DepartmentId)
	}
	return res, nil
}

//func (c *LarkClient) GetChildDepartmentMap(ctx context.Context, departmentId string) (map[string]string, error) {
//	deptMap := make(map[string]string)
//	deptMap["d4e276efc6ac5fee"] = "上海本社"
//	parentMap := make(map[string]string)
//	for hasMore, pageToken := true, ""; hasMore; {
//		getChildrenDeptReq := larkcontact.NewChildrenDepartmentReqBuilder().
//			DepartmentId(departmentId).
//			UserIdType("user_id").
//			DepartmentIdType("department_id").
//			FetchChild(true).
//			PageToken(pageToken).
//			PageSize(50).
//			Build()
//		getChildrenDeptResp, err := c.Client.Contact.Department.Children(ctx, getChildrenDeptReq)
//		if err != nil {
//			return nil, err
//		}
//		if !getChildrenDeptResp.Success() {
//			fmt.Println(getChildrenDeptResp.Code, getChildrenDeptResp.Msg, getChildrenDeptResp.RequestId())
//			return nil, errors.New(getChildrenDeptResp.Msg)
//		}
//		hasMore = *getChildrenDeptResp.Data.HasMore
//		if hasMore {
//			pageToken = *getChildrenDeptResp.Data.PageToken
//		}
//		for _, department := range getChildrenDeptResp.Data.Items {
//			deptMap[*department.DepartmentId] = *department.Name
//			parentMap[*department.DepartmentId] = *department.ParentDepartmentId
//		}
//	}
//	deptMap = buildDeptId2PathMap(deptMap, parentMap)
//	return deptMap, nil
//}

//func buildPath(deptId string, deptId2NameMap map[string]string, deptId2ParentIdMap map[string]string) string {
//	if parent, ok := deptId2ParentIdMap[deptId]; ok {
//		return buildPath(parent, deptId2NameMap, deptId2ParentIdMap) + "-" + deptId2NameMap[deptId]
//	}
//	return deptId2NameMap[deptId]
//}

//func buildDeptId2PathMap(deptId2NameMap map[string]string, deptId2ParentIdMap map[string]string) map[string]string {
//	deptId2PathMap := make(map[string]string)
//	for deptId := range deptId2NameMap {
//		deptId2PathMap[deptId] = buildPath(deptId, deptId2NameMap, deptId2ParentIdMap)
//	}
//	return deptId2PathMap
//}

//func (c *LarkClient) GetDepartmentManagerByDfs(ctx context.Context, userId string) ([]string, error) {
//	res := make([]string, 0)
//	userInfo, err := c.GetEmpByUserId(ctx, userId)
//	if err != nil {
//		return nil, err
//	}
//	if userInfo.SystemFields.Manager != nil {
//		res = append(res, *userInfo.SystemFields.Manager.UserId)
//		managers, err := c.GetDepartmentManagerByDfs(ctx, *userInfo.SystemFields.Manager.UserId)
//		if err != nil {
//			return nil, err
//		}
//		res = append(res, managers...)
//	}
//	return res, nil
//}

// AllEmp finish
func (c *LarkClient) AllEmp(ctx context.Context) ([]*larkehr.Employee, error) {
	res := make([]*larkehr.Employee, 0)
	for hasMore, pageToken := true, ""; hasMore; {
		employeeReqBuilder := larkehr.NewListEmployeeReqBuilder().
			View("full").
			PageSize(100).UserIdType("user_id")
		if pageToken != "" {
			employeeReqBuilder.PageToken(pageToken)
		}
		req := employeeReqBuilder.Build()
		resp, err := c.Client.Ehr.Employee.List(ctx, req)
		if err != nil {
			return nil, err
		}
		if !resp.Success() {
			fmt.Println(resp.Code, resp.Msg, resp.RequestId())
			return nil, errors.New(resp.Msg)
		}
		hasMore = *resp.Data.HasMore
		if hasMore {
			pageToken = *resp.Data.PageToken
		}
		for _, item := range resp.Data.Items {
			res = append(res, item)
		}
	}
	return res, nil
}

// AllUserId finish
func (c *LarkClient) AllUserId(ctx context.Context) ([]string, error) {
	userIds := make([]string, 0)
	allEmp, err := c.AllEmp(ctx)
	if err != nil {
		return nil, err
	}
	for _, emp := range allEmp {
		userIds = append(userIds, *emp.UserId)
	}
	return userIds, nil
}

// SubscribeApproval finish
func (c *LarkClient) SubscribeApproval(ctx context.Context, code string) error {
	req := larkapproval.NewSubscribeApprovalReqBuilder().
		ApprovalCode(code).
		Build()
	resp, err := c.Client.Approval.Approval.Subscribe(ctx, req)
	if err != nil {
		return err
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return errors.New(resp.Msg)
	}
	return nil
}

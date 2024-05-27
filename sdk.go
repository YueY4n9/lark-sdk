package lark_sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/YueY4n9/gotools/echo"
	_map "github.com/YueY4n9/gotools/map"
	_slice "github.com/YueY4n9/gotools/slice"
	"github.com/google/uuid"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkapproval "github.com/larksuite/oapi-sdk-go/v3/service/approval/v4"
	larkattendance "github.com/larksuite/oapi-sdk-go/v3/service/attendance/v1"
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

func (c *LarkClient) GetEmpNameMap(ctx context.Context, userIds []string) (map[string]string, error) {
	emps := make([]*larkehr.Employee, 0)
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
		emps = append(emps, resp.Data.Items...)
	}
	res := make(map[string]string)
	for _, emp := range emps {
		res[*emp.UserId] = *emp.SystemFields.Name
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

// SendCardMsg finish
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

func (c *LarkClient) ListUserByDeptId(ctx context.Context, deptId string) ([]*larkcontact.User, error) {
	res := make([]*larkcontact.User, 0)
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
				res = append(res, user)
			}
		}
	}
	userSet := make(map[string]*larkcontact.User)
	for _, userInfo := range res {
		userSet[*userInfo.UserId] = userInfo
	}
	res = _map.GetValues(userSet)
	return res, nil
}

// ListUserIdByDeptId finish
func (c *LarkClient) ListUserIdByDeptId(ctx context.Context, deptId string) ([]string, error) {
	res := make([]string, 0)
	users, err := c.ListUserByDeptId(ctx, deptId)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		res = append(res, *user.UserId)
	}
	return _slice.RemoveDuplication(res), nil
}

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
	return _slice.RemoveDuplication(res), nil
}

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

// AllUser finish
func (c *LarkClient) AllUser(ctx context.Context) ([]*larkcontact.User, error) {
	res, err := c.ListUserByDeptId(ctx, "0")
	if err != nil {
		return nil, err
	}
	echo.Json(len(res))
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

// GetApprovalByCode
// Deprecated, this function renamed to GetApprovalDefineByCode, v1.0.0 will remove this function
func (c *LarkClient) GetApprovalByCode(ctx context.Context, code string) (*larkapproval.GetApprovalRespData, error) {
	req := larkapproval.NewGetApprovalReqBuilder().
		ApprovalCode(code).
		Locale("zh-CN").
		WithAdminId(true).
		UserIdType("user_id").
		Build()
	resp, err := c.Client.Approval.Approval.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return nil, errors.New(resp.Msg)
	}
	return resp.Data, nil
}

// GetApprovalDefineByCode finish
func (c *LarkClient) GetApprovalDefineByCode(ctx context.Context, code string) (*larkapproval.GetApprovalRespData, error) {
	req := larkapproval.NewGetApprovalReqBuilder().
		ApprovalCode(code).
		Locale("zh-CN").
		WithAdminId(true).
		UserIdType("user_id").
		Build()
	resp, err := c.Client.Approval.Approval.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return nil, errors.New(resp.Msg)
	}
	return resp.Data, nil
}

// ListApprovalInstIdByCode finish
func (c *LarkClient) ListApprovalInstIdByCode(ctx context.Context, code, startTime, endTime string) ([]string, error) {
	res := make([]string, 0)
	for hasMore, pageToken := true, ""; hasMore; {
		req := larkapproval.NewListInstanceReqBuilder().
			ApprovalCode(code).
			StartTime(startTime).
			EndTime(endTime).
			PageToken(pageToken).
			PageSize(100).
			Build()
		resp, err := c.Client.Approval.Instance.List(ctx, req)
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
		res = append(res, resp.Data.InstanceCodeList...)
	}
	return res, nil
}

// GetApprovalInstById finish
func (c *LarkClient) GetApprovalInstById(ctx context.Context, instId string) (*larkapproval.GetInstanceRespData, error) {
	req := larkapproval.NewGetInstanceReqBuilder().
		InstanceId(instId).
		Build()
	resp, err := c.Client.Approval.Instance.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return nil, errors.New(resp.Msg)
	}
	return resp.Data, nil
}

// SearchApprovalInst user_id、approval_code、instance_code、instance_external_id、group_external_id 不得均为空
// approval_code 和 group_external_id 查询结果取并集，instance_code 和 instance_external_id 查询结果取并集，其他查询条件都对应取交集
// 查询时间跨度不得大于30天，开始和结束时间必须都设置，或者都不设置
func (c *LarkClient) SearchApprovalInst(ctx context.Context, userId, approvalCode, instCode, instStatus, timeFrom, timeTo string) ([]*larkapproval.InstanceSearchItem, error) {
	req := larkapproval.NewQueryInstanceReqBuilder().
		UserIdType("user_id").
		InstanceSearch(larkapproval.NewInstanceSearchBuilder().
			UserId(userId).
			ApprovalCode(approvalCode).
			InstanceCode(instCode).
			InstanceStatus(instStatus).
			InstanceStartTimeFrom(timeFrom).
			InstanceStartTimeTo(timeTo).
			Build()).
		Build()
	resp, err := c.Client.Approval.Instance.Query(ctx, req)
	if err != nil {
		return nil, err
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return nil, errors.New(resp.Msg)
	}
	return resp.Data.InstanceList, nil
}

// CreateApprovalInst finish
func (c *LarkClient) CreateApprovalInst(ctx context.Context, approvalCode, userId string, form interface{}) error {
	bytes, err := json.Marshal(form)
	if err != nil {
		return errors.WithStack(err)
	}
	req := larkapproval.NewCreateInstanceReqBuilder().
		InstanceCreate(larkapproval.NewInstanceCreateBuilder().
			ApprovalCode(approvalCode).
			UserId(userId).
			Form(string(bytes)).
			Build()).
		Build()
	resp, err := c.Client.Approval.Instance.Create(ctx, req)
	if err != nil {
		return err
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return errors.WithStack(errors.New(resp.Msg))
	}
	return nil
}

// ListAttendanceRecord dataFrom:20060102
func (c *LarkClient) ListAttendanceRecord(ctx context.Context, userIds []string, dateFrom, dateTo int) ([]*larkattendance.UserTask, error) {
	res := make([]*larkattendance.UserTask, 0)
	for _, chunk := range _slice.ChunkSlice(userIds, 50) {
		req := larkattendance.NewQueryUserTaskReqBuilder().
			EmployeeType("employee_id").
			IncludeTerminatedUser(false).
			Body(larkattendance.NewQueryUserTaskReqBodyBuilder().
				UserIds(chunk).
				CheckDateFrom(dateFrom).
				CheckDateTo(dateTo).
				NeedOvertimeResult(false).
				Build()).
			Build()
		resp, err := c.Client.Attendance.UserTask.Query(ctx, req)
		if err != nil {
			return nil, err
		}
		if !resp.Success() {
			fmt.Println(resp.Code, resp.Msg, resp.RequestId())
			return nil, errors.New(resp.Msg)
		}
		for _, userTask := range resp.Data.UserTaskResults {
			res = append(res, userTask)
		}
	}
	return res, nil
}

// ListRoleMember finish
func (c *LarkClient) ListRoleMember(ctx context.Context, roleId string) ([]*larkcontact.FunctionalRoleMember, error) {
	req := larkcontact.NewListFunctionalRoleMemberReqBuilder().
		RoleId(roleId).
		UserIdType(`user_id`).
		DepartmentIdType(`department_id`).
		Build()
	resp, err := c.Client.Contact.FunctionalRoleMember.List(ctx, req)
	if err != nil {
		return nil, err
	}
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return nil, errors.New(resp.Msg)
	}
	return resp.Data.Members, nil
}

// RollbackApprovalTask TODO
func (c *LarkClient) RollbackApprovalTask(ctx context.Context, userId, taskId, reason string, defKeys []string) error {
	return nil
}

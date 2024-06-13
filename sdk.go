package lark_sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/YueY4n9/gotools/echo"
	_map "github.com/YueY4n9/gotools/map"
	_slice "github.com/YueY4n9/gotools/slice"
	"github.com/google/uuid"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkapplication "github.com/larksuite/oapi-sdk-go/v3/service/application/v6"
	larkapproval "github.com/larksuite/oapi-sdk-go/v3/service/approval/v4"
	larkattendance "github.com/larksuite/oapi-sdk-go/v3/service/attendance/v1"
	larkcalendar "github.com/larksuite/oapi-sdk-go/v3/service/calendar/v4"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	larkehr "github.com/larksuite/oapi-sdk-go/v3/service/ehr/v1"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkvc "github.com/larksuite/oapi-sdk-go/v3/service/vc/v1"
)

type LarkClient interface {
	Client() *lark.Client

	GetUserByUserId(ctx context.Context, userId string) (*larkcontact.User, error)
	GetEmpByUserId(ctx context.Context, userId string) (*larkehr.Employee, error)
	GetEmpNameMap(ctx context.Context, userIds []string) (map[string]string, error)
	ListEmp(ctx context.Context, userIds []string) ([]*larkehr.Employee, error)
	AllUser(ctx context.Context) ([]*larkcontact.User, error)
	AllEmp(ctx context.Context) ([]*larkehr.Employee, error)
	ListUserByDeptId(ctx context.Context, deptId string) ([]*larkcontact.User, error)
	AllUserId(ctx context.Context) ([]string, error)
	ListUserIdByDeptId(ctx context.Context, deptId string) ([]string, error)

	GetDeptById(ctx context.Context, departmentId string) (*larkcontact.Department, error)
	ListChildDeptByDeptId(ctx context.Context, deptId string) ([]*larkcontact.Department, error)
	ListChildDeptIdByDeptId(ctx context.Context, deptId string) ([]string, error)

	SendMsg(ctx context.Context, receiveIdType, receivedId, msgType, content string) error
	SendCardMsg(ctx context.Context, receiveIdType, receivedId, cardId string, templateVar interface{}) error

	SubscribeApproval(ctx context.Context, code string) error
	GetApprovalDefineByCode(ctx context.Context, code string) (*larkapproval.GetApprovalRespData, error)
	ListApprovalInstIdByCode(ctx context.Context, code, startTime, endTime string) ([]string, error)
	GetApprovalInstById(ctx context.Context, instId string) (*larkapproval.GetInstanceRespData, error)
	SearchApprovalInst(ctx context.Context, userId, approvalCode, instCode, instStatus, timeFrom, timeTo string) ([]*larkapproval.InstanceSearchItem, error)
	CreateApprovalInst(ctx context.Context, approvalCode, userId string, form interface{}) error
	RollbackApprovalTask(ctx context.Context, currUserId, currTaskId, reason string, defKeys []string) error
	AddSign(ctx context.Context, operatorId, approvalCode, instCode, taskId, comment string, addSignUserIds []string, addSignType, approvalMethod int) error
	ApproveTask(ctx context.Context, approvalCode, instCode, userId, comment, taskId, form string) error

	ListRoom(ctx context.Context, roomLevelId string) ([]*larkvc.Room, error)
	CheckRoomFree(ctx context.Context, roomId, timeMin, timeMax string) (bool, error)
	SetCalendarRoom(ctx context.Context, calendarId, eventId, roomId string) error
	SetCalendarUsers(ctx context.Context, calendarId, eventId, userId string) error
	ListCalendarEvent(ctx context.Context, calendarId string) ([]*larkcalendar.CalendarEvent, error)
	CreateCalendarEvent(ctx context.Context, calendarId, summary, startTs, endTs string) (*larkcalendar.CalendarEvent, error)

	GetAttachment(ctx context.Context, token string) error
	ListAttendanceRecord(ctx context.Context, userIds []string, dateFrom, dateTo int) ([]*larkattendance.UserTask, error)
	ListRoleMember(ctx context.Context, roleId string) ([]*larkcontact.FunctionalRoleMember, error)
	GetAppInfo(appId string) *larkapplication.Application
	Alert(err error)
}

type larkClient struct {
	appId       string
	appSecret   string
	debugId     string
	debugSecret string
	client      *lark.Client
}

func NewClient(appId, appSecret string, debug ...string) LarkClient {
	c := &larkClient{
		appId:     appId,
		appSecret: appSecret,
		client:    lark.NewClient(appId, appSecret, lark.WithEnableTokenCache(true)),
	}
	if len(debug) == 2 {
		c.debugId = debug[0]
		c.debugSecret = debug[1]
	}
	return c
}

func (c *larkClient) Client() *lark.Client {
	return c.client
}

func (c *larkClient) GetUserByUserId(ctx context.Context, userId string) (*larkcontact.User, error) {
	resp, err := c.client.Contact.User.Get(ctx, larkcontact.NewGetUserReqBuilder().
		UserId(userId).
		UserIdType(UserId).
		DepartmentIdType(DepartmentId).
		Build())
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(resp)
		return nil, resp
	}
	return resp.Data.User, nil
}
func (c *larkClient) GetEmpByUserId(ctx context.Context, userId string) (*larkehr.Employee, error) {
	req := larkehr.NewListEmployeeReqBuilder().
		View("full").
		UserIdType(UserId).
		UserIds([]string{userId}).
		Build()
	resp, err := c.client.Ehr.Employee.List(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(resp)
		return nil, resp
	}
	return resp.Data.Items[0], nil
}
func (c *larkClient) GetEmpNameMap(ctx context.Context, userIds []string) (map[string]string, error) {
	employees, err := c.ListEmp(ctx, userIds)
	if err != nil {
		return nil, err
	}
	res := make(map[string]string)
	for _, emp := range employees {
		if emp.UserId != nil && emp.SystemFields.Name != nil {
			res[*emp.UserId] = *emp.SystemFields.Name
		}
	}
	return res, nil
}
func (c *larkClient) ListEmp(ctx context.Context, userIds []string) ([]*larkehr.Employee, error) {
	res := make([]*larkehr.Employee, 0)
	for _, chunk := range _slice.ChunkSlice(userIds, 100) {
		req := larkehr.NewListEmployeeReqBuilder().
			View("full").
			UserIdType(UserId).
			UserIds(chunk).
			Build()
		resp, err := c.client.Ehr.Employee.List(ctx, req)
		if err != nil {
			c.Alert(err)
			return nil, err
		}
		if !resp.Success() {
			c.Alert(resp)
			return nil, resp
		}
		res = append(res, resp.Data.Items...)
	}
	return res, nil
}
func (c *larkClient) AllEmp(ctx context.Context) ([]*larkehr.Employee, error) {
	res := make([]*larkehr.Employee, 0)
	for hasMore, pageToken := true, ""; hasMore; {
		employeeReqBuilder := larkehr.NewListEmployeeReqBuilder().
			View("full").
			PageSize(100).
			UserIdType(UserId).
			Status([]int{2})
		if pageToken != "" {
			employeeReqBuilder.PageToken(pageToken)
		}
		req := employeeReqBuilder.Build()
		resp, err := c.client.Ehr.Employee.List(ctx, req)
		if err != nil {
			c.Alert(err)
			return nil, err
		}
		if !resp.Success() {
			c.Alert(resp)
			return nil, resp
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
func (c *larkClient) AllUser(ctx context.Context) ([]*larkcontact.User, error) {
	res, err := c.ListUserByDeptId(ctx, "0")
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (c *larkClient) ListUserByDeptId(ctx context.Context, deptId string) ([]*larkcontact.User, error) {
	res := make([]*larkcontact.User, 0)
	deptIds, err := c.ListChildDeptIdByDeptId(ctx, deptId)
	if err != nil {
		return nil, err
	}
	for _, deptId := range deptIds {
		for hasMore, pageToken := true, ""; hasMore; {
			req := larkcontact.NewFindByDepartmentUserReqBuilder().
				UserIdType(UserId).
				DepartmentIdType(DepartmentId).
				DepartmentId(deptId).
				PageToken(pageToken).
				PageSize(50).
				Build()
			resp, err := c.client.Contact.User.FindByDepartment(ctx, req)
			if err != nil {
				c.Alert(err)
				return nil, err
			}
			if !resp.Success() {
				c.Alert(resp)
				return nil, resp
			}
			hasMore = *resp.Data.HasMore
			if hasMore {
				pageToken = *resp.Data.PageToken
			}
			for _, user := range resp.Data.Items {
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
func (c *larkClient) AllUserId(ctx context.Context) ([]string, error) {
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
func (c *larkClient) ListUserIdByDeptId(ctx context.Context, deptId string) ([]string, error) {
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
func (c *larkClient) GetDeptById(ctx context.Context, departmentId string) (*larkcontact.Department, error) {
	req := larkcontact.NewGetDepartmentReqBuilder().
		DepartmentId(departmentId).
		UserIdType(UserId).
		DepartmentIdType(`department_id`).
		Build()
	resp, err := c.client.Contact.Department.Get(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(resp)
		return nil, resp
	}
	return resp.Data.Department, nil
}
func (c *larkClient) ListChildDeptByDeptId(ctx context.Context, deptId string) ([]*larkcontact.Department, error) {
	res := make([]*larkcontact.Department, 0)
	deptInfo, err := c.GetDeptById(ctx, deptId)
	if err != nil {
		return nil, err
	}
	res = append(res, deptInfo)
	for hasMore, pageToken := true, ""; hasMore; {
		req := larkcontact.NewChildrenDepartmentReqBuilder().
			DepartmentId(deptId).
			UserIdType(UserId).
			DepartmentIdType(DepartmentId).
			FetchChild(true).
			PageToken(pageToken).
			PageSize(50).
			Build()
		resp, err := c.client.Contact.Department.Children(ctx, req)
		if err != nil {
			c.Alert(err)
			return nil, err
		}
		if !resp.Success() {
			c.Alert(resp)
			return nil, resp
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
func (c *larkClient) ListChildDeptIdByDeptId(ctx context.Context, deptId string) ([]string, error) {
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
func (c *larkClient) SendMsg(ctx context.Context, receiveIdType, receivedId, msgType, content string) error {
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIdType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(receivedId).
			MsgType(msgType).
			Content(content).
			Uuid(uuid.New().String()).
			Build()).
		Build()
	resp, err := c.client.Im.Message.Create(ctx, req)
	if err != nil {
		return err
	}
	if !resp.Success() {
		return resp
	}
	return nil
}
func (c *larkClient) SendCardMsg(ctx context.Context, receiveIdType, receivedId, cardId string, templateVar interface{}) error {
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
func (c *larkClient) SubscribeApproval(ctx context.Context, code string) error {
	req := larkapproval.NewSubscribeApprovalReqBuilder().
		ApprovalCode(code).
		Build()
	resp, err := c.client.Approval.Approval.Subscribe(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(resp)
		return resp
	}
	return nil
}
func (c *larkClient) GetApprovalDefineByCode(ctx context.Context, code string) (*larkapproval.GetApprovalRespData, error) {
	req := larkapproval.NewGetApprovalReqBuilder().
		ApprovalCode(code).
		Locale("zh-CN").
		WithAdminId(true).
		UserIdType(UserId).
		Build()
	resp, err := c.client.Approval.Approval.Get(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(resp)
		return nil, resp
	}
	return resp.Data, nil
}
func (c *larkClient) ListApprovalInstIdByCode(ctx context.Context, code, startTime, endTime string) ([]string, error) {
	res := make([]string, 0)
	for hasMore, pageToken := true, ""; hasMore; {
		req := larkapproval.NewListInstanceReqBuilder().
			ApprovalCode(code).
			StartTime(startTime).
			EndTime(endTime).
			PageToken(pageToken).
			PageSize(100).
			Build()
		resp, err := c.client.Approval.Instance.List(ctx, req)
		if err != nil {
			c.Alert(err)
			return nil, err
		}
		if !resp.Success() {
			c.Alert(resp)
			return nil, resp
		}
		hasMore = *resp.Data.HasMore
		if hasMore {
			pageToken = *resp.Data.PageToken
		}
		res = append(res, resp.Data.InstanceCodeList...)
	}
	return res, nil
}
func (c *larkClient) GetApprovalInstById(ctx context.Context, instId string) (*larkapproval.GetInstanceRespData, error) {
	req := larkapproval.NewGetInstanceReqBuilder().
		InstanceId(instId).
		Build()
	resp, err := c.client.Approval.Instance.Get(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(resp)
		return nil, resp
	}
	return resp.Data, nil
}

// SearchApprovalInst user_id、approval_code、instance_code、instance_external_id、group_external_id 不得均为空
// approval_code 和 group_external_id 查询结果取并集，instance_code 和 instance_external_id 查询结果取并集，其他查询条件都对应取交集
// 查询时间跨度不得大于30天，开始和结束时间必须都设置，或者都不设置
func (c *larkClient) SearchApprovalInst(ctx context.Context, userId, approvalCode, instCode, instStatus, timeFrom, timeTo string) ([]*larkapproval.InstanceSearchItem, error) {
	req := larkapproval.NewQueryInstanceReqBuilder().
		UserIdType(UserId).
		InstanceSearch(larkapproval.NewInstanceSearchBuilder().
			UserId(userId).
			ApprovalCode(approvalCode).
			InstanceCode(instCode).
			InstanceStatus(instStatus).
			InstanceStartTimeFrom(timeFrom).
			InstanceStartTimeTo(timeTo).
			Build()).
		Build()
	resp, err := c.client.Approval.Instance.Query(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(resp)
		return nil, resp
	}
	return resp.Data.InstanceList, nil
}
func (c *larkClient) CreateApprovalInst(ctx context.Context, approvalCode, userId string, form interface{}) error {
	bytes, err := json.Marshal(form)
	if err != nil {
		return err
	}
	req := larkapproval.NewCreateInstanceReqBuilder().
		InstanceCreate(larkapproval.NewInstanceCreateBuilder().
			ApprovalCode(approvalCode).
			UserId(userId).
			Form(string(bytes)).
			Build()).
		Build()
	resp, err := c.client.Approval.Instance.Create(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(resp)
		return resp
	}
	return nil
}
func (c *larkClient) GetAttachment(ctx context.Context, token string) error {
	resp, err := c.client.Ehr.Attachment.Get(ctx, larkehr.NewGetAttachmentReqBuilder().
		Token(token).
		Build())
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(resp)
		return resp
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

// ListAttendanceRecord dataFrom,dataTo:20060102
func (c *larkClient) ListAttendanceRecord(ctx context.Context, userIds []string, dateFrom, dateTo int) ([]*larkattendance.UserTask, error) {
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
		resp, err := c.client.Attendance.UserTask.Query(ctx, req)
		if err != nil {
			c.Alert(err)
			return nil, err
		}
		if !resp.Success() {
			c.Alert(resp)
			return nil, resp
		}
		for _, userTask := range resp.Data.UserTaskResults {
			res = append(res, userTask)
		}
	}
	return res, nil
}
func (c *larkClient) ListRoleMember(ctx context.Context, roleId string) ([]*larkcontact.FunctionalRoleMember, error) {
	req := larkcontact.NewListFunctionalRoleMemberReqBuilder().
		RoleId(roleId).
		UserIdType(UserId).
		DepartmentIdType(`department_id`).
		Build()
	resp, err := c.client.Contact.FunctionalRoleMember.List(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(resp)
		return nil, resp
	}
	return resp.Data.Members, nil
}
func (c *larkClient) RollbackApprovalTask(ctx context.Context, currUserId, currTaskId, reason string, defKeys []string) error {
	req := larkapproval.NewSpecifiedRollbackInstanceReqBuilder().
		UserIdType(UserId).
		SpecifiedRollback(larkapproval.NewSpecifiedRollbackBuilder().
			UserId(currUserId).
			TaskId(currTaskId).
			Reason(reason).
			TaskDefKeyList(defKeys).
			Build()).
		Build()
	resp, err := c.client.Approval.Instance.SpecifiedRollback(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(resp)
		return resp
	}
	return nil
}
func (c *larkClient) GetAppInfo(appId string) *larkapplication.Application {
	req := larkapplication.NewGetApplicationReqBuilder().
		AppId(appId).
		Lang(`zh_cn`).
		Build()
	resp, err := c.client.Application.Application.Get(context.Background(), req)
	if err != nil || !resp.Success() {
		return nil
	}
	return resp.Data.App
}
func (c *larkClient) Alert(err error) {
	if c.debugId == "" {
		return
	}
	client := NewClient(c.debugId, c.debugSecret)
	appInfo := client.GetAppInfo(c.appId)
	appName := "未知"
	if appInfo != nil {
		appName = *appInfo.AppName
	}
	obj := struct {
		AppId   string `json:"app_id"`
		AppName string `json:"app_name"`
		Err     string `json:"err"`
	}{
		AppId:   c.appId,
		AppName: appName,
		Err:     err.Error(),
	}
	err = client.SendCardMsg(context.Background(), UserId, "3291738c", "AAq3zkrIEYCqR", obj)
	if err != nil {
		echo.Json(err)
	}
}

func (c *larkClient) AddSign(ctx context.Context, operatorId, approvalCode, instCode, taskId, comment string, addSignUserIds []string, addSignType, approvalMethod int) error {
	req := larkapproval.NewAddSignInstanceReqBuilder().
		Body(larkapproval.NewAddSignInstanceReqBodyBuilder().
			UserId(operatorId).
			ApprovalCode(approvalCode).
			InstanceCode(instCode).
			TaskId(taskId).
			Comment(comment).
			AddSignUserIds(addSignUserIds).
			AddSignType(addSignType).
			ApprovalMethod(approvalMethod).
			Build()).
		Build()
	resp, err := c.client.Approval.V4.Instance.AddSign(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(err)
		return resp
	}
	return nil
}
func (c *larkClient) ApproveTask(ctx context.Context, approvalCode, instCode, userId, comment, taskId, form string) error {
	req := larkapproval.NewApproveTaskReqBuilder().
		TaskApprove(larkapproval.NewTaskApproveBuilder().
			ApprovalCode(approvalCode).
			InstanceCode(instCode).
			UserId(userId).
			Comment(comment).
			TaskId(taskId).
			Form(form).
			Build()).
		Build()
	resp, err := c.client.Approval.Task.Approve(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(err)
		return resp
	}
	return nil
}
func (c *larkClient) ListRoom(ctx context.Context, roomLevelId string) ([]*larkvc.Room, error) {
	res := make([]*larkvc.Room, 0)
	req := larkvc.NewListRoomReqBuilder().
		UserIdType(UserId).
		RoomLevelId(roomLevelId).
		PageSize(100).
		Build()
	resp, err := c.client.Vc.Room.List(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(resp)
		return nil, resp
	}
	for _, room := range resp.Data.Rooms {
		res = append(res, room)
	}
	return res, nil
}
func (c *larkClient) CheckRoomFree(ctx context.Context, roomId, timeMin, timeMax string) (bool, error) {
	req := larkcalendar.NewListFreebusyReqBuilder().
		UserIdType(UserId).
		Body(larkcalendar.NewListFreebusyReqBodyBuilder().
			TimeMin(timeMin).
			TimeMax(timeMax).
			RoomId(roomId).
			Build()).
		Build()
	resp, err := c.client.Calendar.Freebusy.List(ctx, req)
	if err != nil {
		c.Alert(err)
		return false, err
	}
	if !resp.Success() {
		c.Alert(resp)
		return false, resp
	}
	if len(resp.Data.FreebusyList) == 0 {
		return true, nil
	}
	echo.Json(resp.Data.FreebusyList)
	return false, nil
}
func (c *larkClient) SetCalendarRoom(ctx context.Context, calendarId, eventId, roomId string) error {
	req := larkcalendar.NewCreateCalendarEventAttendeeReqBuilder().
		CalendarId(calendarId).
		EventId(eventId).
		UserIdType(UserId).
		Body(larkcalendar.NewCreateCalendarEventAttendeeReqBodyBuilder().
			Attendees([]*larkcalendar.CalendarEventAttendee{
				larkcalendar.NewCalendarEventAttendeeBuilder().
					Type(`resource`).
					RoomId(roomId).
					ApprovalReason("会议室申请嘻嘻").
					Build(),
			}).
			NeedNotification(true).
			IsEnableAdmin(false).
			AddOperatorToAttendee(false).
			Build()).
		Build()
	resp, err := c.client.Calendar.CalendarEventAttendee.Create(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(resp)
		return resp
	}
	return nil
}
func (c *larkClient) SetCalendarUsers(ctx context.Context, calendarId, eventId, userId string) error {
	req := larkcalendar.NewCreateCalendarEventAttendeeReqBuilder().
		CalendarId(calendarId).
		EventId(eventId).
		UserIdType(UserId).
		Body(larkcalendar.NewCreateCalendarEventAttendeeReqBodyBuilder().
			Attendees([]*larkcalendar.CalendarEventAttendee{
				larkcalendar.NewCalendarEventAttendeeBuilder().
					Type(User).
					UserId(userId).
					Build(),
			}).
			AddOperatorToAttendee(false).
			Build()).
		Build()
	resp, err := c.client.Calendar.CalendarEventAttendee.Create(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(resp)
		return resp
	}
	return nil
}
func (c *larkClient) ListCalendarEvent(ctx context.Context, calendarId string) ([]*larkcalendar.CalendarEvent, error) {
	now := time.Now()
	zeroTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	req := larkcalendar.NewListCalendarEventReqBuilder().
		CalendarId(calendarId).
		PageSize(500).
		StartTime(fmt.Sprint(zeroTime.Unix())).
		UserIdType(UserId).
		Build()
	resp, err := c.client.Calendar.CalendarEvent.List(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(resp)
		return nil, resp
	}
	return resp.Data.Items, nil
}
func (c *larkClient) CreateCalendarEvent(ctx context.Context, calendarId, summary, startTs, endTs string) (*larkcalendar.CalendarEvent, error) {
	req := larkcalendar.NewCreateCalendarEventReqBuilder().
		CalendarId(calendarId).
		UserIdType(UserId).
		CalendarEvent(larkcalendar.NewCalendarEventBuilder().
			Summary(summary).
			Description("").
			NeedNotification(false).
			StartTime(larkcalendar.NewTimeInfoBuilder().
				Timestamp(startTs).
				Build()).
			EndTime(larkcalendar.NewTimeInfoBuilder().
				Timestamp(endTs).
				Build()).
			AttendeeAbility(`can_invite_others`).
			Color(-1).
			Build()).
		Build()
	resp, err := c.client.Calendar.CalendarEvent.Create(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(resp)
		return nil, resp
	}
	return resp.Data.Event, nil
}

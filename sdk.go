package lark_sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/YueY4n9/gotools/echo"
	_map "github.com/YueY4n9/gotools/map"
	_slice "github.com/YueY4n9/gotools/slice"
	"github.com/google/uuid"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkapplication "github.com/larksuite/oapi-sdk-go/v3/service/application/v6"
	larkapproval "github.com/larksuite/oapi-sdk-go/v3/service/approval/v4"
	larkattendance "github.com/larksuite/oapi-sdk-go/v3/service/attendance/v1"
	larkauthen "github.com/larksuite/oapi-sdk-go/v3/service/authen/v1"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	larkcalendar "github.com/larksuite/oapi-sdk-go/v3/service/calendar/v4"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	larkcorehr "github.com/larksuite/oapi-sdk-go/v3/service/corehr/v2"
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	larkehr "github.com/larksuite/oapi-sdk-go/v3/service/ehr/v1"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larksecurityandcompliance "github.com/larksuite/oapi-sdk-go/v3/service/security_and_compliance/v1"
	larkvc "github.com/larksuite/oapi-sdk-go/v3/service/vc/v1"
	larkwiki "github.com/larksuite/oapi-sdk-go/v3/service/wiki/v2"
	"github.com/pkg/errors"
)

const (
	maxRetry  int = 3
	sleepTime     = time.Second
)

type LarkClient interface {
	Client() *lark.Client
	GetAppName() string

	GetUserAccessToken(ctx context.Context, code string) (string, error)

	GetUserById(ctx context.Context, id, userIdType, deptIdType string) (*larkcontact.User, error)
	GetUserByUserId(ctx context.Context, userId string) (*larkcontact.User, error)
	GetUserByOpenId(ctx context.Context, openId string) (*larkcontact.User, error)
	GetEmpByUserId(ctx context.Context, userId string) (*larkehr.Employee, error)
	GetEmpNameMap(ctx context.Context, userIds []string) (map[string]string, error)
	ListEmp(ctx context.Context, userIds []string) ([]*larkehr.Employee, error)
	AllUser(ctx context.Context) ([]*larkcontact.User, error)
	AllEmp(ctx context.Context) ([]*larkehr.Employee, error)
	ListUserByDeptId(ctx context.Context, deptIdType, deptId string) ([]*larkcontact.User, error)
	AllUserId(ctx context.Context) ([]string, error)
	ListUserIdByDeptId(ctx context.Context, deptIdType, deptId string) ([]string, error)

	//部门
	GetDeptById(ctx context.Context, deptIdType, deptId string) (*larkcontact.Department, error)
	ListChildDeptByDeptId(ctx context.Context, deptIdType string, deptId string) ([]*larkcontact.Department, error)
	ListChildDeptIdByDeptId(ctx context.Context, deptIdType string, deptId string) ([]string, error)
	ListParentDeptByDeptId(ctx context.Context, deptIdType string, deptId string) ([]*larkcontact.Department, error)

	// 消息
	SendMsg(ctx context.Context, receiveIdType, receivedId, msgType, content string) error
	SendCardMsg(ctx context.Context, receiveIdType, receivedId, cardId string, templateVar interface{}) error

	// 审批
	SubscribeApproval(ctx context.Context, code string) error
	UnsubscribeApproval(ctx context.Context, code string) error
	GetApprovalDefineByCode(ctx context.Context, code string) (*larkapproval.GetApprovalRespData, error)
	ListApprovalInstIdByCode(ctx context.Context, code, startTime, endTime string) ([]string, error)
	GetApprovalInstById(ctx context.Context, instId string) (*larkapproval.GetInstanceRespData, error)
	SearchApprovalInst(ctx context.Context, userId, approvalCode, instCode, instStatus string) ([]*larkapproval.InstanceSearchItem, error)
	CreateApprovalInst(ctx context.Context, approvalCode, userId string, form interface{}, nodeApprover []*larkapproval.NodeApprover) error
	RollbackApprovalTask(ctx context.Context, currUserId, currTaskId, reason string, defKeys []string) error
	AddSign(ctx context.Context, operatorId, approvalCode, instCode, taskId, comment string, addSignUserIds []string, addSignType, approvalMethod int) error
	ApproveTask(ctx context.Context, approvalCode, instCode, userId, comment, taskId, form string) error
	CcApprovalInst(ctx context.Context, approvalCode, instCode, fromUserId, comment string, ccUserIds []string) error
	AddInstComment(ctx context.Context, instCode, userId, comment string) error
	SearchUserApprovalTask(ctx context.Context, userId, taskStatus string) ([]*larkapproval.TaskSearchItem, error)
	RejectTask(ctx context.Context, approvalCode, instCode, userId, comment, taskId string) error

	// 假勤
	ListLeaveData(ctx context.Context, from, to time.Time, userIds []string) ([]*larkattendance.UserApproval, error)
	GetAttendanceGroup(ctx context.Context, groupId string) (*larkattendance.GetGroupRespData, error)
	SetShift(ctx context.Context, groupId, shiftId string, userIds []string, date time.Time) error
	ListAttendanceStats(ctx context.Context, from, to time.Time, userIds []string) ([]*larkattendance.UserStatsData, error)

	// 会议室
	ListRoom(ctx context.Context, roomLevelId string) ([]*larkvc.Room, error)
	CheckRoomFree(ctx context.Context, roomId, timeMin, timeMax string) (bool, error)
	SetCalendarRoom(ctx context.Context, calendarId, eventId, roomId string) error
	SetCalendarUsers(ctx context.Context, calendarId, eventId string, userIds []string) error
	ListCalendarEvent(ctx context.Context, calendarId string) ([]*larkcalendar.CalendarEvent, error)
	CreateCalendarEvent(ctx context.Context, calendarId, summary, startTs, endTs string) (*larkcalendar.CalendarEvent, error)

	// 云文档
	GetSpaceNode(ctx context.Context, objType, token string) (*larkwiki.Node, error)
	ListBitableRecord(ctx context.Context, appToken, tableId, userIdType string, fieldNames []string, sort []*larkbitable.Sort, filter *larkbitable.FilterInfo) ([]*larkbitable.AppTableRecord, error)
	InsertBitableRecord(ctx context.Context, appToken, tableId, userIdType string, records []*larkbitable.AppTableRecord) error
	InsertBitable1Record(ctx context.Context, appToken, tableId, userIdType string, record *larkbitable.AppTableRecord) error
	UpdateBitableRecord(ctx context.Context, appToken, tableId, userIdType string, records []*larkbitable.AppTableRecord) error
	CopySpaceNode(ctx context.Context, spaceId, nodeToken, targetParentToken, nodeName string) (*larkwiki.Node, error)
	SubscribeFile(ctx context.Context, fileToken, fileType string) error
	GetRecord(ctx context.Context, appToken, tableId, recordId string) (*larkbitable.AppTableRecord, error)

	// 人事企业版流程
	GetProcess(ctx context.Context, processId string) (*larkcorehr.GetProcessRespData, error)

	// 其他
	GetAttachment(ctx context.Context, token string) error
	ListAttendanceRecord(ctx context.Context, userIds []string, dateFrom, dateTo int) ([]*larkattendance.UserTask, error)
	ListRoleMember(ctx context.Context, roleId string) ([]*larkcontact.FunctionalRoleMember, error)
	GetAppInfo(appId string) *larkapplication.Application
	AddAttendanceFlow(ctx context.Context, userId, locationName, checkTime string) error
	GetLog(ctx context.Context, appId, apiKey string) ([]*larksecurityandcompliance.OpenapiLog, error)
	Alert(err error)
}

type larkClient struct {
	appId       string
	appSecret   string
	appName     string
	debugId     string
	debugSecret string
	adminUserId string
	client      *lark.Client
}

func NewClient(appId, appSecret string, debug ...string) LarkClient {
	c := &larkClient{
		appId:     appId,
		appSecret: appSecret,
		client:    lark.NewClient(appId, appSecret, lark.WithEnableTokenCache(true)),
	}
	if len(debug) >= 2 {
		c.debugId = debug[0]
		c.debugSecret = debug[1]
		c.appName = *NewClient(c.debugId, c.debugSecret).GetAppInfo(appId).AppName
		c.adminUserId = "3291738c"
	}
	if len(debug) >= 3 {
		c.adminUserId = debug[2]
	}
	return c
}

func (c *larkClient) Client() *lark.Client {
	return c.client
}
func (c *larkClient) GetAppName() string {
	return c.appName
}

func (c *larkClient) GetUserByUserId(ctx context.Context, userId string) (*larkcontact.User, error) {
	return c.GetUserById(ctx, userId, UserId, DepartmentId)
}
func (c *larkClient) GetUserByOpenId(ctx context.Context, openId string) (*larkcontact.User, error) {
	return c.GetUserById(ctx, openId, OpenId, DepartmentId)
}
func (c *larkClient) GetUserById(ctx context.Context, id, userIdType, deptIdType string) (*larkcontact.User, error) {
	resp, err := c.client.Contact.User.Get(ctx, larkcontact.NewGetUserReqBuilder().
		UserId(id).
		UserIdType(userIdType).
		DepartmentIdType(deptIdType).
		Build())
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return nil, resp
	}
	return resp.Data.User, nil
}
func (c *larkClient) getEmp(ctx context.Context, userIdType string, userIds []string) ([]*larkehr.Employee, error) {
	res := make([]*larkehr.Employee, 0)
	for _, chunk := range _slice.ChunkSlice(userIds, 100) {
		req := larkehr.NewListEmployeeReqBuilder().
			View("full").
			UserIdType(userIdType).
			UserIds(chunk).
			Build()
		resp, err := c.client.Ehr.Employee.List(ctx, req)
		if err != nil {
			c.Alert(err)
			return nil, err
		}
		if !resp.Success() {
			if resp.Code == 1241001 {
				time.Sleep(sleepTime)
				return c.getEmp(ctx, userIdType, userIds)
			} else {
				c.Alert(errors.New(string(resp.RawBody)))
				return nil, resp
			}
		}
		res = append(res, resp.Data.Items...)
	}
	return res, nil
}
func (c *larkClient) GetEmpByUserId(ctx context.Context, userId string) (*larkehr.Employee, error) {
	employees, err := c.getEmp(ctx, UserId, []string{userId})
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if len(employees) == 0 {
		return nil, nil
	}
	return employees[0], nil
}
func (c *larkClient) GetEmpNameMap(ctx context.Context, userIds []string) (map[string]string, error) {
	employees, err := c.getEmp(ctx, UserId, userIds)
	if err != nil {
		c.Alert(err)
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
	return c.getEmp(ctx, UserId, userIds)
}
func (c *larkClient) AllEmp(ctx context.Context) ([]*larkehr.Employee, error) {
	res := make([]*larkehr.Employee, 0)
	for hasMore, pageToken := true, ""; hasMore; {
		employeeReqBuilder := larkehr.NewListEmployeeReqBuilder().
			View("full").
			PageSize(100).
			UserIdType(UserId).
			Status([]int{2, 4})
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
			c.Alert(errors.New(string(resp.RawBody)))
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
func (c *larkClient) AllUser(ctx context.Context) (res []*larkcontact.User, err error) {
	startTime := time.Now()
	for i := 0; i < maxRetry; i++ {
		res, err = c.ListUserByDeptId(ctx, DepartmentId, "0")
		if err == nil {
			break
		} else {
			c.Alert(err)
			time.Sleep(sleepTime)
		}
	}
	endTime := time.Now()
	echo.Json(endTime.Sub(startTime).String())
	return res, err
}
func (c *larkClient) ListUserByDeptId(ctx context.Context, deptIdType, deptId string) ([]*larkcontact.User, error) {
	res := make([]*larkcontact.User, 0)
	childDeptIds, err := c.ListChildDeptIdByDeptId(ctx, deptIdType, deptId)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	for _, childDeptId := range childDeptIds {
		for hasMore, pageToken := true, ""; hasMore; {
			req := larkcontact.NewFindByDepartmentUserReqBuilder().
				UserIdType(UserId).
				DepartmentIdType(DepartmentId).
				DepartmentId(childDeptId).
				PageToken(pageToken).
				PageSize(50).
				Build()
			resp, err := c.client.Contact.User.FindByDepartment(ctx, req)
			if err != nil {
				c.Alert(err)
				return nil, err
			}
			if !resp.Success() {
				c.Alert(errors.New(string(resp.RawBody)))
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
		c.Alert(err)
		return nil, err
	}
	for _, emp := range allEmp {
		userIds = append(userIds, *emp.UserId)
	}
	return userIds, nil
}
func (c *larkClient) ListUserIdByDeptId(ctx context.Context, deptIdType, deptId string) ([]string, error) {
	res := make([]string, 0)
	users, err := c.ListUserByDeptId(ctx, deptIdType, deptId)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	for _, user := range users {
		res = append(res, *user.UserId)
	}
	return _slice.RemoveDuplication(res), nil
}
func (c *larkClient) GetDeptById(ctx context.Context, deptIdType, deptId string) (*larkcontact.Department, error) {
	req := larkcontact.NewGetDepartmentReqBuilder().
		DepartmentId(deptId).
		UserIdType(UserId).
		DepartmentIdType(deptIdType).
		Build()
	resp, err := c.client.Contact.Department.Get(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return nil, resp
	}
	return resp.Data.Department, nil
}
func (c *larkClient) ListChildDeptByDeptId(ctx context.Context, deptIdType string, deptId string) ([]*larkcontact.Department, error) {
	res := make([]*larkcontact.Department, 0)
	deptInfo, err := c.GetDeptById(ctx, deptIdType, deptId)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	res = append(res, deptInfo)
	for hasMore, pageToken := true, ""; hasMore; {
		req := larkcontact.NewChildrenDepartmentReqBuilder().
			DepartmentId(deptId).
			UserIdType(UserId).
			DepartmentIdType(deptIdType).
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
			c.Alert(errors.New(string(resp.RawBody)))
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
func (c *larkClient) ListChildDeptIdByDeptId(ctx context.Context, deptIdType string, deptId string) ([]string, error) {
	res := make([]string, 0)
	deptInfoList, err := c.ListChildDeptByDeptId(ctx, deptIdType, deptId)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	for _, dept := range deptInfoList {
		if dept.DepartmentId == nil {
			return nil, errors.New("DepartmentId is nil")
		}
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
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		c.Alert(errors.New(fmt.Sprintf("sendMsg to %s error, content: %s", receivedId, content)))
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
		c.Alert(err)
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
		c.Alert(errors.New(string(resp.RawBody)))
		return resp
	}
	return nil
}
func (c *larkClient) UnsubscribeApproval(ctx context.Context, code string) error {
	req := larkapproval.NewUnsubscribeApprovalReqBuilder().
		ApprovalCode(code).
		Build()
	resp, err := c.client.Approval.Approval.Unsubscribe(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
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
		c.Alert(errors.New(string(resp.RawBody)))
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
			c.Alert(errors.New(string(resp.RawBody)))
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
		c.Alert(errors.New(string(resp.RawBody)))
		return nil, resp
	}
	return resp.Data, nil
}
func (c *larkClient) SearchApprovalInst(ctx context.Context, userId, approvalCode, instCode, instStatus string) ([]*larkapproval.InstanceSearchItem, error) {
	res := make([]*larkapproval.InstanceSearchItem, 0)
	for hasMore, pageToken := true, ""; hasMore; {
		req := larkapproval.NewQueryInstanceReqBuilder().
			PageSize(200).
			PageToken(pageToken).
			UserIdType(UserId).
			InstanceSearch(larkapproval.NewInstanceSearchBuilder().
				ApprovalCode(approvalCode).
				InstanceCode(instCode).
				InstanceStatus(instStatus).
				UserId(userId).
				Locale(`zh-CN`).
				Build()).
			Build()
		resp, err := c.client.Approval.Instance.Query(ctx, req)
		if err != nil {
			c.Alert(err)
			return nil, err
		}
		if !resp.Success() {
			c.Alert(errors.New(string(resp.RawBody)))
			return nil, resp
		}
		hasMore = *resp.Data.HasMore
		if hasMore {
			pageToken = *resp.Data.PageToken
		}
		for _, item := range resp.Data.InstanceList {
			res = append(res, item)
		}
	}
	return res, nil
}
func (c *larkClient) CreateApprovalInst(ctx context.Context, approvalCode, userId string, form interface{}, nodeApprover []*larkapproval.NodeApprover) error {
	bytes, err := json.Marshal(form)
	if err != nil {
		c.Alert(err)
		return err
	}
	req := larkapproval.NewCreateInstanceReqBuilder().
		InstanceCreate(larkapproval.NewInstanceCreateBuilder().
			ApprovalCode(approvalCode).
			UserId(userId).
			Form(string(bytes)).
			NodeApproverUserIdList(nodeApprover).
			Build()).
		Build()
	resp, err := c.client.Approval.Instance.Create(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
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
		c.Alert(errors.New(string(resp.RawBody)))
		return resp
	}
	data, err := io.ReadAll(resp.File)
	if err != nil {
		c.Alert(err)
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
			c.Alert(errors.New(string(resp.RawBody)))
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
		c.Alert(errors.New(string(resp.RawBody)))
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
		c.Alert(errors.New(string(resp.RawBody)))
		return resp
	}
	return nil
}
func (c *larkClient) GetAppInfo(appId string) *larkapplication.Application {
	req := larkapplication.NewGetApplicationReqBuilder().
		AppId(appId).
		UserIdType(UserId).
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
		ErrTime string `json:"err_time"`
		Logid   string `json:"logid"`
	}{
		AppId:   c.appId,
		AppName: appName,
		Err:     err.Error(),
		ErrTime: time.Now().Format(time.DateTime),
	}
	err = client.SendCardMsg(context.Background(), UserId, c.adminUserId, "AAq3zkrIEYCqR", obj)
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
		c.Alert(errors.New(string(resp.RawBody)))
		return resp
	}
	return nil
}
func (c *larkClient) ApproveTask(ctx context.Context, approvalCode, instCode, userId, comment, taskId, form string) error {
	taskApproveBuilder := larkapproval.NewTaskApproveBuilder().
		ApprovalCode(approvalCode).
		InstanceCode(instCode).
		UserId(userId).
		Comment(comment).
		TaskId(taskId)
	if len(form) > 0 {
		taskApproveBuilder = taskApproveBuilder.Form(form)
	}
	req := larkapproval.NewApproveTaskReqBuilder().
		UserIdType(UserId).
		TaskApprove(taskApproveBuilder.Build()).
		Build()
	resp, err := c.client.Approval.Task.Approve(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
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
		c.Alert(errors.New(string(resp.RawBody)))
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
		c.Alert(errors.New(string(resp.RawBody)))
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
		c.Alert(errors.New(string(resp.RawBody)))
		return resp
	}
	return nil
}
func (c *larkClient) SetCalendarUsers(ctx context.Context, calendarId, eventId string, userIds []string) error {
	attendees := make([]*larkcalendar.CalendarEventAttendee, 0)
	for _, userId := range userIds {
		attendees = append(attendees, larkcalendar.NewCalendarEventAttendeeBuilder().Type(User).UserId(userId).Build())
	}
	req := larkcalendar.NewCreateCalendarEventAttendeeReqBuilder().
		CalendarId(calendarId).
		EventId(eventId).
		UserIdType(UserId).
		Body(larkcalendar.NewCreateCalendarEventAttendeeReqBodyBuilder().
			Attendees(attendees).
			AddOperatorToAttendee(false).
			Build()).
		Build()
	resp, err := c.client.Calendar.CalendarEventAttendee.Create(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
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
		c.Alert(errors.New(string(resp.RawBody)))
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
		c.Alert(errors.New(string(resp.RawBody)))
		return nil, resp
	}
	return resp.Data.Event, nil
}
func (c *larkClient) CcApprovalInst(ctx context.Context, approvalCode, instCode, fromUserId, comment string, ccUserIds []string) error {
	req := larkapproval.NewCcInstanceReqBuilder().
		UserIdType(UserId).
		InstanceCc(larkapproval.NewInstanceCcBuilder().
			ApprovalCode(approvalCode).
			InstanceCode(instCode).
			UserId(fromUserId).
			CcUserIds(ccUserIds).
			Comment(comment).
			Build()).
		Build()
	resp, err := c.client.Approval.Instance.Cc(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return resp
	}
	c.Alert(errors.Errorf("instCode: %v from: %v cc: %v", instCode, fromUserId, ccUserIds))
	return nil
}
func (c *larkClient) AddInstComment(ctx context.Context, instCode, userId, comment string) error {
	req := larkapproval.NewCreateInstanceCommentReqBuilder().
		InstanceId(instCode).
		UserIdType(UserId).
		UserId(userId).
		CommentRequest(larkapproval.NewCommentRequestBuilder().
			Content(comment).
			DisableBot(true).
			Build()).
		Build()
	resp, err := c.client.Approval.InstanceComment.Create(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return resp
	}
	return nil
}
func (c *larkClient) SearchAppTableRecord(ctx context.Context, appToken, tableId string, fieldNames []string, info *larkbitable.FilterInfo) ([]*larkbitable.AppTableRecord, error) {
	req := larkbitable.NewSearchAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableId).
		PageSize(500).
		Body(larkbitable.NewSearchAppTableRecordReqBodyBuilder().
			FieldNames(fieldNames).
			Filter(info).
			AutomaticFields(false).
			Build()).
		Build()
	resp, err := c.client.Bitable.AppTableRecord.Search(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return nil, resp
	}
	return resp.Data.Items, nil
}
func (c *larkClient) GetUserAccessToken(ctx context.Context, code string) (string, error) {
	req := larkauthen.NewCreateOidcAccessTokenReqBuilder().
		Body(larkauthen.NewCreateOidcAccessTokenReqBodyBuilder().
			GrantType(`authorization_code`).
			Code(code).
			Build()).
		Build()
	resp, err := c.client.Authen.OidcAccessToken.Create(ctx, req)
	if err != nil {
		c.Alert(err)
		return "", err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return "", resp
	}
	return *resp.Data.AccessToken, nil
}
func (c *larkClient) AddAttendanceFlow(ctx context.Context, userId, locationName, checkTime string) error {
	req := larkattendance.NewBatchCreateUserFlowReqBuilder().
		EmployeeType(`employee_id`).
		Body(larkattendance.NewBatchCreateUserFlowReqBodyBuilder().
			FlowRecords([]*larkattendance.UserFlow{
				larkattendance.NewUserFlowBuilder().
					UserId(userId).
					CreatorId(userId).
					LocationName(locationName).
					CheckTime(checkTime).
					Comment("").
					Build(),
			}).
			Build()).
		Build()
	resp, err := c.client.Attendance.UserFlow.BatchCreate(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return resp
	}
	return nil
}
func (c *larkClient) ListBitableRecord(ctx context.Context, appToken, tableId, userIdType string, fieldNames []string, sort []*larkbitable.Sort, filter *larkbitable.FilterInfo) ([]*larkbitable.AppTableRecord, error) {
	res := make([]*larkbitable.AppTableRecord, 0)
	for hasMore, pageToken := true, ""; hasMore; {
		req := larkbitable.NewSearchAppTableRecordReqBuilder().
			AppToken(appToken).
			TableId(tableId).
			UserIdType(userIdType).
			PageSize(500).
			PageToken(pageToken).
			Body(larkbitable.NewSearchAppTableRecordReqBodyBuilder().
				FieldNames(fieldNames).
				Sort(sort).
				Filter(filter).
				Build()).
			Build()
		resp, err := c.client.Bitable.AppTableRecord.Search(ctx, req)
		if err != nil {
			c.Alert(err)
			return nil, err
		}
		if !resp.Success() {
			c.Alert(errors.New(string(resp.RawBody)))
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
func (c *larkClient) InsertBitableRecord(ctx context.Context, appToken, tableId, userIdType string, records []*larkbitable.AppTableRecord) error {
	for _, chunk := range _slice.ChunkSlice(records, 500) {
		req := larkbitable.NewBatchCreateAppTableRecordReqBuilder().
			AppToken(appToken).
			TableId(tableId).
			UserIdType(userIdType).
			Body(larkbitable.NewBatchCreateAppTableRecordReqBodyBuilder().
				Records(chunk).
				Build()).
			Build()
		resp, err := c.client.Bitable.AppTableRecord.BatchCreate(ctx, req)
		if err != nil {
			c.Alert(err)
			return err
		}
		if !resp.Success() {
			c.Alert(errors.New(string(resp.RawBody)))
			return resp
		}
	}
	return nil
}
func (c *larkClient) InsertBitable1Record(ctx context.Context, appToken, tableId, userIdType string, record *larkbitable.AppTableRecord) error {
	req := larkbitable.NewCreateAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableId).
		UserIdType(userIdType).
		AppTableRecord(record).
		Build()
	resp, err := c.client.Bitable.AppTableRecord.Create(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return resp
	}
	return nil
}
func (c *larkClient) UpdateBitableRecord(ctx context.Context, appToken, tableId, userIdType string, records []*larkbitable.AppTableRecord) error {
	for _, chunk := range _slice.ChunkSlice(records, 500) {
		req := larkbitable.NewBatchUpdateAppTableRecordReqBuilder().
			AppToken(appToken).
			TableId(tableId).
			UserIdType(userIdType).
			Body(larkbitable.NewBatchUpdateAppTableRecordReqBodyBuilder().
				Records(chunk).
				Build()).
			Build()
		resp, err := c.client.Bitable.AppTableRecord.BatchUpdate(ctx, req)
		if err != nil {
			c.Alert(err)
			return err
		}
		if !resp.Success() {
			c.Alert(errors.New(string(resp.RawBody)))
			return resp
		}
	}
	return nil
}
func (c *larkClient) GetSpaceNode(ctx context.Context, objType, token string) (*larkwiki.Node, error) {
	req := larkwiki.NewGetNodeSpaceReqBuilder().
		ObjType(objType).
		Token(token).
		Build()
	resp, err := c.client.Wiki.Space.GetNode(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return nil, resp
	}
	return resp.Data.Node, nil
}
func (c *larkClient) GetLog(ctx context.Context, appId, apiKey string) ([]*larksecurityandcompliance.OpenapiLog, error) {
	req := larksecurityandcompliance.NewListDataOpenapiLogReqBuilder().
		ListOpenapiLogRequest(larksecurityandcompliance.NewListOpenapiLogRequestBuilder().
			ApiKeys([]string{apiKey}).
			StartTime(1724896800).
			EndTime(1724904000).
			AppId(appId).
			PageSize(100).
			Build()).
		Build()
	resp, err := c.client.SecurityAndCompliance.OpenapiLog.ListData(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return nil, resp
	}
	return resp.Data.Items, nil
}
func (c *larkClient) ListParentDeptByDeptId(ctx context.Context, deptIdType string, deptId string) ([]*larkcontact.Department, error) {
	req := larkcontact.NewParentDepartmentReqBuilder().
		UserIdType(UserId).
		DepartmentIdType(deptIdType).
		DepartmentId(deptId).
		PageSize(20).
		Build()
	resp, err := c.client.Contact.Department.Parent(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return nil, resp
	}
	_slice.Reverse(resp.Data.Items)
	deptInfo, err := c.GetDeptById(ctx, deptIdType, deptId)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	resp.Data.Items = append(resp.Data.Items, deptInfo)
	return resp.Data.Items, nil
}
func (c *larkClient) RejectTask(ctx context.Context, approvalCode, instCode, userId, comment, taskId string) error {
	req := larkapproval.NewRejectTaskReqBuilder().
		UserIdType(UserId).
		TaskApprove(larkapproval.NewTaskApproveBuilder().
			ApprovalCode(approvalCode).
			InstanceCode(instCode).
			UserId(userId).
			TaskId(taskId).
			Comment(comment).
			Build()).
		Build()
	resp, err := c.client.Approval.Task.Reject(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return resp
	}
	return nil
}
func (c *larkClient) SearchUserApprovalTask(ctx context.Context, userId, taskStatus string) ([]*larkapproval.TaskSearchItem, error) {
	req := larkapproval.NewSearchTaskReqBuilder().UserIdType(`user_id`).
		TaskSearch(larkapproval.NewTaskSearchBuilder().
			UserId(userId).
			TaskStatus(taskStatus).
			Build()).
		Build()
	resp, err := c.client.Approval.Task.Search(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return nil, resp
	}
	return resp.Data.TaskList, nil
}
func (c *larkClient) ListLeaveData(ctx context.Context, from, to time.Time, userIds []string) ([]*larkattendance.UserApproval, error) {
	const maxDays = 30 // 最大支持的时间间隔
	result := make([]*larkattendance.UserApproval, 0)
	// 切分时间区间，保证每个区间不超过30天
	var timeRanges []struct {
		Start time.Time
		End   time.Time
	}
	for start := from; !start.After(to); {
		end := start.AddDate(0, 0, maxDays-1) // 每段最多30天
		if end.After(to) {
			end = to
		}
		timeRanges = append(timeRanges, struct {
			Start time.Time
			End   time.Time
		}{Start: start, End: end})
		start = end.AddDate(0, 0, 1) // 下一段从 end 的下一天开始
	}
	for _, timeRange := range timeRanges {
		fromNum, err := strconv.Atoi(timeRange.Start.Format("20060102"))
		if err != nil {
			return nil, err
		}
		toNum, err := strconv.Atoi(timeRange.End.Format("20060102"))
		if err != nil {
			return nil, err
		}
		for _, chunk := range _slice.ChunkSlice(userIds, 50) {
			req := larkattendance.NewQueryUserApprovalReqBuilder().
				EmployeeType("employee_id").
				Body(larkattendance.NewQueryUserApprovalReqBodyBuilder().
					UserIds(chunk).
					CheckDateFrom(fromNum).
					CheckDateTo(toNum).
					Build()).
				Build()
			resp, err := c.client.Attendance.UserApproval.Query(ctx, req)
			if err != nil {
				c.Alert(err)
				return nil, err
			}
			if !resp.Success() {
				c.Alert(errors.New(string(resp.RawBody)))
				return nil, resp
			}
			for _, approval := range resp.Data.UserApprovals {
				if len(approval.Leaves) > 0 {
					result = append(result, approval)
				}
			}
		}
	}
	return result, nil
}

func (c *larkClient) SetShift(ctx context.Context, groupId, shiftId string, userIds []string, date time.Time) error {
	month := date.Year()*100 + int(date.Month())
	dayNo := date.Day()
	for _, chunk := range _slice.ChunkSlice(userIds, 50) {
		userDailyShifts := make([]*larkattendance.UserDailyShift, 0)
		for i := range chunk {
			userDailyShifts = append(userDailyShifts, larkattendance.NewUserDailyShiftBuilder().
				GroupId(groupId).
				ShiftId(shiftId).
				Month(month).
				UserId(chunk[i]).
				DayNo(dayNo).
				Build())
		}
		req := larkattendance.NewBatchCreateUserDailyShiftReqBuilder().
			EmployeeType("employee_id").
			Body(larkattendance.NewBatchCreateUserDailyShiftReqBodyBuilder().
				UserDailyShifts(userDailyShifts).
				OperatorId(`manjuurobot`).
				Build()).
			Build()
		resp, err := c.client.Attendance.UserDailyShift.BatchCreate(ctx, req)
		if err != nil {
			c.Alert(err)
			return err
		}
		if !resp.Success() {
			c.Alert(errors.New(string(resp.RawBody)))
			return resp
		}
	}
	return nil
}
func (c *larkClient) GetAttendanceGroup(ctx context.Context, groupId string) (*larkattendance.GetGroupRespData, error) {
	req := larkattendance.NewGetGroupReqBuilder().
		GroupId(groupId).
		EmployeeType("employee_id").
		DeptType("open_id").
		Build()
	resp, err := c.client.Attendance.Group.Get(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return nil, resp
	}
	return resp.Data, nil
}
func (c *larkClient) CopySpaceNode(ctx context.Context, spaceId, nodeToken, targetParentToken, nodeName string) (*larkwiki.Node, error) {
	req := larkwiki.NewCopySpaceNodeReqBuilder().
		SpaceId(spaceId).
		NodeToken(nodeToken).
		Body(larkwiki.NewCopySpaceNodeReqBodyBuilder().
			TargetParentToken(targetParentToken).
			Title(nodeName).
			Build()).
		Build()
	resp, err := c.client.Wiki.SpaceNode.Copy(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return nil, resp
	}
	return resp.Data.Node, nil
}
func (c *larkClient) SubscribeFile(ctx context.Context, fileToken, fileType string) error {
	req := larkdrive.NewSubscribeFileReqBuilder().
		FileToken(fileToken).
		FileType(fileType).
		Build()
	resp, err := c.client.Drive.File.Subscribe(ctx, req)
	if err != nil {
		c.Alert(err)
		return err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return resp
	}
	return nil
}
func (c *larkClient) ListAttendanceStats(ctx context.Context, from, to time.Time, userIds []string) ([]*larkattendance.UserStatsData, error) {
	startDate, err := strconv.Atoi(from.Format("20060102"))
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	endDate, err := strconv.Atoi(to.Format("20060102"))
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	var res []*larkattendance.UserStatsData
	for _, chunk := range _slice.ChunkSlice(userIds, 200) {
		req := larkattendance.NewQueryUserStatsDataReqBuilder().
			EmployeeType(`employee_id`).
			Body(larkattendance.NewQueryUserStatsDataReqBodyBuilder().
				Locale(`zh`).
				StatsType(`month`).
				StartDate(startDate).
				EndDate(endDate).
				UserIds(chunk).
				NeedHistory(true).
				CurrentGroupOnly(true).
				UserId(c.adminUserId).
				Build()).
			Build()
		resp, err := c.client.Attendance.UserStatsData.Query(ctx, req)
		if err != nil {
			c.Alert(err)
			return nil, err
		}
		if !resp.Success() {
			c.Alert(errors.New(string(resp.RawBody)))
			return nil, resp
		}
		res = append(res, resp.Data.UserDatas...)
	}
	return res, nil
}

func (c *larkClient) GetRecord(ctx context.Context, appToken, tableId, recordId string) (*larkbitable.AppTableRecord, error) {
	req := larkbitable.NewGetAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableId).
		RecordId(recordId).
		Build()
	resp, err := c.client.Bitable.V1.AppTableRecord.Get(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return nil, resp
	}
	return resp.Data.Record, nil
}

func (c *larkClient) GetProcess(ctx context.Context, processId string) (*larkcorehr.GetProcessRespData, error) {
	req := larkcorehr.NewGetProcessReqBuilder().
		ProcessId(processId).
		UserIdType(UserId).
		Build()
	resp, err := c.client.Corehr.V2.Process.Get(ctx, req)
	if err != nil {
		c.Alert(err)
		return nil, err
	}
	if !resp.Success() {
		c.Alert(errors.New(string(resp.RawBody)))
		return nil, resp
	}
	return resp.Data, nil
}

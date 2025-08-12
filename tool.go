package lark_sdk

import (
	"encoding/json"
	larkapproval "github.com/larksuite/oapi-sdk-go/v3/service/approval/v4"
	larkcorehr "github.com/larksuite/oapi-sdk-go/v3/service/corehr/v2"
)

func ParseForm(formStr string) ([]FormWidget, error) {
	res := make([]FormWidget, 0)
	err := json.Unmarshal([]byte(formStr), &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func ParseForm2Map(formStr string) (map[string]FormWidget, error) {
	widgets, err := ParseForm(formStr)
	if err != nil {
		return nil, err
	}
	res := make(map[string]FormWidget)
	for _, widget := range widgets {
		res[widget.Name] = widget
	}
	return res, nil
}

func ParseAbstractItem(items []*larkcorehr.ProcessAbstractItem) map[string]string {
	res := make(map[string]string)
	for _, item := range items {
		res[*item.Name.ZhCn] = *item.Value.ZhCn
	}
	return res
}

func CheckNode(instInfo *larkapproval.GetInstanceRespData, nodeName string) (string, bool) {
	for _, task := range instInfo.TaskList {
		if *task.Status == larkapproval.StatusPending && *task.NodeName == nodeName {
			return *task.Id, true
		}
	}
	return "", false
}

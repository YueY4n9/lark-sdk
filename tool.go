package lark_sdk

import "encoding/json"

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

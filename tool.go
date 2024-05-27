package lark_sdk

import "encoding/json"

func parseForm(formStr string) ([]FormWidget, error) {
	res := make([]FormWidget, 0)
	err := json.Unmarshal([]byte(formStr), &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

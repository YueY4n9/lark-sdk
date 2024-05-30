package lark_sdk

type FormWidget struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Value   interface{} `json:"value"`
	Ext     interface{} `json:"ext"`
	Option  interface{} `json:"option"`
	OpenIds []string    `json:"open_ids,omitempty"`
}

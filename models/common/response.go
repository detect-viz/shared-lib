package common

// 下拉選單響應
type OptionResponse struct {
	Text  string `json:"text"  from:"text"`
	Value string `json:"value" from:"value"`
}

type Response struct {
	Msg     string
	Success bool
}

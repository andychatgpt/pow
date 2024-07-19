package models

type ParamGetPow struct {
	Seed      string `json:"seed"`
	Diff      string `json:"diff"`
	UserAgent string `json:"user_agent"`
	Proxy     string `json:"proxy"`
	Auth      string `json:"auth"`
}

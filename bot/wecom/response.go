package wecom

import "fmt"

// ResponseMeta 企业微信返回数据
type ResponseMeta struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (r *ResponseMeta) Succeed() bool {
	return r.ErrCode == 0
}

func (r *ResponseMeta) String() string {
	return fmt.Sprintf(
		"errcode=%d, errmsg=%s",
		r.ErrCode,
		r.ErrMsg,
	)
}

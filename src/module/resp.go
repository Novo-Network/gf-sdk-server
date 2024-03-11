package module

type Resp struct {
	Data string `json:"data"`
	Msg  string `json:"msg"`
	Code int32  `json:"code"`
	Id   string `json:"id"`
}

func NewResp(msg string, code int32, id string) Resp {
	return Resp{
		Data: "",
		Msg:  msg,
		Code: code,
		Id:   id,
	}
}

func NewErrResp(err error, id string) Resp {
	return Resp{
		Data: "",
		Msg:  err.Error(),
		Code: -1,
		Id:   id,
	}
}

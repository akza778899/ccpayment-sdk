package util

type AutoGenerated struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Data   `json:"data"`
}

type Data struct {
	BillID      string `json:"bill_id"`
	RedirectUrl string `json:"redirect_url"`
}

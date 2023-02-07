package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

/**

 */
func main() {
	router := gin.Default()

	// Examples of the three use scenarios of RSA encryption
	router.POST("/create/order", CreateOrder)         // Method for creating orders
	router.POST("/webhook/verify", DemoPayNotifyBack) // Example of Webhook Verification
	router.POST("/concise/url/get", GetPaymentUrl)

	// Examples of three scenes signed with SHA-256
	router.POST("/create/order/simple", CreateSimpleOrder)         // Method for creating orders
	router.POST("/webhook/verify/simple", DemoSimplePayNotifyBack) // Example of Webhook Verification
	router.POST("/concise/url/get/simple", GetSimplePaymentUrl)    // url mode

	s := &http.Server{
		Addr:           ":8089",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}

var TestCreateOrderUrl = "https://admin.ccpayment.com/ccpayment/v1/pay/CreateTokenTradeOrder"

// Method for creating orders
func CreateOrder(ctx *gin.Context) {
	bill := BillId()
	jsonContent := &JsonContent{
		TokenId:    "e8f64d3d-df5b-411d-897f-c6d8d30206b7",       // from ccpayment support token list
		Chain:      "BSC",                                        // according to user selected
		Amount:     "1",                                          // pay amount
		Contract:   "0x2170ed0880ac9a755fd29b2688956bd959f933f8", //selected token contract
		OutOrderNo: bill,                                         //merchant order id
		FiatName:   "USD",                                        //fiat name just support usd currently
	}
	//content, _ := json.Marshal(jsonContent)
	//timestamps := int64(1672261484)
	//times := strconv.Itoa(int(timestamps))
	//randStr := util.RandStr(5)
	//serviceStr := "ccpayment_id=" + mchid + "&app_id=" + arr.Appid+"&app_secret=xxxxxxx" + "&json_content=" + string(content) + "&timestamp=" + times + "&noncestr=" + randStr
	// todo 1. Concatenating signature string, Please make sure the field order
	serviceStr := "ccpayment_id=CP10001&app_id=202301170950281615285414881132544&app_secret=xxxxxxxx&json_content={\"token_id\":\"e8f64d3d-df5b-411d-897f-c6d8d30206b7\",\"chain\":\"BSC\",\"amount\":\"1\",\"contract\":\"0x2170ed0880ac9a755fd29b2688956bd959f933f8\",\"out_order_no\":\"" + bill + "\",\"fiat_name\":\"USD\"}&timestamp=1672299548&noncestr=ylaDo"
	fmt.Println(serviceStr)
	// todo 2. Use the private key for encryption
	bt, err := RsaSignWithSha256([]byte(serviceStr), []byte(PrivateKey))
	if err != nil {
		fmt.Println("Sign err:", err)
		return
	}
	req := &SubmitCreateTradeOrderRequest{
		CcpaymentId: "CP10001",
		// it can get from merchant center payment settings(web terminal), only support len(APPID) = 33
		Appid:       "202301170950281615285414881132544",
		Timestamp:   1672299548, // current time unix
		JsonContent: jsonContent,
		Sign:        bt,
		// notify url(sync notice merchant change order status) it must be set in the payment settings(web terminal), otherwise program can not work normally
		NotifyUrl: "https://admin.ccpayment.com/merchant/v1/demo/pay/notify",
		Remark:    "",
		//device type only support app currently
		Device: "web",
		//rand str
		Noncestr: "ylaDo", // Random string。util.RandStr(5)
	}
	bytes, _ := json.Marshal(req)
	// todo 3 Send an order request to CCPayment
	response, err := http.Post(TestCreateOrderUrl, "application/json", strings.NewReader(string(bytes)))
	if err != nil {
		fmt.Println("err:", err.Error())
		return
	}
	if response.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(response.Body)
		obj := AutoGenerated{}
		err = json.Unmarshal(body, &obj)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

// Example of Webhook Verification
func DemoPayNotifyBack(ctx *gin.Context) {

	encryptParam := struct {
		EncryptData []byte `json:"encrypt_data"`
	}{}
	if err := ctx.BindJSON(&encryptParam); err != nil {
		fmt.Printf("1111 %+v\n", err)
		ctx.String(http.StatusOK, "Failed")
		return
	}

	// RSA Decrypt
	decryptData, err := RsaDecrypt(encryptParam.EncryptData, []byte(PrivateKey))

	if err != nil {
		fmt.Printf("2222 %+v\n", err)
		ctx.String(http.StatusOK, "Failed")
		return
	}

	data := &EncryptData{}
	err = json.Unmarshal(decryptData, data)
	if err != nil {
		fmt.Printf("3333 %+v\n", err)
		ctx.String(http.StatusOK, "Failed")
		return
	}

	fmt.Printf("4444 %+v\n", *data.JsonContent)

	ctx.String(http.StatusOK, "Success")
	return
}

func GetPaymentUrl(ctx *gin.Context) {
	args := &GetPaymentUrlReq{}
	if err := ctx.BindJSON(args); err != nil {
		fmt.Printf("1111 %+v\n", err)
		ctx.String(http.StatusOK, "Failed")
		return
	}
	str := fmt.Sprintf("ccpayment_id=%s&app_id=%s&app_secret=%s&timestamp=%d&amount=%s&out_order_no=%s&product_name=%s&noncestr=%s",
		args.MerchantId, args.AppId, "62fbff1f796c42c50bb44d4d3d065390", args.Timestamp, args.Amount, args.OutOrderNo, args.ProductName, args.Noncestr)

	fmt.Println(str)
	dd, err := RsaSignWithSha256([]byte(str), []byte(PrivateKey))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(dd)

	bol := RsaVerySignWithSha256([]byte(str), []byte(dd), PublicKey)

	fmt.Println(bol)
	fmt.Println(err)
	ctx.String(http.StatusOK, "success")
	return
}

// Method for creating orders  -- SHA-256
func CreateSimpleOrder(ctx *gin.Context) {
	bill := BillId()
	jsonContent := &JsonContent{
		TokenId:    "e8f64d3d-df5b-411d-897f-c6d8d30206b7",       // from ccpayment support token list
		Chain:      "BSC",                                        // according to user selected
		Amount:     "1",                                          // pay amount
		Contract:   "0x2170ed0880ac9a755fd29b2688956bd959f933f8", //selected token contract
		OutOrderNo: bill,                                         //merchant order id
		FiatName:   "USD",                                        //fiat name just support usd currently
	}
	//content, _ := json.Marshal(jsonContent)
	//timestamps := int64(1672261484)
	//times := strconv.Itoa(int(timestamps))
	//randStr := util.RandStr(5)
	// todo 1. Concatenating signature string, Please make sure the field order
	// todo app_secret
	serviceStr := "ccpayment_id=CP10001&app_id=202301170950281615285414881132544&app_secret=62fbff1f796c42c50bb44d4d3d065390&json_content={\"token_id\":\"e8f64d3d-df5b-411d-897f-c6d8d30206b7\",\"chain\":\"BSC\",\"amount\":\"1\",\"contract\":\"0x2170ed0880ac9a755fd29b2688956bd959f933f8\",\"out_order_no\":\"" + bill + "\",\"fiat_name\":\"USD\"}&timestamp=1672299548&noncestr=ylaDo"
	fmt.Println(serviceStr)
	// todo 2. get hash256
	bt := Hash256([]byte(serviceStr))

	req := &SubmitCreateTradeOrderRequest{
		CcpaymentId: "CP10001",
		// it can get from merchant center payment settings(web terminal), only support len(APPID) = 33
		Appid:       "202301170950281615285414881132544",
		Timestamp:   1672299548, // current time unix
		JsonContent: jsonContent,
		Sign:        bt,
		// notify url(sync notice merchant change order status) it must be set in the payment settings(web terminal), otherwise program can not work normally
		NotifyUrl: "https://admin.ccpayment.com/merchant/v1/demo/pay/notify",
		Remark:    "",
		//device type only support app currently
		Device: "web",
		//rand str
		Noncestr: "ylaDo", // Random string。util.RandStr(5)
	}
	bytes, _ := json.Marshal(req)
	// todo 3 Send an order request to CCPayment
	response, err := http.Post(TestCreateOrderUrl, "application/json", strings.NewReader(string(bytes)))
	if err != nil {
		fmt.Println("err:", err.Error())
		return
	}
	if response.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(response.Body)
		obj := AutoGenerated{}
		err = json.Unmarshal(body, &obj)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

/**
webhook returns parameter samples
{
    "app_id": "202301170950281615285414881132544",
    "timestamp": 1675409439,
    "out_order_no": "202211181708161593531525754785795",
    "pay_status": "success",
    "sign": "871f0223c66ea72435208d03603a0cb00b90f6ac4a4ba725d00164d967e291f6",
    "noncestr": "4itw7",
    "json_content": {
        "origin_amount": "0.94382",
        "fiat_amount": "0",
        "paid_amount": "0.5",
        "current_rate": "0",
        "chain": "ETH",
        "contract": "0x7E9D8f07A64e363e97A648904a89fb4cd5fB94CD",
        "order_no": "202211181708171593531528620023808",
        "symbol": "ETH"
    }
}
*/
// Example of Webhook Verification -- SHA-256
func DemoSimplePayNotifyBack(ctx *gin.Context) {

	data := &EncryptData{}
	if err := ctx.BindJSON(&data); err != nil {
		fmt.Printf("1111 %+v\n", err)
		ctx.String(http.StatusOK, "Failed")
		return
	}

	// todo app_secret
	serviceStr := fmt.Sprintf("app_id=%s&app_secret=%s&out_order_no=%s&timestamp=%d&noncestr=%s", data.AppId, "62fbff1f796c42c50bb44d4d3d065390", data.OutOrderNo, data.Timestamp, data.Noncestr)
	fmt.Println(serviceStr)
	// todo 2. get hash256
	bt := Hash256([]byte(serviceStr))
	if data.Sign != bt { // Verification signature
		fmt.Printf("3333 sign err \n")
		ctx.String(http.StatusOK, "Failed")
		return
	}

	fmt.Printf("4444 %+v\n", *data.JsonContent)

	ctx.String(http.StatusOK, "Success")
	return
}

type GetPaymentUrlReq struct {
	MerchantId     string `json:"ccpayment_id" binding:"required"`
	AppId          string `json:"app_id" binding:"required"`
	Timestamp      uint64 `json:"timestamp" binding:"required"`
	ValidTimestamp uint64 `json:"valid_timestamp"`
	Amount         string `json:"amount" binding:"required"`
	OutOrderNo     string `json:"out_order_no" binding:"required"`
	ProductName    string `json:"product_name" binding:"required"`
	Sign           string `json:"sign" binding:"required"`
	Noncestr       string `json:"noncestr" binding:"required"`
	ReturnUrl      string `json:"return_url"`
	Uuid           string `json:"uuid"`
	Mid            int64  `json:"mid"`
	MerchantTitle  string `json:"merchant_title"`
	MerchantLogo   string `json:"merchant_logo"`
}

//  -- SHA-256
func GetSimplePaymentUrl(ctx *gin.Context) {
	args := &GetPaymentUrlReq{}
	if err := ctx.BindJSON(args); err != nil {
		fmt.Printf("1111 %+v\n", err)
		ctx.String(http.StatusOK, "Failed")
		return
	}
	str := fmt.Sprintf("ccpayment_id=%s&app_id=%s&app_secret=%s&timestamp=%d&amount=%s&out_order_no=%s&product_name=%s&noncestr=%s",
		args.MerchantId, args.AppId, "62fbff1f796c42c50bb44d4d3d065390", args.Timestamp, args.Amount, args.OutOrderNo, args.ProductName, args.Noncestr)

	fmt.Println(str)
	if Hash256([]byte(str)) != args.Sign {
		fmt.Printf("22222 %+v\n", args.Sign)
		ctx.String(http.StatusOK, "Failed")
		return
	}
	fmt.Printf("33333 %+v\n", args)
	ctx.String(http.StatusOK, "success")
	return
}

package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	//"io/ioutil"
	"net/http"
	//"strings"
	"time"
)

type TokenReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type TokenData struct {
	Token string `json:"token"`
}
type TokenResp struct {
	RtnCode string    `json:"rtnCode"`
	RtnMsg  string    `json:"rtnMsg"`
	RtnData TokenData `json:"rtnData"`
}

type OrderReq struct {
	FaceValue  string `json:"faceValue"`
	Channel    string `json:"channel"`
	Prov       string `json:"prov"`
	ReceiveNum string `json:"receiveNum"`
}

type RtnDataType struct {
	Id                int64   `json:"id"`
	AddTime           string  `json:"addtime"`
	QueryTime         string  `json:"queryTime"`
	Prov              string  `json:"prov"`
	PhoneChannel      int     `json:"phoneChannel"`
	PhoneFacevalue    float64 `json:"phoneFacevalue"`
	PhoneNo           string  `json:"phoneNo"`
	PhoneBalance      float64 `json:"phoneBalance"`
	PhoneBalanceAfter float64 `json:"phoneBalanceAfter"`
	Status            int     `json:"status"`
	Amount            float64 `json:"amout"`
	Note              string  `json:"note"`
}

type OrderResp struct {
	RtnCode string        `json:"rtnCode"`
	RtnMsg  string        `json:"rtnMsg"`
	RtnData []RtnDataType `json:"rtnData"`
}

func main() {
	faceValue := flag.String("v", "", "面值")
	channel := flag.String("c", "", "运营商渠道。1-移动，2-联通，3-电信，不填-任意")
	prov := flag.String("p", "", "省份，填写中文")
	receiveNum := flag.String("n", "", "接单数")
	flag.Parse()

	if *faceValue == "" {
		flag.PrintDefaults()
		return
	}
	orderReq := &OrderReq{*faceValue, *channel, *prov, *receiveNum}
	fmt.Println("orderReq:", *orderReq)
	token := getToken()
	tryTime := 0
	for {
		time.Sleep(time.Duration(1) * time.Second)
		orderResp := getOrder(orderReq, token)
		if orderResp == nil {
			fmt.Println("system error")
			os.Exit(1)
		}
		fmt.Println("orderResp", *orderResp)
		tryTime++
		if orderResp.RtnCode == "000000" {
			fmt.Printf("get %s faceValue success\n", *faceValue)
			return
		}
		if orderResp.RtnCode == "899991" {
			time.Sleep(time.Duration(1) * time.Second)
		}
		fmt.Printf("have tried %d times\n", tryTime)
	}

}

func getOrder(orderReq *OrderReq, token string) (oderResp *OrderResp) {
	jsonBytes, err := json.Marshal(*orderReq)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	bf := bytes.NewBuffer(jsonBytes)

	url := "http://99shou.cn/api/phonecharge-server/phonecharge/phone/receive"
	channelid := "OP0001"
	txntime := fmt.Sprintf("%v", time.Now().UnixNano()/1e6)
	secretKey := "your secretKey"
	readyStr := bf.String() + secretKey + txntime
	readyData := []byte(readyStr)
	md5Data := md5.Sum(readyData)
	sign := fmt.Sprintf("%x", md5Data)
	request, err := http.NewRequest("POST", url, bf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	request.Header.Set("channelid", channelid)
	request.Header.Set("token", token)
	request.Header.Set("txntime", txntime)
	request.Header.Set("sign", sign)
	client := http.Client{}
	resp, err := client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	orderResp := new(OrderResp)
	err = json.NewDecoder(resp.Body).Decode(orderResp)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	return orderResp

}

func getToken() (token string) {
	url := "http://99shou.cn/api/user-server/user/login"
	tokenReq := TokenReq{"username", "password"}
	bf := new(bytes.Buffer)
	err := json.NewEncoder(bf).Encode(tokenReq)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	request, err := http.NewRequest("POST", url, bf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err := client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	tokenResp := new(TokenResp)
	err = json.NewDecoder(resp.Body).Decode(tokenResp)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	return tokenResp.RtnData.Token
}

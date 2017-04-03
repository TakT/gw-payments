package main

import (
	"encoding/json"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"log"
	"strconv"
	"github.com/jinzhu/gorm"
	_ "github.com/go-sql-driver/mysql"
	"bytes"
	"encoding/base64"
)

func main() {

	r := fasthttprouter.New()

	r.GET("/check/:id", AddMiddleware(check, AuthMiddleware))
	r.POST("/check/:id/do", do)

	//h = fasthttp.CompressHandler(check)

	log.Fatal(fasthttp.ListenAndServe(":9898", r.Handler))
}

func check(ctx *fasthttp.RequestCtx) {
	idValue := ctx.UserValue("id").(string)
	id, err := strconv.Atoi(idValue)
	if err != nil {
		println("err", err.Error())
	}
	println("id", id, idValue)

	response := CheckResponse{
		Id: id,
	}
	res, err := json.Marshal(&response)
	if err != nil {
		println("json.Marshal", err.Error())
	}

	ctx.Response.Header.Add("Content-Type", "application/json; charset=utf-8")
	ctx.Response.Header.Add("Accept", "application/json")
	ctx.Response.SetBody(res)
}

func do(ctx *fasthttp.RequestCtx) {
	idValue := ctx.UserValue("id").(string)
	id, err := strconv.Atoi(idValue)
	if err != nil {
		println("err", err.Error())
	}
	println("id", id, idValue)

	var doRequest DoRequest
	body := ctx.PostBody()
	if err := json.Unmarshal(body, &doRequest); err != nil {
		println("err", err.Error())
	}

	response := DoResponse{
		State: 0,
	}
	res, err := json.Marshal(&response)
	if err != nil {
		println("json.Marshal", err.Error())
	}

	if err := doRequest.Save(); err != nil {
		println("doRequest.Save", err.Error())
	}

	xmlResponse, err := doRequest.Send()
	if err != nil {
		println("doRequest.Send", err.Error())
	}
	println("xmlResponse", xmlResponse)

	ctx.Response.Header.Add("Content-Type", "application/json; charset=utf-8")
	ctx.Response.Header.Add("Accept", "application/json")
	ctx.Response.SetBody(res)
}

type APIError struct {
	Err_code int   `json:"err_code"`
	Err      error `json:"err_text,omitempty"`
}

type CheckResponse struct {
	Id  int      `json:"id"`
	Err APIError `json:"err"`
}

type DoRequest struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

type DoResponse struct {
	State int      `json:"state"`
	Err   APIError `json:"err"`
}

func (r DoRequest) Save() error {
	db := getDB()
	defer db.Close()

	return nil
}

func (r DoRequest) Send() (string, error) {

	body := ""

	req := fasthttp.AcquireRequest()
	req.SetRequestURI("https://192.168.1.168:5858/issuingws/services/Issuing")
	req.Header.SetMethod("POST")
	req.SetBodyString(body)

	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	client.Do(req, resp)

	bodyBytes := resp.Body()
	println(string(bodyBytes))

	req.Header.Add("Content-Type", "applcation/soap+xml")
	req.Header.Add("SOAPAction", "''")

	return string(bodyBytes), nil
}

func getDB() *gorm.DB {
	db, err := gorm.Open("mysql", "u_payments:u_payments@/a_payments?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Println("failed to connect database")
	}
	return db.LogMode(true)
}

func AddMiddleware(h fasthttp.RequestHandler, middleware ...func(handler fasthttp.RequestHandler) fasthttp.RequestHandler) fasthttp.RequestHandler {
	for _, mw := range middleware {
		h = mw(h)
	}
	return h
}

var basicAuthPrefix = []byte("Basic ")

func AuthMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx){
		// Get the Basic Authentication credentials
		auth := ctx.Request.Header.Peek("Authorization")
		if bytes.HasPrefix(auth, basicAuthPrefix) {
			// Check credentials
			payload, err := base64.StdEncoding.DecodeString(string(auth[len(basicAuthPrefix):]))
			if err == nil {
				user := []byte("test")
				password:= []byte("test")
				pair := bytes.SplitN(payload, []byte(":"), 2)
				if len(pair) == 2 &&
					bytes.Equal(pair[0], user) &&
					bytes.Equal(pair[1], password) {
					// Delegate request to the given handle
					next(ctx)
					return
				}
			}
		}

		// Request Basic Authentication otherwise
		ctx.Response.Header.Set("WWW-Authenticate", "Basic realm=Restricted")
		ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
	})
}
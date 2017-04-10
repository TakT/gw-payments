package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"github.com/buaazp/fasthttprouter"
	"github.com/jinzhu/gorm"
	"github.com/valyala/fasthttp"
	"log"
	"strconv"
) // подгрузка библиотек

func main() {

	r := fasthttprouter.New() // переменная для https запроса

	/*baseAuth := base64.StdEncoding.EncodeToString([]byte("Basic:dGVzdDp0ZXN0"))
	println("BaseAuth", baseAuth)*/

	r.GET("/check/:id", AddMiddleware(check, AuthMiddleware, HeadersMiddleware))  // получить данные для проверки
	r.POST("/check/:id/do", AddMiddleware(do, AuthMiddleware, HeadersMiddleware)) // выполнить действие

	//h = fasthttp.CompressHandler(check)

	log.Fatal(fasthttp.ListenAndServe(":9898", r.Handler)) // если ошибка
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

	ctx.Response.SetBody(res)
} // функция проверки, отправляет запрос по https, получает значение id выводит его на экран

// функция на исполнение
func do(ctx *fasthttp.RequestCtx) {

	idValue := ctx.UserValue("id").(string) // Переменная, в которой значение типа строка. Закладывем занчение контекста
	id, err := strconv.Atoi(idValue)        // В переменные присваимваем значение idValue - Конвертация с строки в число
	if err != nil {
		println("err", err.Error()) // Если ошибка, то вывести на экран
	}
	println("id", id, idValue) // а так выводит значение

	var doRequest DoRequest // переменная типа doRequest
	body := ctx.PostBody()  // переменная body помещаем туда контекст
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

	var responseStructure Catalog
	xmlResponse, err := doRequest.Send() // отправляет ответ запроса
	if err != nil {
		println("doRequest.Send", err.Error())
	}

	if err := xml.Unmarshal([]byte(xmlResponse), &responseStructure); err != nil {
		println("xml.Unmarshal", err.Error())
	}

	for i, item := range responseStructure.CD {
		println(i, item.Title, item.Artist, item.Country, item.Company)
	}

	//println("xmlResponse", xmlResponse) // выводит ответ в xml

	ctx.Response.SetBody(res) // формируем тело
}

type CatalogItem struct {
	Title   string `xml:"TITLE"`
	Artist  string `xml:"ARTIST"`
	Country string `xml:"COUNTRY"`
	Company string `xml:"COMPANY"`
}

type Catalog struct {
	CD []CatalogItem `xml:"CD"`
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
	if db != nil {
		defer db.Close()
	}

	return nil
}

func (r DoRequest) Send() (string, error) {

	body := ""
	req := fasthttp.AcquireRequest() // запрос

	req.SetRequestURI("https://www.w3schools.com/xml/cd_catalog.xml") //???
	req.Header.SetMethod("GET")                                       // запрос ПОСТ/GET
	req.SetBodyString(body)                                           // задать тело запроса

	resp := fasthttp.AcquireResponse() // ответ

	req.Header.Add("Content-Type", "applcation/soap+xml")
	req.Header.Add("SOAPAction", "''")

	client := &fasthttp.Client{} // создаем клиента, надо тело
	client.Do(req, resp)         // выполнить запрос ответ

	bodyBytes := resp.Body()
	println("resp.StatusCode()", resp.StatusCode())

	return string(bodyBytes), nil
}

func getDB() *gorm.DB {
	db, err := gorm.Open("postgres", "host=localhost user=asel1 password=Asel1 dbname=ibskg sslmode=disable")

	if err != nil {
		log.Println("failed to connect database")
		return nil
	}
	return db.LogMode(true)
}
/*
func AddMiddleware(h fasthttp.RequestHandler, middleware ...func(handler fasthttp.RequestHandler) fasthttp.RequestHandler) fasthttp.RequestHandler {
	for _, mw := range middleware {
		h = mw(h)
	}
	return h
}*/
/*
// что это? авторизация, но куда?
var basicAuthPrefix = []byte("Basic")

func AuthMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		// Get the Basic Authentication credentials
		auth := ctx.Request.Header.Peek("Authorization")
		if bytes.HasPrefix(auth, basicAuthPrefix) {
			// Check credentials
			payload, err := base64.StdEncoding.DecodeString(string(auth[len(basicAuthPrefix):]))
			if err == nil {
				user := []byte("Basic")
				password := []byte("dGVzdDp0ZXN0")
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
	})*/
}

func HeadersMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("application/json; charset=utf-8")
		next(ctx)
	})
}

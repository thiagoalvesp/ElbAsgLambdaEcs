package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
)

type MessageResponse struct {
	Counter int64
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("loggingMiddleware")
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func CounterHandler(w http.ResponseWriter, r *http.Request) {

	json := simplejson.New()
	json.Set("Counter", 30)

	payload, err := json.MarshalJSON()
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)

}

func AddCounterHandler(w http.ResponseWriter, r *http.Request) {

	json := simplejson.New()
	json.Set("Counter", 10)

	payload, err := json.MarshalJSON()
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)

}

func main() {

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Not found", r.RequestURI)
		http.Error(w, fmt.Sprintf("Not found: %s", r.RequestURI), http.StatusNotFound)
	})

	r.HandleFunc("/add", AddCounterHandler)
	r.HandleFunc("/", CounterHandler)

	r.Use(loggingMiddleware)

	if runtime_api, _ := os.LookupEnv("AWS_LAMBDA_RUNTIME_API"); runtime_api != "" {
		log.Println("Starting up in Lambda Runtime")
		adapter := gorillamux.NewALB(r)
		lambda.Start(adapter.ProxyWithContext)
	} else {
		log.Println("Starting up on own")
		srv := &http.Server{
			Addr:    ":8080",
			Handler: r,
		}
		_ = srv.ListenAndServe()
	}

}

// func handler(request events.ALBTargetGroupRequest) (
// 	events.ALBTargetGroupResponse, error) {

// 	fmt.Println("Start Proccess")

// 	fmt.Println("Body")
// 	fmt.Println(request.Body)

// 	messageResponse := MessageResponse{Counter: 0}
// 	jsonMessageResponse, err := json.Marshal(messageResponse)
// 	jsonMessageResponseString := string(jsonMessageResponse)

// 	if err != nil {
// 		fmt.Println("Erro ao codificar para JSON:", err)
// 		response := events.ALBTargetGroupResponse{
// 			StatusCode: 500,
// 		}
// 		return response, err
// 	}

// 	headers := make(map[string]string)
// 	headers["Content-Type"] = "application/json"

// 	response := events.ALBTargetGroupResponse{
// 		StatusCode:      200,
// 		Body:            jsonMessageResponseString,
// 		IsBase64Encoded: false,
// 		Headers:         headers,
// 	}

// 	return response, nil
// }

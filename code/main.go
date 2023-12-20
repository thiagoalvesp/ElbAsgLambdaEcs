package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

func main() {

	log.Printf("Gin cold start")
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		log.Printf("chegou no healthy")
		c.JSON(200, gin.H{
			"message": "healthy",
		})
	})
	r.GET("/ping", func(c *gin.Context) {
		log.Printf("chegou no ping")
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/pong", func(c *gin.Context) {
		log.Printf("chegou no pong")
		c.JSON(200, gin.H{
			"message": "ping",
		})
	})

	if runtime_api, _ := os.LookupEnv("AWS_LAMBDA_RUNTIME_API"); runtime_api != "" {
		log.Println("Starting up in Lambda Runtime gin")
		ginLambda := ginadapter.NewALB(r)
		lambda.Start(func(ctx context.Context, req events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
			log.Printf(req.Path)
			evalbresponse, _ := ginLambda.ProxyWithContext(ctx, req)
			//headers vazio da erro no alb
			headers := make(map[string]string)
			headers["Content-Type"] = "application/json"
			evalbresponse.Headers = headers
			return evalbresponse, nil

		})
	} else {
		log.Println("Starting up on own")
		r.Run()
	}

}

//to do
// estudar os meios de balanceamento

// configurar o gateway para bater no alb via vpc link
//criar o vpc link
//api rest
//atribuir para o vpc
// integracao com o loadbalancer
// atribuir a lambda para vpc *

// jmeter
// criar um redirecionamento para o ecs
// fazer a app em container
// subir o ecs
// configurar os eventos para desligar ou ligar o ecs pela metrica da lambda

// cloudformation
// documentação

//http://meuloadbalancer-1598794745.sa-east-1.elb.amazonaws.com/

//go melhorar o código e fazer teste unitário

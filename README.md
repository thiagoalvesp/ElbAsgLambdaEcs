# Utilizando Lambda com redundância no ECS fargate 

O que você vai encontrar nesse material:
- Objetivo
- Desenho de Solução
- Componentes necessários na AWS
- Código da aplicação
- Deploy Lambda
- Deploy Imagem Docker
- Configuração ECS
- Configuração ALB para Lambda
- Configuração Cloud Watch Alarm
- Configuracão Event Bridge
- Criacão da lambda para provisionamento e ajuste no ALB
- Teste no Jmeter
- Conclusão
- Fontes

### Objetivo

O objetivo dessa prova conceito é mostrar que podemos ter um ambiente com lambdas (serveless) para hospedar uma web api sem se preocupar com os hard limits pois podemos usar ecs fargate compartilhando a mesma base de código da aplicação como redundância. 

### Desenho de Solução
![image](https://github.com/thiagoalvesp/ElbAsgLambdaEcs/assets/10868308/0abe2073-ac9e-4c53-a9c0-c81a91dde261)

### Componentes necessários na AWS

Para essa prova de conceito é necessário: 
- Conta aws (free tier)
- ECS
- Lambda
- Cloud Watch Alarm
- Event Bridge
- ALB
- GO Lang
- Python

### Código da aplicação

O código foi customizado para rodar tanto em servidores web tradicionais (containers) quanto em ambiente AWS Lambda, para isso utilizamos o adapter aws-lambda-go-api-proxy e o gin com framework para cria as rotas de API.
Utilizamos a variavel de ambiente AWS_LAMBDA_RUNTIME_API para determinar se a aplicação está rodando em lambdas ou ambiente de container.


```go
func main() {

	log.Printf("Gin cold start")
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		log.Printf("chegou no healthy")

		time.Sleep(15 * time.Second)

		c.JSON(200, gin.H{
			"message": "healthy",
		})
	})
	r.GET("/bang", func(c *gin.Context) {
		log.Printf("chegou no bang")
		c.JSON(200, gin.H{
			"message": "boom",
		})
	})
	r.GET("/pong", func(c *gin.Context) {
		log.Printf("chegou no pong")
		c.JSON(200, gin.H{
			"message": "ping",
		})
	})

	r.GET("/sleep", func(c *gin.Context) {
		log.Printf("chegou no sleep")

		time.Sleep(15 * time.Second)

		c.JSON(200, gin.H{
			"message": "awaken",
		})
	})

	r.GET("/env", func(c *gin.Context) {
		log.Printf("chegou no sleep")
		if runtime_api, _ := os.LookupEnv("AWS_LAMBDA_RUNTIME_API"); runtime_api != "" {
			c.JSON(200, gin.H{
				"message": "lambda",
			})
		} else {
			c.JSON(200, gin.H{
				"message": "server",
			})
		}
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
````

### Deploy Lambda

Para fazer o deploy podemos utilizar a extensão da propria IDE para agilizar o processo.

![image](https://github.com/thiagoalvesp/ElbAsgLambdaEcs/assets/10868308/bbf5af33-c27d-4f90-84a1-466037b9431c)

 

### Conclusão

Fica claro que podemos utilizar o ecs fargate para suportar as requisições extras da lambda.
    

### Fontes

- https://www.youtube.com/watch?v=9bjBOOfPtRk&t=573s
- https://docs.aws.amazon.com/elasticloadbalancing/latest/application/introduction.html
- https://boto3.amazonaws.com/v1/documentation/api/latest/guide/examples.html
- https://github.com/aws/aws-lambda-go

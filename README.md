# Utilizando Lambda com redundância no ECS fargate 

O que você vai encontrar nesse material:
- Objetivo
- Desenho de Solução
- Componentes necessários na AWS
- Código da aplicação
- Deploy Lambda
- Deploy Imagem Docker no ECR
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
- ECR
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

### Deploy Imagem Docker no ECR

Para subir a imagem no ECR precisamos previamente construir nossa imagem localmente para isso utilizamos esse Dockerfile
```Dockerfile
FROM golang:1.20-alpine
WORKDIR /code
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o ./out/dist
CMD ./out/dist
```

Depois podemos utilizar o AWS CLI para fazer o push 

```bash
#login no ECR utilizando as credencias do cli
aws ecr get-login-password --region sa-east-1 | docker login --username AWS --password-stdin 281303628498.dkr.ecr.sa-east-1.amazonaws.com
#build da imagem
docker build -t golangapppbangpong .
#tag antes do push
docker tag golangapppbangpong:latest 281303628498.dkr.ecr.sa-east-1.amazonaws.com/golangapppbangpong:latest
#push para o ECR
docker push 281303628498.dkr.ecr.sa-east-1.amazonaws.com/golangapppbangpong:latest
``` 
### Configuração ECS

Primeiro precisamos criar o cluster e esse passo não tem segredo utilizando a console da aws.
Para essa prova de conceito utilizamos o provedor fargate para subir nosso workload.

![image](https://github.com/thiagoalvesp/ElbAsgLambdaEcs/assets/10868308/46d6c2eb-0373-4142-9f3c-f03ecf048df2)

Segundo passo é criar uma definição de tarefa. Podemos utilizar a console da aws ou subir um json como o do exemplo abaixo.

```json
{
    "taskDefinitionArn": "arn:aws:ecs:sa-east-1:281303628498:task-definition/golangapptaskdefinition:1",
    "containerDefinitions": [
        {
            "name": "goapp",
            "image": "281303628498.dkr.ecr.sa-east-1.amazonaws.com/golangapppbangpong",
            "cpu": 0,
            "portMappings": [
                {
                    "name": "goapp-8080-tcp",
                    "containerPort": 8080,
                    "hostPort": 8080,
                    "protocol": "tcp",
                    "appProtocol": "http"
                }
            ],
            "essential": true,
            "environment": [],
            "environmentFiles": [],
            "mountPoints": [],
            "volumesFrom": [],
            "ulimits": [],
            "logConfiguration": {
                "logDriver": "awslogs",
                "options": {
                    "awslogs-create-group": "true",
                    "awslogs-group": "/ecs/golangapptaskdefinition",
                    "awslogs-region": "sa-east-1",
                    "awslogs-stream-prefix": "ecs"
                },
                "secretOptions": []
            }
        }
    ],
    "family": "golangapptaskdefinition",
    "executionRoleArn": "arn:aws:iam::281303628498:role/ecsTaskExecutionRole",
    "networkMode": "awsvpc",
    "revision": 1,
    "volumes": [],
    "status": "ACTIVE",
    "requiresAttributes": [
        {
            "name": "com.amazonaws.ecs.capability.logging-driver.awslogs"
        },
        {
            "name": "ecs.capability.execution-role-awslogs"
        },
        {
            "name": "com.amazonaws.ecs.capability.ecr-auth"
        },
        {
            "name": "com.amazonaws.ecs.capability.docker-remote-api.1.19"
        },
        {
            "name": "ecs.capability.execution-role-ecr-pull"
        },
        {
            "name": "com.amazonaws.ecs.capability.docker-remote-api.1.18"
        },
        {
            "name": "ecs.capability.task-eni"
        },
        {
            "name": "com.amazonaws.ecs.capability.docker-remote-api.1.29"
        }
    ],
    "placementConstraints": [],
    "compatibilities": [
        "EC2",
        "FARGATE"
    ],
    "requiresCompatibilities": [
        "FARGATE"
    ],
    "cpu": "256",
    "memory": "512",
    "runtimePlatform": {
        "cpuArchitecture": "X86_64",
        "operatingSystemFamily": "LINUX"
    },
    "registeredAt": "2023-12-21T00:14:24.574Z",
    "registeredBy": "arn:aws:iam::281303628498:root",
    "tags": []
}
```
Terceiro passo é criar o serviço para instanciar nossa aplicação no ECS.
Recomendo para criar o Aplication Load Balancer junto com o serviço pois a AWS gerencia o Target Group de forma automatica, se for criado separado precisamos fazer a gestão do ip para cada nova tarefa que é criada.

![image](https://github.com/thiagoalvesp/ElbAsgLambdaEcs/assets/10868308/d661ba0e-52a4-4bd1-972f-414504b72ada)

Nesse estágio estamos com a aplicação publicada na lambda e no ECS, porém o ecs não possui containers rodando pois colocamos as Tarefas desejadas como 0.

### Configuração ALB para Lambda
Como criamos o ALB junto com o ECS, agora precisamos criar um target group para lambda para ser atribuido ao listener do ALB.

![image](https://github.com/thiagoalvesp/ElbAsgLambdaEcs/assets/10868308/e0120a52-c8ff-4a36-9bc7-e063c3b9993d)

![image](https://github.com/thiagoalvesp/ElbAsgLambdaEcs/assets/10868308/b687d203-50a8-48bc-aba3-8f68c72f66bd)

Nesse ponto a load balancer vai direcionar as requisições para lambda por conta do peso.

O peso funciona da seguinte forma quando estiver 0 o load balancer vai ignorar aquele target group se ambos estiverem com 1 as requisições serão dividas 50%/50%.

### Configuração Cloud Watch Alarm


### Conclusão

Fica claro que podemos utilizar o ecs fargate para suportar as requisições extras da lambda.
    

### Fontes

- https://www.youtube.com/watch?v=9bjBOOfPtRk&t=573s
- https://docs.aws.amazon.com/elasticloadbalancing/latest/application/introduction.html
- https://boto3.amazonaws.com/v1/documentation/api/latest/guide/examples.html
- https://github.com/aws/aws-lambda-go

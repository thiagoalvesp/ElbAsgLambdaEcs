# Utilizando Lambda com redundância no ECS fargate 

O que você vai encontrar nesse material:
- Objetivo
- Desenho de Solução
  

- Diferença entre a url pré assinada Cloudfront e a url pré assinada do S3
- Componentes necessários na AWS
- Criando o Bucket S3
- Criando o CloudFront
- Criando a chave rsa publica e privada 
- Configurando a chave no CloudFront
- Diferença entre CannedSignedURL e CustomSignedURL
- Utilizando o SDK AWS .Net
- Implementação do CannedSignedURL
- Implementação do CustomSignedURL

- Conclusão
- Fontes

### Objetivo

O objetivo dessa prova conceito é mostrar que podemos ter um ambiente com lambdas (serveless) sem se preocupar com os hard limits da aws sendo apoiado pelo ecs fargate compartilhando a mesma base de código da aplicação. 

### Desenho de Solução
![image](https://github.com/thiagoalvesp/ElbAsgLambdaEcs/assets/10868308/0abe2073-ac9e-4c53-a9c0-c81a91dde261)


### Diferença entre a url pré assinada Cloudfront e a url pré assinada do S3

Antes de iniciar é valido lembra que o S3 possui uma funcionalidade parecida que é focada nos objetos de forma granular(criar, alterar e download) enquanto o CloudFront foca na distribuição de conteúdo não se limitando ao S3 como origem.

### Componentes necessários na AWS

Para essa prova de conceito é necessário: 
- Conta aws (free tier)
- Configurar uma distribuição CloudFront
- Configurar um Bucket S3
- Upload Arquivo de teste (.jpg)
- Aplicação console (.net core 3.1) e o pacote AWSSDK.CloudFront.

Agora que já sabemos o que precisa ser feito, mão na massa!!!

### Criando o Bucket S3

Para criar o bucket na aws podemos utilizar a console (https://s3.console.aws.amazon.com/s3/bucket/create?region=sa-east-1) ou utilizando CLI. 
```bash
aws s3api create-bucket --bucket <NOME_DO_BUCKET> --region <NOME_DA_REGIÃO>
```

Para o upload podemos utilizar o CLI ou a console.
```
aws s3 cp <CAMINHO_DO_ARQUIVO_LOCAL> s3://<NOME_DO_BUCKET>/<CAMINHO_NO_BUCKET>
```

Observação: Foi utilizando a criptografia SSE-S3 e todos os objetos do bucket estão com acesso bloqueado ao público. Podemos disponibilizar arquivos de outras origens além do S3, por exemplo EC2 ou ECS.

### Criando o CloudFront

Via CLI temos que criar um json com as especificações. conforme o exemplo abaixo.
```json
{
  "Comment": "Minha distribuição do CloudFront",
  "Origins": {
    "Quantity": 1,
    "Items": [
      {
        "Id": "minha-origem-s3",
        "DomainName": "<ORIGEM_S3_URL>",
        "S3OriginConfig": {
          "OriginAccessIdentity": ""
        }
      }
    ]
  },
  "DefaultCacheBehavior": {
    "TargetOriginId": "minha-origem-s3",
    "ViewerProtocolPolicy": "redirect-to-https",
    "DefaultTTL": 86400
  },
  "Enabled": true,
  "DefaultRootObject": "index.html",
  "PriceClass": "PriceClass_100",
  "ViewerCertificate": {
    "CloudFrontDefaultCertificate": true
  },
  "AllowedMethods": {
    "Quantity": 2,
    "Items": ["GET", "HEAD"]
  },
  "Compress": false,
  "SmoothStreaming": false,
  "Restrictions": {
    "GeoRestriction": {
      "RestrictionType": "whitelist",
      "Quantity": 2,
      "Items": ["US", "CA"]
    }
  }
}
```
Depois executar o comando.
```bash
aws cloudfront create-distribution --distribution-config file://<NOME_DA_CONFIG_JSON> --output json > distribution-output.json
```

Para facilitar o entendimento farei o passo a passo na console.

Escolhendo o S3 como origem.
![image](https://github.com/thiagoalvesp/CloudfrontSignedUrl/assets/10868308/aa957850-57f9-4251-9183-0211a7ecf68b)

Restringindo o acesso do bucket somente para o CloudFront

![image](https://github.com/thiagoalvesp/CloudfrontSignedUrl/assets/10868308/5ad9607f-415c-49f9-ace7-a9f82141a2fc)

Para evitar custos não é necessário habilitar o Shield, WAF, Standard logging e o price class utilizar a opção *Use only North America and Europe*.

Também é necessário criar uma policy do bucket para liberar o uso no cloudfront.
```json
{
    "Version": "2008-10-17",
    "Id": "PolicyForCloudFrontPrivateContent",
    "Statement": [
        {
            "Sid": "AllowCloudFrontServicePrincipal",
            "Effect": "Allow",
            "Principal": {
                "Service": "cloudfront.amazonaws.com"
            },
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::[nome_bucket]/*",
            "Condition": {
                "StringEquals": {
                    "AWS:SourceArn": "arn:aws:cloudfront::[id_conta]:distribution/[id_distribution]"
                }
            }
        }
    ]
}
```

### Criando a chave rsa publica e privada e exemplos

Para gerar as chaves publica e privada devemos executar o seguinte comando no bash:
```bash
openssl rsa -pubout -int cloudfront-test-key.pem -out cloudfront-test-key.pub
```
Exemplos do output: 
```
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyU2eOqzOjeu5TxNShbRc
XV3wshxSxo0Fk7GZvNdb33dXb9VC8c4vVrnByp7ET7H1a8OqRqZU7B8c9chSOVpP
a0NVhrV7PZEy7ukk6ks/Ch4iuSf+/Zfzu90nB/BTU7UJE3oI0rZ2fnkDd2Xes6wE
9IKSSGfa6NbIUK+0aWwg8Y2jxUR7wxDYT2R+7NWwqPb5aPc08VmzScBDGgdhLnVl
xSk3DT1ArQZfAjEHkLTPxe/GEDitCmHDLoBMvQwe9kPQ8RDXstPUuG7Z/AA+zLDD
xaubvPQVxDw8RreqlOOlPI/Q/E7SroBjbOBVFagZNF9Ehn3Bilf7QPBAjcQ1caoh
FwIDAQAB
-----END PUBLIC KEY-----
```

```
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAyU2eOqzOjeu5TxNShbRcXV3wshxSxo0Fk7GZvNdb33dXb9VC
8c4vVrnByp7ET7H1a8OqRqZU7B8c9chSOVpPa0NVhrV7PZEy7ukk6ks/Ch4iuSf+
/Zfzu90nB/BTU7UJE3oI0rZ2fnkDd2Xes6wE9IKSSGfa6NbIUK+0aWwg8Y2jxUR7
wxDYT2R+7NWwqPb5aPc08VmzScBDGgdhLnVlxSk3DT1ArQZfAjEHkLTPxe/GEDit
CmHDLoBMvQwe9kPQ8RDXstPUuG7Z/AA+zLDDxaubvPQVxDw8RreqlOOlPI/Q/E7S
roBjbOBVFagZNF9Ehn3Bilf7QPBAjcQ1caohFwIDAQABAoIBAQCm2vFWjTpApKzb
AJccQF12/pCt8ZAjB20h+MoHnzKFzfPpvIlayJ6wchRRkLwDmuxkQLD5EpG9jiSB
DWQqKdM+g3d2yyK1646ePR3eHjTIfCYn9yECrlrW0v6xM+C9t4coX7TEg31AY6od
45BuuRz6VuhNn9fxu2YaiyktYPUFgfjZGE0Vf80SId10vEaNDLQDlY442648DMrM
S5BHGg1s0ufcXx8cBr/FHmz/32pn6xR5OHWqbb3x/I7QMV/MtslFUnl0Efv6UxTn
US9BTMxocnzFVSHTYZJ2nkjz8SPbEgkh2aepD9UU+vHIMghWRqLf3icfZ/fplhkR
yHQJ+PAxAoGBAP9NZO4DVU/q9As7Cr7MZXZsz/ZMyrJqTqXpeA3QGTBjiW2dKYOJ
tT8nwilRx2nlNNkbhLRLU5nknvldo2oV5Bh/Wpb6VzeFrzsa9I79j9E2l3c973c8
hHVNw0eFr9JSLOgnzHWxhvNXQp/zAlAAaTpx4qq/0NRRoJ3vTMzrSjbzAoGBAMna
cmOoFvvLt5FOT2k3kq9Pim/vYH5ydFfmOvxN3UkE/bAMrGl0dbRz7D4eS4lNGr27
2tGIcEk913teusDuVHnoLfDWvpwQ1NTFXPoicqqBlj5NrVrE91eQlcJZDvjVHTCT
bIleUi+CDcLYABF0qTo0Hn3ZZVjJhmvZahTwhH5NAoGAUHli4SunzqMu/gNEZdQj
/2pZOzgFhKvB0sZ/E0uPRRN7FFQ/67iSqy+rIj8m7phTSkREVliQJ6hK/CuqARyZ
Y6dxNLoAl/3JuIXMpO4EUVw17l5Vh25KCnfSoE7hlxhUE3HIHykwcrAEzkpZZkJa
6RNQ8aW4+9QnHuF5gfaA1EUCgYAiIeg55dCNH3OZBI71EcqiDmcwal/8wcnemzXa
OCh1Enz7agk1g9Xrf7axAlpvizQ8ZSmpSNMD74sid3BI84QhYRtzoDx3E3mJyR3h
xjVxk5weSPBJawkQK4jHZlvbw929uxAdYm+vTOSaz/+i9AExsGJ/kWVL0DgEwKzp
gYpF+QKBgCmLiXmMs8obISwK3h5cWrkEGBeJyt2BFDliddAxYGewnjkuFHO5jkje
zmkYyDzFHNrwgD3TX0DL2pYkaZB2ejs93QHWLDw8WWpXsrnbzjXjyE3ZSVkhjX3n
qx7d7LsTY4Z4oY3C/4kCc+eoaplREtd7ImsNiPCjtr7u9O5BOwzE
-----END RSA PRIVATE KEY-----
```

Não utilizar as chaves de exemplo!!!

### Configurando a chave no CloudFront

Primeiro precisamos registrar a nossa chave publica no Cloudfront.

![image](https://github.com/thiagoalvesp/CloudfrontSignedUrl/assets/10868308/22e312ea-4882-48f7-bfed-fbe8a859ae3f)

Depois precisamos atribuir essa chave a um grupo.

![image](https://github.com/thiagoalvesp/CloudfrontSignedUrl/assets/10868308/911a4ce0-fdb1-41b1-8e23-781aef4eca50)

E por fim atribuir esse grupo as restrições de acesso do Cloudfront nos comportamentos.

![image](https://github.com/thiagoalvesp/CloudfrontSignedUrl/assets/10868308/334c126b-bbc0-486d-908c-b60c354333cb)

Existem a opção de atribuir a chave na conta IAM porém não é recomendado.

### Diferença entre CannedSignedURL e CustomSignedURL


![image](https://github.com/thiagoalvesp/CloudfrontSignedUrl/assets/10868308/6037d9d2-5874-470b-b0c7-7a4b02c467dc)

A Abordagem custom fornece mais opções de configurações, como por exemplo inicio do uso da url e restrição por ip, além de ser possivel utilizar a mesma url para diversos arquivos.

A url gerada é maior pois leva as policy em base64, conforme podemos ver abaixo:

`https://d1w1uj9kqe1ebe.cloudfront.net/narutoperfil.jpg?Expires=1694740255&Signature=FnhGvpj-zs11d~V2b8X82~uh09rrHnSuSNvfmrpXrk5ms7ulWvXpwp2x7vyLE5kd34Pq9mle5JNGb~i2XpjrqunwITMeogCSyAscdwYZWS0sSKztNs48cmpHP~fk5Zw1fpnE-2H3A6Q63htvaFw-ujquN-hKiS2vfSlqWDC0fO7R03yGfwo1tGdvxUufT6NC55BpsyHUSepDGI5VDx4a7Cyo~hBuW3m~0OCO~MMvpj6G-RTY0qYjWj2AgfmeRvkVZy~o6Z2McwTkaxk~lojWp-ZR5o5kJjYgk2gu9Q2bDqv1ntaTSn7EKPZL3a07yWHPJz2ShbZ~~Q0wYEnhCTeV4g__&Key-Pair-Id=K2FBX61W5Y2MOL`

### Criando a aplicação 

Para criar a aplicação podemos executar o seguinte comando:
```
dotnet new console --framework netcoreapp3.1
```

A chave privada (arquivo .pem) deve estar acessivel para aplicação.

### Utilizando o SDK AWS .Net

Para interagir com o CloudFront podemos utilizar o SDK AWS, para instalar o pacote basta executar o seguinte comando:
```
dotnet add package AWSSDK.CloudFront --version 3.7.201.32
```

### Implementação do CannedSignedURL
```c#
var protocol = AmazonCloudFrontUrlSigner.Protocol.https;
var distributionDomain =  "d1w1uj9kqe1ebe.cloudfront.net";
var resourcePath = "narutoperfil.jpg";
var keyPairId = "K2FBX61W5Y2MOL";
var expiresOn = DateTime.Now.AddSeconds(45);

using (StreamReader reader = File.OpenText(System.IO.Path.GetFullPath("cloudfront-test-key.pem")))
{
       var text =   AmazonCloudFrontUrlSigner.GetCannedSignedURL(protocol, distributionDomain, reader, resourcePath, keyPairId, expiresOn);
       Console.WriteLine("CannedSignedURL");
       Console.WriteLine(text);
       Console.WriteLine("----------------------------------");
}
```

### Implementação do CustomSignedURL
```c#
var protocol = AmazonCloudFrontUrlSigner.Protocol.https;
var distributionDomain =  "d1w1uj9kqe1ebe.cloudfront.net";
var resourcePath = "narutoperfil.jpg";
var keyPairId = "K2FBX61W5Y2MOL";
var expiresOn = DateTime.Now.AddSeconds(45);
var ipRange = "2804:1b3:a200:9dc7:cfc:d226:d38:7dff";


using (StreamReader reader = File.OpenText(System.IO.Path.GetFullPath("cloudfront-test-key.pem")))
{
       var text =   AmazonCloudFrontUrlSigner.GetCustomSignedURL(protocol,distributionDomain, reader, resourcePath, keyPairId, expiresOn, ipRange);
       Console.WriteLine("CustomSignedURL");
       Console.WriteLine(text);
       Console.WriteLine("----------------------------------");
}
```

### Conclusão

Fica claro que com poucas linhas de código e algumas configurações podemos implementar uma camada de acesso/segurança sofisticada no CloudFront.
    

### Fontes

- https://www.udemy.com/course/aws-certified-developer-associate-dva-c01
- https://www.youtube.com/watch?v=NTOCzsn7b4A
- https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-signed-urls.html

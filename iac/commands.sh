#create an s3 bucket
aws s3 mb s3://simpleapi-code

#package cloudformation
aws cloudformation package --s3-bucket simpleapi-code --template-file template.yaml --output-template-file gen/template-generated.yaml 
sam package --template-file template.yaml --s3-bucket simpleapi-code --output-template-file gen/template-generated.yaml 

# deploy
aws cloudformation deploy --template-file C:\Users\thiag\Desktop\ElbAsgLambdaEcs\iac\gen\template-generated.yaml --stack-name simpleapi-go --capabilities CAPABILITY_IAM
sam deploy --template-file C:\Users\thiag\Desktop\ElbAsgLambdaEcs\iac\gen\template-generated.yaml --stack-name simpleapi-go120


sam build

sam package --template-file .aws-sam/build/template.yaml --s3-bucket simpleapi-code --output-template-file packaged-template.yaml --debug 

sam deploy --template-file C:\Users\thiag\Desktop\ElbAsgLambdaEcs\iac\packaged-template.yaml --stack-name simpleapi-go120 --debug 


#---- cloudformation
aws cloudformation create-stack --stack-name SimpleApiGo --template-body template.yaml

aws cloudformation deploy --template-file template.yaml --stack-name SimpleApiGo --region sa-east-1 --capabilities CAPABILITY_NAMED_IAM


aws cloudformation update-stack --stack-name SimpleApiGo --template-body template.yaml

aws cloudformation delete-stack --stack-name aws-sam-cli-managed-default

#----- go build linux
set GOOS=linux
set GOOS=windows
go build .

#---- s3 font 
aws s3 mb s3://simpleapi-go-code

powershell
Compress-Archive -Force -Path "C:\Users\thiag\Desktop\ElbAsgLambdaEcs\code\simpleapi" -DestinationPath "C:\Users\thiag\Desktop\ElbAsgLambdaEcs\iac\simpleapi-go-code-files.zip"
exit

aws s3 cp C:\Users\thiag\Desktop\ElbAsgLambdaEcs\iac\simpleapi-go-code-files.zip s3://simpleapi-go-code
import boto3

client = boto3.client('ecs')

cluster = 'AlbECSCluster'
service = 'servicegoapp'

response = client.update_service(cluster=cluster, service=service, desiredCount=0)

print(response)
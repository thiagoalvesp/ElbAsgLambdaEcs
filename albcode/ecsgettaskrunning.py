import boto3

client = boto3.client('ecs')

response = client.list_tasks(
    cluster='AlbECSCluster',
)

taskarns = response['taskArns']
describe_tasks_response = client.describe_tasks(cluster='AlbECSCluster',tasks=taskarns)

#Espero o Ecs ligar
container_RUNNING = False
while container_RUNNING == False:
    for t in describe_tasks_response['tasks']:
        for c in t['containers']:
            if c['lastStatus'] == 'RUNNING' : 
                print('container RUNNING')
                container_RUNNING = True



import json
import boto3

#somente para efeitos de poc o ideal seria mapear os eventos separadamente adicionar um para quando o ecs ligar

# ecs config - pegar do env
ECSclient = boto3.client('ecs')
cluster = 'AlbECSCluster'
service = 'goappservice'
# elb listener config 
ELBclient = boto3.client('elbv2')  # Criando cliente para o Elastic Load Balancing
listenerArn = 'arn:aws:elasticloadbalancing:sa-east-1:281303628498:listener/app/AlbPocGoApp/e8477991703569a5/c95458dc9d6e2c4a'
port = 80
targetGroupLambda = 'arn:aws:elasticloadbalancing:sa-east-1:281303628498:targetgroup/GrupoDestinoLambda/07dbbd92b1ef7d80'
targetGroupECS = 'arn:aws:elasticloadbalancing:sa-east-1:281303628498:targetgroup/goappservicetargetgroup/cf082f1fd5baa9e8'


def lambda_handler(event, context):
    
    
    previousState = event['detail']['previousState']['value']
    state = event['detail']['state']['value']
    
    print(state)
    
    if state == 'ALARM' :
        
        #provisiono uma instancia do ecs fargate para apoiar o app lambda
        response = ECSclient.update_service(cluster=cluster, service=service, desiredCount=1)
        print(response)
        
        #Espero o Ecs ligar
        container_RUNNING = False
        while container_RUNNING == False:
            
            response = ECSclient.list_tasks(cluster='AlbECSCluster')
            taskarns = response['taskArns']
            if len(taskarns) > 0 :
                describe_tasks_response = ECSclient.describe_tasks(cluster='AlbECSCluster',tasks=taskarns)
                
                for t in describe_tasks_response['tasks']:
                    for c in t['containers']:
                        if c['lastStatus'] == 'RUNNING' : 
                            print('container RUNNING')
                            container_RUNNING = True
                            
            
                if container_RUNNING :
                    # Definindo as novas regras para o listener
                    new_rules = [
                        {
                            'Type' : 'forward',
                            'ForwardConfig': {
                                'TargetGroups': [
                                    {
                                        'TargetGroupArn': targetGroupLambda,
                                        'Weight': 0
                                    },
                                    {
                                        'TargetGroupArn': targetGroupECS,
                                        'Weight': 1
                                    }]}
                        }
                    ]
            
            
                    # Modificando as regras do listener do ALB
                    response = ELBclient.modify_listener(ListenerArn=listenerArn, DefaultActions=new_rules)
                    print(response)
                
        
    if state == 'OK' and previousState != 'OK'  :
        
        # Definindo as novas regras para o listener
        new_rules = [
            {
                'Type' : 'forward',
                'ForwardConfig': {
                    'TargetGroups': [
                        {
                            'TargetGroupArn': targetGroupLambda,
                            'Weight': 1
                        },
                        {
                            'TargetGroupArn': targetGroupECS,
                            'Weight': 0
                        }]}
            }
        ]


        # Modificando as regras do listener do ALB
        response = ELBclient.modify_listener(ListenerArn=listenerArn, DefaultActions=new_rules)
        print(response)

        #Removo a instancia do ecs
        response = ECSclient.update_service(cluster=cluster, service=service, desiredCount=0)
        print(response)
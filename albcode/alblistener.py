import boto3

listenerArn = 'arn:aws:elasticloadbalancing:sa-east-1:281303628498:listener/app/AlbPocGoApp/e8477991703569a5/c95458dc9d6e2c4a'
port = 80
targetGroupLambda = 'arn:aws:elasticloadbalancing:sa-east-1:281303628498:targetgroup/GrupoDestinoLambda/07dbbd92b1ef7d80'
targetGroupECS = 'arn:aws:elasticloadbalancing:sa-east-1:281303628498:targetgroup/tgECS/77418f2ae84a3928'

client = boto3.client('elbv2')  # Criando cliente para o Elastic Load Balancing

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
                    'Weight': 1
                }]}
    }
    # Adicione mais regras conforme necessário
]

# Modificando as regras do listener do ALB
response = client.modify_listener(
    ListenerArn=listenerArn,
    DefaultActions=new_rules
)

print(response)



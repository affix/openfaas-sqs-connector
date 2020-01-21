## openfaas-sqs-connector

The SQS connector connects OpenFaaS functions to AWS SQS Queues.

Goals:

* Allow functions to subscribe to SQS Queues
* Ingest data from sidekiq and execute functions
* Work with the OpenFaaS REST API / Gateway
* Formulate and validate a generic "connector-pattern" to be used for various event sources like sidekiq, AWS SNS, RabbitMQ etc

## Try it out

### Deploy Swarm

Deploy the stack which contains sidekiq and the connector:

```bash
docker stack deploy sidekiq -c ./yaml/connector-swarm.yml
```

* Deploy or update a function so it has an annotation `topic=faas-request` or some other topic

As an example:

```shell
$ faas store deploy figlet --annotation topic="faas-request"
```

The function can advertise more than one topic by using a comma-separated list i.e. `topic=topic1,topic2,topic3`

* Publish some messages to the topic in question i.e. `faas-request`

Instructions are below for publishing messages

* Watch the logs of the sidekiq-connector


### Deploy on Kubernetes

The following instructions show how to run `openfaas-sqs-connector` on Kubernetes.

Deploy a function with a `topic` annotation:

```bash
$ faas store deploy figlet --annotation topic="faas-request" --gateway <faas-netes-gateway-url>
```

Our deployment relies on a secret called aws-secret so lets create a secret

```bash
kubectl create secret generic aws-credentials --from-literal=AWS_ACCESS_KEY_ID=<ACCESS_KEY> --from-literal=AWS_SECRET_ACCESS_KEY_ID=<SECRET_KEY> --from-literal=AWS_REGION=<REGION>

```

Now deploy the connector

```bash
kubectl apply -f ./yaml/kubernetes/connector-dep.yml
```

## Configuration

This configuration can be set in the YAML files for Kubernetes or Swarm.

| env_var               | description                                                 |
| --------------------- |----------------------------------------------------------   |
| `AWS_ACCESS_KEY_ID`      | AWS Access Key ID for IAM User    |
| `AWS_ACCESS_SECRET_KEY_ID`      | AWS Secret Access Key ID for IAM User    |
| `AWS_REGION`            | AWS Region, e.g eu-west-1   |
| `AWS_SQS_QUEUE_NAME`    | SQS Queue for the functions                    |
| `GATEWAY_URL`           | The URL for the API gateway i.e. http://gateway:8080 or http://gateway.openfaas:8080 for Kubernetes       |
| `PRINT_RESPONSE`        | Default is `false` - this will output the response of calling a function in the logs |
| `PRINT_RESPONSE_BODY`        | Default is `false` - this will output the response of calling a function in the logs |

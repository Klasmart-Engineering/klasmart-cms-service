### create table schedules:

```shell script
aws dynamodb create-table \
--table-name schedules \
--attribute-definitions \
AttributeName=id,AttributeType=S \
--key-schema \
AttributeName=id,KeyType=HASH \
--provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1 \
--endpoint-url http://192.168.1.234:18000
```

### create table teachers_schedules:

```shell script
aws dynamodb create-table \
--table-name teachers_schedules \
--attribute-definitions \
AttributeName=teacher_id,AttributeType=S \
AttributeName=schedule_id,AttributeType=S \
--key-schema \
AttributeName=teacher_id,KeyType=HASH \
AttributeName=schedule_id,KeyType=RANGE \
--provisioned-throughput ReadCapacityUnits=10,WriteCapacityUnits=5 \
--endpoint-url http://192.168.1.234:18000
```

### create global-secondary-index OrgIDAndStartAt
```shell script
aws dynamodb update-table \
    --table-name schedules \
    --attribute-definitions AttributeName=org_id,AttributeType=S AttributeName=start_at,AttributeType=N \
    --global-secondary-index-updates \
    "[{\"Create\":{\"IndexName\": \"OrgIDAndStartAt\",\"KeySchema\":[{\"AttributeName\":\"org_id\",\"KeyType\":\"HASH\"},{\"AttributeName\":\"start_at\",\"KeyType\":\"RANGE\"}], \
    \"ProvisionedThroughput\": {\"ReadCapacityUnits\": 10, \"WriteCapacityUnits\": 5      },\"Projection\":{\"ProjectionType\":\"ALL\"}}}]" \
--endpoint-url http://192.168.1.234:18000
```

### create global secondary index repeat_id_and_start_at
```shell script
aws dynamodb update-table \
    --table-name schedules \
    --attribute-definitions AttributeName=repeat_id,AttributeType=S AttributeName=start_at,AttributeType=N \
    --global-secondary-index-updates \
    "[{\"Create\":{\"IndexName\": \"repeat_id_and_start_at\",\"KeySchema\":[{\"AttributeName\":\"repeat_id\",\"KeyType\":\"HASH\"},{\"AttributeName\":\"start_at\",\"KeyType\":\"RANGE\"}], \
    \"ProvisionedThroughput\": {\"ReadCapacityUnits\": 10, \"WriteCapacityUnits\": 5      },\"Projection\":{\"ProjectionType\":\"ALL\"}}}]" \
--endpoint-url http://192.168.1.234:18000
```

### create global-secondary-index teachers_schedules teacher_id_and_start_at
```shell script
aws dynamodb update-table \
    --table-name teachers_schedules \
    --attribute-definitions AttributeName=teacher_id,AttributeType=S AttributeName=start_at,AttributeType=N \
    --global-secondary-index-updates \
    "[{\"Create\":{\"IndexName\": \"teacher_id_and_start_at\",\"KeySchema\":[{\"AttributeName\":\"teacher_id\",\"KeyType\":\"HASH\"},{\"AttributeName\":\"start_at\",\"KeyType\":\"RANGE\"}], \
    \"ProvisionedThroughput\": {\"ReadCapacityUnits\": 10, \"WriteCapacityUnits\": 5      },\"Projection\":{\"ProjectionType\":\"ALL\"}}}]" \
--endpoint-url http://192.168.1.234:18000
```


### delete table template:

```shell script
aws dynamodb delete-table --table-name schedules \
--endpoint-url http://192.168.1.234:18000
```

### delete global-secondary-index template

```shell script
aws dynamodb update-table \                                                                                                                                                                                                   255 ↵
    --table-name schedules \
--global-secondary-index-updates \
"[{\"Delete\":{\"IndexName\": \"OrgID-index\"}}]" \
--endpoint-url http://192.168.1.234:18000
```

### show global-secondary-index template

```shell script
 aws dynamodb describe-table --table-name schedules --endpoint-url http://192.168.1.234:18000
```
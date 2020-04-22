# lambda-s3-logrotate


## S3 event sample

```json

[
{
    "eventVersion": "2.1",
    "eventSource": "aws:s3",
    "awsRegion": "ap-southeast-1",
    "eventTime": "2020-04-21T02:05:27.824Z",
    "eventName": "ObjectCreated:Put",
    "userIdentity": {
        "principalId": "AWS:xxxx"
    },
    "requestParameters": {
        "sourceIPAddress": "1.1.1.1"
    },
    "responseElements": {
        "x-amz-id-2": "xxxxxx",
        "x-amz-request-id": "xxxxx"
    },
    "s3": {
        "s3SchemaVersion": "1.0",
        "configurationId": "xxxx",
        "bucket": {
            "name": "xxxx",
            "ownerIdentity": {
                "principalId": "xxxxx"
            },
            "arn": "arn:aws:s3:::xxxx"
        },
        "object": {
            "key": "%E4%BD%A0%E5%AE%B6%E5%A8%83%F0%9F%90%B1.png",
            "size": 1603233,
            "urlDecodedKey": "",
            "versionId": "",
            "eTag": "xxxx",
            "sequencer": "xxxxx"
        }
    }
}
]

```

## lambda role 

```json

{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "logs:CreateLogGroup",
            "Resource": "arn:aws:logs:ap-southeast-1:<account name>:*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": [
                "arn:aws:logs:ap-southeast-1:<account name>:log-group:/aws/lambda/s3-logrotate:*"
            ]
        },
        {
            "Action": [
                "s3:ListBucket"
            ],
            "Effect": "Allow",
            "Resource": [
                "arn:aws:s3:::<bucket name>"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject"
            ],
            "Resource": [
                "arn:aws:s3:::<bucket name>/access-logs/*"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject"
            ],
            "Resource": [
                "arn:aws:s3:::<bucket name>/access-logs-logrotate/*"
            ]
        }
    ]
}

```


## migration ruby script

``` bury
# encoding: UTF-8
require 'date'

bucket_name="<bucket name>"

start_date = "2018-12-21"
date_range = (Date.parse(start_date)..DateTime.now)
region="cn-northwest-1"


date_range.sort.uniq.each do |d|
  d1 = d.strftime("%Y-%m-%d")
  d2 = d.strftime("%Y/%m/%d")
  puts cmd=%Q~aws s3 sync s3://#{bucket_name}/access-logs/ s3://#{bucket_name}/access-logs-logrotate/#{d2}/  --exclude "*" --include="#{d1}-*" --region=#{region}~
end





```

## create s3 object lifecycle rule

```
Scope: access-logs/
Expire after 7 days
```
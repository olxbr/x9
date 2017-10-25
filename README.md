# x9
X-9 is a brazilian slang for "informer".

A X-9 person is motivated by envy.

This X-9 was motivated by keep learning golang.

It is a tool that alerts in Slack missbehavior of AWS instances.

A missbehavior is when an instance runs for short time (check the option TOLERANCE).

It is most usefull to monitor auto scaling groups (ASG). When an instance is member of a ASG the ASG name is used as its key.

It also provides an easy API that anwsers easy questions we are tired to give everyday, like:


* What are the instances running right now, by tags, regions ?
* How many instances we created in the last 24 hours ?
* How many spot instances we created today ?
* How many instances run for less than an hour ?
* How many spot instances run for less than an hour today ?
* How many instances we run by region ?
* How many instaces are run with a certain tag (for example by environment)?
* How many instances by type ?
* What are the missbehaving auto scaling groups right now ?

(check /api for a complete list)

# How does it work ?

Simply by getting the result of "describe instances" and sumarizes it in redis.
The data expires in 24 hours.

# Tags

A rudimentary tag support is provide in this first release. You must use the tags bellow.
If don't use tags or different tags, don't worry, they will be show as "none".

 |Tag|Description|
 |---|---|
 |Env|Environment, ie: QA, DEV, STG, PROD|
 |Product|System|
 |App|Component of System|

# Running it locally

```
$ # make sure your aws cli is working
$ brew install go
$ brew install redis
$ brew install glide
$ mkdir ~/go/src/github.com/vivareal
$ cd ~/go/src/github.com/vivareal
$ git clone git@github.com:VivaReal/x9.git
$ export GOPATH=~/go/
$ glide install
$ redis-server &
$ export SLACK_WEBHOOK_URL="https://myslack...."
$ go run main.go
$ curl localhost:6969
```


# Using AWS key pairs

```
$ export AWS_ACCESS_KEY_ID=XXXXXXXXX
$ export AWS_SECRET_ACCESS_KEY=XXXXXXX
$ go run main.go
```

# How to build it
```
$ make build
(will create the x9 executable file)
```

# How to build it and run in Docker
```
$ make docker_image DOCKER_REPO="myrepo"
$ docker run -d -n redis redis
$ docker run -p 6969:6969 \
-e REDIS_SERVER=redis:6379 \
-e AWS_ACCESS_KEY_ID=XXX \
-e AWS_SECRET_ACCESS_KEY="XXX" \
-e SLACK_WEBHOOK_URL="https://myslack...." \
myrepo/x9
$ curl localhost:6969/api
```

# All options and their defaults

|*Environment variable*|*Default value*|*Description*|
|---|---|---|
|AWS_ACCESS_KEY_ID|-|optional, default provided by the aws-cli configuration|
|AWS_SECRET_ACCESS_KEY|-|optional, default provided by the aws-cli configuration|
|SLACK_WEBHOOK_URL|error|Slack bot URL|
|REDIS_SERVER|localhost:6379|Redis server address and port|
|TOLERANCE|3000|Minimum amount of time instances must run to not be considered missbehave| 
|ALERT_TIMEFRAME|1200|Interval between checks|
|SERVICE_PORT|6969|Webserver port|
|REGIONS|"sa-east-1,us-east-1"|Regions were to run "describe-instances"|


# API reference

```
$ lynx --dump localhost:6969/api
```

# Redis keys reference

|*Starting by*|*Meaning*|
|---|---|
|r_*|Sum of all regions|
|w_*|Sum of "wasted" instances (run for less time than $TOLERANCE)|
|s_*|Sum of spot instances|
|tmp_*|temporary keys|

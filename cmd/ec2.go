package cmd

import (
	"fmt"
	"github.com/araddon/dateparse"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/go-redis/redis"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Instance is the object that holds all configurations about a AWS instance.
type Instance struct {
	Region          string
	Env             string
	App             string
	Product         string
	isSpot          string
	Type            string
	Expires         int64
	Status          string
	isWasted        string
	last24Hours     bool
	lastFrameWasted bool
	isASG           bool
	Asg             string
}

func updateRedis(current map[string]*Instance) {
	fmt.Printf("%v - [Update redis requested]\n", time.Now())
	rc := redis.NewClient(&redis.Options{
		Addr:     getOptEnv("REDIS_SERVER", "localhost:6379"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	for instanceId, instance := range current {

		if instance.last24Hours == true {
			member := redis.Z{Score: float64(instance.Expires), Member: instanceId}

			// by waste
			if instance.isWasted == "1" {
				rc.ZAddNX("w_region-"+instance.Region, member).Result()
				rc.ZAddNX("w_env-"+instance.Env, member).Result()
				rc.ZAddNX("w_app-"+instance.App, member).Result()
				rc.ZAddNX("w_product-"+instance.Product, member).Result()
				rc.ZAddNX("w_spot-"+instance.isSpot, member).Result()
				rc.ZAddNX("w_type-"+instance.Type, member).Result()
				rc.ZAddNX("w_wasted-"+instance.isWasted, member).Result()
			}

			// by spot
			if instance.isSpot == "1" {
				rc.ZAddNX("s_region-"+instance.Region, member).Result()
				rc.ZAddNX("s_env-"+instance.Env, member).Result()
				rc.ZAddNX("s_app-"+instance.App, member).Result()
				rc.ZAddNX("s_product-"+instance.Product, member).Result()
				rc.ZAddNX("s_spot-"+instance.isSpot, member).Result()
				rc.ZAddNX("s_type-"+instance.Type, member).Result()
				rc.ZAddNX("s_wasted-"+instance.isWasted, member).Result()
			} else {
				// by regular
				rc.ZAddNX("r_region-"+instance.Region, member).Result()
				rc.ZAddNX("r_env-"+instance.Env, member).Result()
				rc.ZAddNX("r_app-"+instance.App, member).Result()
				rc.ZAddNX("r_product-"+instance.Product, member).Result()
				rc.ZAddNX("r_spot-"+instance.isSpot, member).Result()
				rc.ZAddNX("r_type-"+instance.Type, member).Result()
				rc.ZAddNX("r_wasted-"+instance.isWasted, member).Result()
			}
		}

		rc.ZIncrBy("tmp_current", 1, "Total").Result()
		rc.ZIncrBy("tmp_current", 1, "Region-"+instance.Region).Result()
		rc.ZIncrBy("tmp_current", 1, "Env-"+instance.Env).Result()
		rc.ZIncrBy("tmp_current", 1, "App-"+instance.App).Result()
		rc.ZIncrBy("tmp_current", 1, "Product-"+instance.Product).Result()
		rc.ZIncrBy("tmp_current", 1, "isSpot-"+instance.isSpot).Result()
		rc.ZIncrBy("tmp_current", 1, "Type-"+instance.Type).Result()
		rc.ZIncrBy("tmp_current", 1, "Status-"+instance.Status+"-"+instance.Region).Result()
		rc.ZIncrBy("tmp_current", 1, "Status-"+instance.Status).Result()
		rc.ZIncrBy("tmp_current", 1, "isWasted-"+instance.isWasted).Result()

		if instance.lastFrameWasted {
			asg := ""
			if instance.isASG {
				asg = "ASG-"
				rc.ZIncrBy("tmp_alertasasg", 1, "Env:"+instance.Env+"----ASG:"+instance.Asg+"----Type:"+instance.Type).Result()
			}
			rc.ZIncrBy("tmp_alertas", 1, instance.Env+"-"+instance.Product+"-"+instance.App+"-"+asg+instance.Type).Result()
		}
	}
	rc.Rename("tmp_current", "current").Result()
	rc.Del("alertas").Result()
	rc.Rename("tmp_alertas", "alertas").Result()
	rc.Del("alertasasg")
	rc.Rename("tmp_alertasasg", "alertasasg").Result()
	cleanRedisKeys()

	fmt.Printf("%v - [Update redis finished]\n", time.Now())
}

func cleanRedisKeys() {
	fmt.Printf("%v - [Starting cleaning keys]\n", time.Now())
	rc := redis.NewClient(&redis.Options{
		Addr:     getOptEnv("REDIS_SERVER", "localhost:6379"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	keys, err := rc.Keys("[a-z]_*").Result()

	lessthan24hours := strconv.FormatInt(time.Now().Unix()-86400, 10)

	if err == redis.Nil {
		fmt.Println("Redis error")
	}

	for i := 0; i < len(keys); i++ {
		rc.ZRemRangeByScore(keys[i], "0", lessthan24hours).Result()
		// 	ZCount(key, min, max string) *IntCmd

		count, _ := rc.ZCount(keys[i], "-inf", "+inf").Result()
		if count == 0 {
			rc.Del(keys[i]).Result()
		}

	}
	fmt.Printf("%v - [Finished cleaning keys]\n", time.Now())
}

func getInstances() {
	fmt.Printf("%v - [Starting get instances]\n", time.Now())
	t := time.Now().Unix()

	tolerance, _ := strconv.ParseInt(getOptEnv("TOLERANCE", "3000"), 10, 64)
	alertframe, _ := strconv.ParseInt(getOptEnv("ALERT_TIMEFRAME", "1200"), 10, 64)

	sess := session.Must(session.NewSession())

	Regions := strings.Split(getOptEnv("REGIONS", "sa-east-1,us-east-1"), ",")
	params := &ec2.DescribeInstancesInput{
	//
	//	Filters: []*ec2.Filter{
	//			{
	//				Name: aws.String("Region"),
	//				Values: aws.String(awsRegion)
	//			},
	//			{	Name:   aws.String("instance-lifecycle"), // "spot" instance lifecycle
	//				Values: []*string{aws.String("spot")},
	//			},
	//{
	//	Name:   aws.String("instance-state-name"),
	//	Values: []*string{aws.String("running")},
	//},
	//	},
	}

	re := regexp.MustCompile("\\((.*)\\)")

	current := make(map[string]*Instance)
	for _, awsRegion := range Regions {

		fmt.Printf("%v - [Starting get instances in %v]\n", time.Now(), awsRegion)

		svc := ec2.New(sess, &aws.Config{Region: aws.String(awsRegion)})

		resp, err := svc.DescribeInstances(params)

		if err != nil {
			fmt.Println("there was an error listing instances in", awsRegion, err.Error())
			log.Fatal(err.Error())
		}

		for _, reserv := range resp.Reservations {

			for _, inst := range reserv.Instances {

				status := *inst.State.Name
				InstanceId := *inst.InstanceId
				InstanceType := *inst.InstanceType

				isSpot := "0"
				if inst.InstanceLifecycle != nil && *inst.InstanceLifecycle == "spot" {
					isSpot = "1"
				}

				Env := "none"
				Product := "none"
				App := "none"
				isASG := false
				Asg := ""
				for _, tag := range inst.Tags {

					switch *tag.Key {

					case "Env":
						Env = *tag.Value
					case "App":
						App = *tag.Value
					case "Product":
						Product = *tag.Value
					case "aws:autoscaling:groupName":
						isASG = true
						Asg = *tag.Value
					}
				}

				isWasted := "0"
				last24Hours := false
				lastFrameWasted := false

				if status == "terminated" && len(*inst.StateTransitionReason) > 0 {
					datestates := re.FindAllStringSubmatch(*inst.StateTransitionReason, 1)

					if len(datestates) > 0 && len(datestates[0]) > 1 {

						datestate := datestates[0][1]
						// Example of datestate
						// `User initiated (2017-07-26 18:55:53 GMT)`
						terminated, err := dateparse.ParseAny(datestate)
						if err != nil {
							panic(err.Error())
						}

						if terminated.Unix()-inst.LaunchTime.Unix() < tolerance {
							isWasted = "1"

							if t-terminated.Unix() < alertframe {
								lastFrameWasted = true
							}
						}
					}
					last24Hours = true

				}

				if status == "running" && t-inst.LaunchTime.Unix() < 86400 {
					last24Hours = true
				}

				current[InstanceId] = &Instance{
									Region:          awsRegion,
									Env:             Env,
									App:             App,
									Product:         Product,
									isSpot:          isSpot,
									Type:            InstanceType,
									Expires:         inst.LaunchTime.Unix(),
									Status:          status,
									isWasted:        isWasted,
									last24Hours:     last24Hours,
									lastFrameWasted: lastFrameWasted,
									isASG:           isASG,
									Asg:             Asg,
								}

			}
		}
		fmt.Printf("%v - [Finished getting instances in %v]\n", time.Now(), awsRegion)
	}
	fmt.Printf("%v - [Finished getting instances]\n", time.Now())
	updateRedis(current)
	go alertSlack()
	time.Sleep(time.Duration(alertframe) * time.Second)
	getInstances()
}

package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"regexp"
	"strings"
)

var (
	GitCommitId string
	BuildTime   string
)

var (
	logrotatePrefix    string
	accessLogPrefix    string
	awsRegion          string
	accessLogKeyRegexp *regexp.Regexp
)

func init() {
	getEnvFunc := func(key, defaultVal string) string {
		v := os.Getenv(key)
		if strings.TrimSpace(v) == "" {
			v = defaultVal
		}
		if strings.HasSuffix(v, "/") {
			return v[0 : len(v)-1]
		}
		return v
	}
	logrotatePrefix = getEnvFunc("LOGROTATE_PREFIX", "access-logs-logrotate")
	accessLogPrefix = getEnvFunc("ACCESS_LOGS_PREFIX", "access-logs")
	awsRegion = os.Getenv("AWS_REGION")

	accessLogKeyRegexp = regexp.MustCompile(fmt.Sprintf(`^%v/(?P<Date>\d{4}-\d{2}-\d{2})-(?P<Time>\d{2}-\d{2}-\d{2})-.+$`, accessLogPrefix))
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, s3Event events.S3Event) {
	cfg := aws.NewConfig()
	if awsRegion != "" {
		cfg = cfg.WithRegion(awsRegion)
	}
	sess, err := session.NewSession(cfg)
	if err != nil {
		logrus.Errorf("make aws session error=%v", err.Error())
		return
	}

	s3v := s3.New(sess)
	for _, record := range s3Event.Records {
		// skip not ObjectCreated:Put event
		if record.EventName != "ObjectCreated:Put" {
			logrus.Warnf("skip event=%v", record.EventName)
			continue
		}
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key
		if !accessLogKeyRegexp.MatchString(key) {
			logrus.Warnf("object not match access logs filename")
			continue
		}
		match := accessLogKeyRegexp.FindStringSubmatch(key)
		result := make(map[string]string)
		for i, name := range accessLogKeyRegexp.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}
		keyWithoutOriginPrefix := key[len(accessLogPrefix):]
		date, ok := result["Date"]
		if !ok {
			logrus.Warnf("get date from object key failed")
			continue
		}
		newObjectKey := path.Join(logrotatePrefix, strings.ReplaceAll(date, "-", "/"), keyWithoutOriginPrefix)
		source := fmt.Sprintf("%s/%s", bucket, key)
		if _, err := s3v.CopyObject(&s3.CopyObjectInput{
			CopySource: aws.String(source),
			Bucket:     aws.String(bucket),
			Key:        aws.String(newObjectKey),
		}); err != nil {
			logrus.Errorf("copy object from %s to %s error,error=%s", source, bucket+"/"+newObjectKey, err.Error())
			continue
		} else {
			logrus.Infof("copy object from %s to %s success", source, bucket+"/"+newObjectKey)
		}
	}
}

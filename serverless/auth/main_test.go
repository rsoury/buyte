package main_test

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/gobwas/glob"
)

var (
	Event events.APIGatewayCustomAuthorizerRequestTypeRequest
	Data  = `
	{
		"type": "REQUEST",
		"methodArn": "arn:aws:execute-api:us-east-1:123456789012:s4x3opwd6i/test/GET/request",
		"resource": "/request",
		"path": "/request",
		"httpMethod": "GET",
		"headers": {
			"X-AMZ-Date": "20170718T062915Z",
			"Accept": "*/*",
			"HeaderAuth1": "headerValue1",
			"CloudFront-Viewer-Country": "US",
			"CloudFront-Forwarded-Proto": "https",
			"CloudFront-Is-Tablet-Viewer": "false",
			"CloudFront-Is-Mobile-Viewer": "false",
			"User-Agent": "...",
			"X-Forwarded-Proto": "https",
			"CloudFront-Is-SmartTV-Viewer": "false",
			"Host": "....execute-api.us-east-1.amazonaws.com",
			"Accept-Encoding": "gzip, deflate",
			"X-Forwarded-Port": "443",
			"X-Amzn-Trace-Id": "...",
			"Via": "...cloudfront.net (CloudFront)",
			"X-Amz-Cf-Id": "...",
			"X-Forwarded-For": "..., ...",
			"Postman-Token": "...",
			"cache-control": "no-cache",
			"CloudFront-Is-Desktop-Viewer": "true",
			"Content-Type": "application/x-www-form-urlencoded"
		},
		"queryStringParameters": {
			"QueryString1": "queryValue1"
		},
		"pathParameters": {},
		"stageVariables": {
			"StageVar1": "stageValue1"
		},
		"requestContext": {
			"path": "/request",
			"accountId": "123456789012",
			"resourceId": "05c7jb",
			"stage": "test",
			"requestId": "...",
			"identity": {
				"apiKey": "...",
				"sourceIp": "..."
			},
			"resourcePath": "/request",
			"httpMethod": "GET",
			"apiId": "s4x3opwd6i"
		}
	}
	`
)

func TestHandler(t *testing.T) {
	// ...
	path := "/v1/public/checkout/a1aeb05b-3f02-4dd5-b9a9-941377fb9c15/"
	g := glob.MustCompile("/v*/public/**")
	if g.Match(path) {
		t.Log("Glob works...")
	} else {
		t.Error("Glob fail")
	}
}

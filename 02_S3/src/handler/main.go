package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

type JsonStruct struct {
	A int
	B string
}

func init() {
	log.Printf("Gin cold start\n")
	// Ginの初期化
	router := gin.Default()
	// GETメソッド
	router.GET("/test", func(ctx *gin.Context) {
		// セッションの作成
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		// S3クライアントインスタンスの作成
		s3client := s3.New(sess)

		// バケット名
		bucketName := "go-gin-sls-test-bucket"
		textFileKey := "test.txt"
		jsonFileKey := "test.json"

		// バケットの作成
		log.Printf("Create bucket: %s\n", bucketName)
		resCreateBucket, errCreateBucket := s3client.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		if errCreateBucket != nil {
			log.Printf("failed to create bucket\n %v\n", errCreateBucket)
			// ctx.String(http.StatusInternalServerError, "%v", errCreateBucket)
			// return
		}
		log.Printf("result: %v", resCreateBucket)
		// バケットが生成されるまで待機
		errWaitUntilBucketExists := s3client.WaitUntilBucketExistsWithContext(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		})
		if errWaitUntilBucketExists != nil {
			log.Printf("failed to create bucket\n %v\n", errWaitUntilBucketExists)
			ctx.String(http.StatusInternalServerError, "%v", errWaitUntilBucketExists)
			return
		}

		// バケットの一覧取得
		log.Printf("List buckets: \n")
		resListBuckets, errListBuckets := s3client.ListBucketsWithContext(ctx, nil)
		if errListBuckets != nil {
			log.Printf("failed to list buckets\n %v\n", errListBuckets)
			ctx.String(http.StatusInternalServerError, "%v", errListBuckets)
			return
		}
		// 取得したバケット一覧の表示
		for _, b := range resListBuckets.Buckets {
			log.Printf("* %s created on %s\n", aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
		}

		// TXTファイルの作成
		log.Printf("Put object: %s\n", textFileKey)
		resPutObject, errPutObject := s3client.PutObjectWithContext(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(textFileKey),
			Body:   bytes.NewReader([]byte("Hello world.")),
		})
		if errPutObject != nil {
			log.Printf("failed to put object\n %v\n", errPutObject)
			ctx.String(http.StatusInternalServerError, "%v", errPutObject)
			return
		}
		log.Printf("result: %v", resPutObject)
		// TXTファイルが生成されるまで待機
		errWaitUntilObjectExists := s3client.WaitUntilObjectExistsWithContext(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(textFileKey),
		})
		if errWaitUntilObjectExists != nil {
			log.Printf("failed to put object\n %v\n", errWaitUntilObjectExists)
			ctx.String(http.StatusInternalServerError, "%v", errWaitUntilObjectExists)
			return
		}
		// JSONファイルの作成
		log.Printf("Put object: %s\n", jsonFileKey)
		jsonStructData := JsonStruct{A: 1, B: "bbb"}
		jsonData, errJSONMarshal := json.Marshal(jsonStructData)
		if errJSONMarshal != nil {
			log.Printf("failed to put object\n %v\n", errJSONMarshal)
			ctx.String(http.StatusInternalServerError, "%v", errJSONMarshal)
			return
		}
		resPutObject, errPutObject = s3client.PutObjectWithContext(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(jsonFileKey),
			Body:   bytes.NewReader(jsonData),
		})
		if errPutObject != nil {
			log.Printf("failed to put object\n %v\n", errPutObject)
			ctx.String(http.StatusInternalServerError, "%v", errPutObject)
			return
		}
		log.Printf("result: %v", resPutObject)
		// JSONファイルが生成されるまで待機
		errWaitUntilObjectExists = s3client.WaitUntilObjectExistsWithContext(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(jsonFileKey),
		})
		if errWaitUntilObjectExists != nil {
			log.Printf("failed to put object\n %v\n", errWaitUntilObjectExists)
			ctx.String(http.StatusInternalServerError, "%v", errWaitUntilObjectExists)
			return
		}

		// TXTファイル内容の取得
		log.Printf("Get object: %s\n", textFileKey)
		resGetTXT, errGetTXT := s3client.GetObjectWithContext(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(textFileKey),
		})
		if errGetTXT != nil {
			log.Printf("failed to get object\n %v\n", errGetTXT)
			ctx.String(http.StatusInternalServerError, "%v", errGetTXT)
			return
		}
		resGetTXTBody, errGetTXTBody := ioutil.ReadAll(resGetTXT.Body)
		defer resGetTXT.Body.Close()
		if errGetTXTBody != nil {
			log.Printf("failed to get object\n %v\n", errGetTXT)
			ctx.String(http.StatusInternalServerError, "%v", errGetTXT)
			return
		}
		// byte[]からstringへの変換
		log.Printf("text: " + string(resGetTXTBody))
		// JSONファイル内容の取得
		log.Printf("Get object: %s\n", jsonFileKey)
		resGetJSON, errGetJSON := s3client.GetObjectWithContext(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(jsonFileKey),
		})
		if errGetJSON != nil {
			log.Printf("failed to get object\n %v\n", errGetJSON)
			ctx.String(http.StatusInternalServerError, "%v", errGetJSON)
			return
		}
		resGetJSONBody, errGetJSONBody := ioutil.ReadAll(resGetJSON.Body)
		defer resGetJSON.Body.Close()
		if errGetJSONBody != nil {
			log.Printf("failed to get object\n %v\n", errGetJSON)
			ctx.String(http.StatusInternalServerError, "%v", errGetJSON)
			return
		}
		// JSON(byte[])から構造体への変換
		errJSONUnmarshal := json.Unmarshal(resGetJSONBody, &jsonStructData)
		if errJSONUnmarshal != nil {
			log.Printf("failed to get object\n %v\n", errJSONUnmarshal)
			ctx.String(http.StatusInternalServerError, "%v", errJSONUnmarshal)
			return
		}
		log.Printf("json: %+v\n", jsonStructData)

		// バケット内のオブジェクト一覧の取得
		log.Printf("List objects: \n")
		resListObjects, errListObjects := s3client.ListObjectsWithContext(ctx, &s3.ListObjectsInput{Bucket: aws.String(bucketName)})
		if errListObjects != nil {
			log.Printf("failed to list objects\n %v\n", errListObjects)
			ctx.String(http.StatusInternalServerError, "%v", errListObjects)
			return
		}
		// 取得したバケット内のオブジェクト一覧の表示
		for idx, item := range resListObjects.Contents {
			log.Println("[", idx, "]")
			log.Println("Name:         ", *item.Key)
			log.Println("Last modified:", *item.LastModified)
			log.Println("Size:         ", *item.Size)
			log.Println("Storage class:", *item.StorageClass)
		}

		// オブジェクトの削除
		for _, item := range resListObjects.Contents {
			resDeleteObject, errDeleteObject := s3client.DeleteObjectWithContext(ctx,
				&s3.DeleteObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String(*item.Key)})
			if errDeleteObject != nil {
				log.Printf("failed to delete object\n %v\n", errDeleteObject)
				ctx.String(http.StatusInternalServerError, "%v", errDeleteObject)
				return
			}
			log.Println(resDeleteObject)
			// オブジェクトが削除されるまで待機
			errWaitUntilObjectNotExists := s3client.WaitUntilObjectNotExistsWithContext(ctx, &s3.HeadObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(*item.Key),
			})
			if errWaitUntilObjectNotExists != nil {
				log.Printf("failed to delete object\n %v\n", errWaitUntilObjectNotExists)
				ctx.String(http.StatusInternalServerError, "%v", errWaitUntilObjectNotExists)
				return
			}
		}

		// バケットの削除
		log.Printf("Delete bucket: %s\n", bucketName)
		resDeleteBucket, errDeleteBucket := s3client.DeleteBucketWithContext(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
		if errDeleteBucket != nil {
			log.Printf("failed to delete bucket\n %v\n", errDeleteBucket)
			ctx.String(http.StatusInternalServerError, "%v", errDeleteBucket)
			return
		}
		log.Printf("result: %v", resDeleteBucket)
		// バケットが削除されるまで待機
		errWaitUntilBucketNotExists := s3client.WaitUntilBucketNotExistsWithContext(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		})
		if errWaitUntilBucketNotExists != nil {
			log.Printf("failed to delete bucket\n %v\n", errWaitUntilBucketNotExists)
			ctx.String(http.StatusInternalServerError, "%v", errWaitUntilBucketNotExists)
			return
		}

		// 200 OK レスポンス
		ctx.String(http.StatusOK, "OK")
	})
	// 各種APIを登録
	ginLambda = ginadapter.New(router)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}

package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

func init() {
	// Lambdaでは標準出力と標準エラー出力はAWS CloudWatch Logsに送信される
	log.Printf("Gin cold start")
	// Ginの初期化
	router := gin.Default()
	// POSTメソッド
	type RequestBodyDataType struct {
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
	}
	router.POST("/user/:name", func(ctx *gin.Context) {
		var json RequestBodyDataType
		// リクエストボディが規定の型を満たさない場合のエラー処理
		if err := ctx.ShouldBindJSON(&json); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// リクエストボディの取得
		ctx.JSON(http.StatusOK,
			gin.H{"firstname": json.Firstname, "lastname": json.Lastname})
	})
	// GETメソッド
	router.GET("/user/:name", func(ctx *gin.Context) {
		// クエリの取得
		firstname := ctx.DefaultQuery("firstname", "Guest")
		lastname := ctx.Query("lastname")
		ctx.JSON(http.StatusOK, gin.H{
			"firstname": firstname,
			"lastname":  lastname,
		})
	})
	// PUTメソッド
	router.PUT("/user/:name", func(ctx *gin.Context) {
		// URLパラメータの取得
		name := ctx.Param("name")
		ctx.String(http.StatusOK, "Hello %s", name)
	})
	// DELETEメソッド
	router.DELETE("/user/:name", func(ctx *gin.Context) {
		ctx.String(200, "OK")
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

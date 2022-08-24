package auth

import (
	"bytes"
	"net/http"
	"time"

	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/log"
	"github.com/gin-gonic/gin"
)

func InitRouter(app OAuthApp) http.Handler {
	router := gin.New()
	router.Use(CorsMiddleWare())

	router.GET("/version", func(c *gin.Context) {
		type version struct {
			Version string
		}
		c.JSON(http.StatusOK, version{Version: core.Version})
	})

	router.POST("/verify", verifyInterceptor(), app.Verify)
	router.POST("/genToken", app.GenerateToken)
	router.GET("/token", app.GetToken)
	router.GET("/tokens", app.Tokens)
	router.DELETE("/token", app.RemoveToken)
	router.POST("/recoverToken", app.RecoverToken)

	userGroup := router.Group("/user")
	userGroup.PUT("/new", app.CreateUser)
	userGroup.POST("/update", app.UpdateUser)
	userGroup.GET("/list", app.ListUsers)
	userGroup.GET("", app.GetUser)
	userGroup.GET("/has", app.HasUser)
	userGroup.POST("/del", app.DeleteUser)
	userGroup.POST("/recover", app.RecoverUser)

	rateLimitGroup := userGroup.Group("/ratelimit")
	rateLimitGroup.POST("/upsert", app.UpsertUserRateLimit)
	rateLimitGroup.POST("/del", app.DelUserRateLimit)
	rateLimitGroup.GET("", app.GetUserRateLimit)

	minerGroup := userGroup.Group("/miner")
	minerGroup.GET("", app.GetUserByMiner)
	minerGroup.POST("/add", app.UpsertMiner)
	minerGroup.GET("/list", app.ListMiners)
	minerGroup.GET("/has", app.HasMiner)
	minerGroup.POST("/del", app.DeleteMiner)

	signerGroup := userGroup.Group("/signer")
	signerGroup.GET("", app.GetUserBySigner)
	signerGroup.POST("/add", app.UpsertSigner)
	signerGroup.GET("/list", app.ListSigner)
	signerGroup.GET("/has", app.HasSigner)
	signerGroup.POST("/del", app.DeleteSigner)

	return router
}

func verifyInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		defer func(begin time.Time) {
			verifyLog(begin, c, blw.body)
		}(time.Now())
		c.Writer = blw
		c.Next()
	}
}

func verifyLog(begin time.Time, c *gin.Context, writer *bytes.Buffer) {
	fields := log.Fields{
		core.FieldIP:      c.ClientIP(),
		core.FieldSpanId:  c.Request.Header["spanId"],  //nolint
		core.FieldPreHost: c.Request.Header["preHost"], //nolint
		core.FieldElapsed: time.Since(begin).Milliseconds(),
		core.FieldToken:   c.Request.Form.Get("token"),
		core.FieldSvcName: c.Request.Header["svcName"], //nolint
	}
	fields[core.MTMethod] = "verify"
	errs := c.Errors
	if len(errs) > 0 {
		log.WithFields(fields).Errorln(errs.String())
		return
	}
	fields[core.FieldName] = c.Keys[core.FieldName]
	log.WithFields(fields).Traceln(writer.String())
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func CorsMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers",
			"DNT,X-Mx-ReqToken,Keep-Alive,User-Agent,X-Requested-With,"+
				"If-Modified-Since,Cache-Control,Content-Type,Authorization,X-Forwarded-For,Origin,"+
				"X-Real-Ip,spanId,preHost,svcName")
		c.Header("Content-Type", "application/json")
		if c.Request.Method == "OPTIONS" {
			c.JSON(http.StatusOK, "ok!")
		}
		c.Next()
	}
}

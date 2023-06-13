package auth

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/etherlabsio/healthcheck/v2"

	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/sophon-auth/core"
	"github.com/ipfs-force-community/sophon-auth/log"
)

// todo: rm checkPermission after v1.13.0
func InitRouter(app OAuthApp) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.ContextWithFallback = true
	router.Use(CorsMiddleWare())
	router.Use(RewriteAddressInUrl())
	router.Use(permMiddleWare(app))

	headlerFunc := healthcheck.HandlerFunc()
	router.GET("/healthcheck", func(c *gin.Context) {
		headlerFunc(c.Writer, c.Request)
	})

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
	userGroup.POST("/verify", app.VerifyUsers)
	userGroup.GET("/has", app.HasUser)
	userGroup.POST("/del", app.DeleteUser)
	userGroup.POST("/recover", app.RecoverUser)

	rateLimitGroup := userGroup.Group("/ratelimit")
	rateLimitGroup.POST("/upsert", app.UpsertUserRateLimit)
	rateLimitGroup.POST("/del", app.DelUserRateLimit)
	rateLimitGroup.GET("", app.GetUserRateLimit)

	// Compatible with older versions(<=v1.6.0)
	minerGroup := router.Group("/miner")
	minerGroup.GET("", app.GetUserByMiner)
	minerGroup.GET("/has-miner", app.HasMiner)
	minerGroup.GET("/list-by-user", app.ListMiners)
	minerGroup.POST("/add-miner", app.UpsertMiner)
	minerGroup.POST("/del", app.DeleteMiner)

	minerGroup.GET("/has", app.HasMiner)

	userMinerGroup := userGroup.Group("/miner")
	userMinerGroup.GET("", app.GetUserByMiner)
	userMinerGroup.POST("/add", app.UpsertMiner)
	userMinerGroup.GET("/exist", app.MinerExistInUser)
	userMinerGroup.GET("/list", app.ListMiners)
	userMinerGroup.POST("/del", app.DeleteMiner)

	userSignerGroup := userGroup.Group("/signer")
	userSignerGroup.GET("", app.GetUserBySigner)
	userSignerGroup.POST("/register", app.RegisterSigners)
	userSignerGroup.GET("/exist", app.SignerExistInUser)
	userSignerGroup.GET("/list", app.ListSigner)
	userSignerGroup.POST("/unregister", app.UnregisterSigners)

	signerGroup := router.Group("/signer")
	signerGroup.GET("/has", app.HasSigner)
	signerGroup.POST("/del", app.DelSigner)

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

func RewriteAddressInUrl() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(c.Request.URL.Query()) > 0 {
			queryParams := c.Request.URL.Query()
			for key, params := range queryParams {
				if key == "miner" || key == "signer" {
					for index, v := range params {
						params[index] = "\"" + v + "\""
					}
				}
			}
			c.Request.RequestURI = c.FullPath() + "?" + queryParams.Encode()
			var err error
			c.Request.URL, err = url.ParseRequestURI(c.Request.RequestURI)
			if err != nil {
				log.Errorf("parse request url `%s` failed: %v", c.Request.RequestURI, err)
				_ = c.AbortWithError(http.StatusInternalServerError, errors.New("fail when rewrite request url"))
			}
		}

		c.Next()
	}
}

func permMiddleWare(app OAuthApp) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get(core.AuthorizationHeader)

		if token == "" {
			token = c.Request.FormValue("token")
			if token != "" {
				token = "Bearer " + token
			}
		}

		if !strings.HasPrefix(token, "Bearer ") {
			log.Warnf("missing Bearer prefix in sophon-auth header")
			c.Writer.WriteHeader(401)
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")

		jwtPayload, err := app.verify(token)
		if err != nil {
			log.Warnf("verify token %s failed: %s", token, err)
			c.Writer.WriteHeader(401)
			return
		}

		reqCtx := core.CtxWithPerm(c.Request.Context(), jwtPayload.Perm)
		if len(jwtPayload.Name) != 0 {
			reqCtx = core.CtxWithName(reqCtx, jwtPayload.Name)
		}
		c.Request = c.Request.WithContext(reqCtx)

		c.Next()
	}
}

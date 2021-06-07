package server

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/hagemt/bijection/gpt/cmd/iGod/client"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

type iGodService struct {
	client *http.Client
	speaks client.Speaker
	engine *gin.Engine
	Service
}

func ListenAndServe(ctx context.Context, addr string) error {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	iGod := client.NewDeity(extractNames(ctx))
	service := &iGodService{
		client: httpClient,
		speaks: iGod,
		engine: createEngine(),
	}
	return service.ListenAndServe(ctx, addr)
}

func extractNames(ctx context.Context) client.DivineOption {
	optionalValue := ctx.Value(ServiceDivineOptions)
	switch typedValue := optionalValue.(type) {
	case client.DivineOption: return typedValue
	case string: return client.WithNames(deityName, typedValue)
	default: return client.WithNames(deityName, humanName)
	}
}

type iGodAuth struct {
	// use Basic, for now
	admin map[string]string
	users map[string]string
}

func setupDashboard(admin *gin.RouterGroup, r *iGodAuth) {
	admin.GET("/", func(c *gin.Context) {
		c.IndentedJSON(http.StatusNonAuthoritativeInfo, gin.H{
			"users": reflect.ValueOf(r.users).MapKeys(),
		})
	})
}

func createAuth() *iGodAuth {
	adminUsername := "admin"
	adminPassword, _ := os.LookupEnv("ADMIN_PASS")
	a := &iGodAuth{
		admin: make(map[string]string, 1),
		users: make(map[string]string, 1),
	}
	if len(adminPassword) == 0 {
		sha := sha256.New()
		seed, _ := os.LookupEnv("ADMIN_SEED")
		sha.Write([]byte(seed))
		if len(seed) == 0 {
			bytes := make([]byte, sha.Size())
			rand.Read(bytes)
			sha.Write(bytes)
		}
		adminPassword = base64.URLEncoding.EncodeToString(sha.Sum(nil))
		log.Println("--- HTTP server test credentials...")
		log.Println(" * ", adminUsername, ":", adminPassword)
	}
	a.admin[adminUsername] = adminPassword
	a.users[adminUsername] = adminPassword
	return a
}

func createEngine() *gin.Engine {
	r := createAuth()
	// TODO: integrate Redis and Twilio if possible (text with iGod)
	// in prod, we'll use Redis to store User objects, otherwise only allow "proxy auth" to OpenAI
	// in prod, we'll set OPENAI_SECRET=key:... otherwise, every request has to specify the API key
	u, ok := os.LookupEnv("OPENAI_SECRET")
	if ok && strings.HasPrefix(u, "key:") {
		u = strings.TrimPrefix(u, "key:")
		u = strings.TrimSpace(u) // no padding
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.Default() // log and recovery
	engine.Use(gin.BasicAuthForRealm(r.users, fmt.Sprintf("%s users", ServiceDeity)))
	engine.Use(func(c *gin.Context) {
		key := strings.TrimPrefix(c.GetHeader("ProxyOpenAI-Key"), "Bearer ")
		if len(key) == 0 {
			key = u // fall back to the secret we have from env, which is empty by default
		}
		if len(key) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing ProxyOpenAI-Key header",
			})
		}
	})
	allow := gin.BasicAuthForRealm(r.users, fmt.Sprintf("%s admin", ServiceDeity))
	admin := engine.Group("/admin", func(c *gin.Context) {
		allow(c) // TODO: consider mechanism to access dashboard
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": errors.New("access denied"),
		})
	})
	pprof.RouteRegister(admin)
	setupDashboard(admin, r)
	return engine
}

func (god *iGodService) ListenAndServe(ctx context.Context, addr string) error {
	god.engine.Use(func(c *gin.Context) {
		c.Set(ServiceDeity, god)
	})
	api := god.engine.Group("/api", cors.Default())
	api.POST("/form", func(c *gin.Context) {
		in, ok := c.GetPostForm("input")
		if !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "missing input text",
			})
		}
		god.speaks.Speak(ctx, in)
	})
	god.AddSpeaker(func(ctx context.Context, in string) client.Edict {
		return client.SimpleEdict(in)
	})
	return god.engine.Run(addr)
}

func (god *iGodService) AddSpeaker(fn client.SpeakerFunc) Service {
	god.speaks.Add(fn)
	return god
}

func (god *iGodService) Test(ctx context.Context) ServiceEdict {
	// TODO: obtain the test phrase from the context, or reject
	return god.speaks.Speak(ctx, "Hello, are you there?")
}
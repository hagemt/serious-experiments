package server

import (
		"crypto/rand"
		"crypto/sha256"
		"encoding/base64"
		"fmt"
		"github.com/gin-contrib/cors"
		"github.com/gin-contrib/pprof"
		"github.com/gin-gonic/gin"
		"log"
		"net/http"
		"os"
		"reflect"
		"strings"
)

type iGodAuth struct {
		// use Basic, for now
		admin map[string]string
		users map[string]string
}

func setupDashboard(admin *gin.RouterGroup, r *iGodAuth) {
		// TODO: serve some kind of web front-end on /ui
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
						bs := sha.Size()
						bytes := make([]byte, bs)
						if _, err := rand.Read(bytes); err != nil {
								panic(err)
						}
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

func createEngine(god *iGodService) *gin.Engine {
		r := createAuth()
		// TODO: integrate Redis and Twilio if possible (text with iGod)
		// in prod, we'll use Redis to store User objects, otherwise only allow "proxy auth" to OpenAI
		// in prod, we'll set OPENAI_PROXY=key:... otherwise, every request has to specify the API key
		u, ok := os.LookupEnv("OPENAI_PROXY")
		if ok && strings.HasPrefix(u, "key:") {
				u = strings.TrimPrefix(u, "key:")
				u = strings.TrimSpace(u) // no padding
				gin.SetMode(gin.ReleaseMode)
		}
		v := "ProxyOpenAI-Key"
		engine := gin.Default() // with log and recovery middleware
		engine.Use(gin.BasicAuthForRealm(r.users, fmt.Sprintf("%s users", ServiceDeity)))
		engine.Use(func(c *gin.Context) {
				key := strings.TrimPrefix(c.GetHeader(v), "Bearer ")
				if len(key) == 0 {
						key = u // fall back to the secret we have from env, which is empty by default
				}
				if len(key) == 0 {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
								"error": fmt.Sprintf("missing %s header", v),
						})
				}
		})
		admin := engine.Group("/admin")
		admin.Use(gin.BasicAuthForRealm(r.admin, fmt.Sprintf("%s admin", ServiceDeity)))
		pprof.RouteRegister(admin)
		setupDashboard(admin, r)

		api := engine.Group("/api", cors.Default())
		api.Use(func(c *gin.Context) {
				c.Set(ServiceDeity, god)
		})
		api.POST("/form", handleFormInput)
		return engine
}

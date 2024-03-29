package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/gin-gonic/gin"
	"github.com/hagemt/serious_experiments/gpt/iGod/client"
)

type iGodService struct {
	client *http.Client
	edicts client.Speaker
	engine *gin.Engine
	values context.Context
	Service
}

func ListenAndServe(ctx context.Context, addr string) error {
	service := &iGodService{}
	service.client = &http.Client{Timeout: time.Second * 15}
	service.edicts = client.NewDeity(extractNames(ctx))
	service.engine = createEngine(service)
	service.values = context.WithValue(ctx, ServiceDeity, service)
	return service.ListenAndServe(ctx, addr)
}

func extractOptionalString(fromArgs context.Context, propertyName, defaultValue string) string {
	if arg := fromArgs.Value(propertyName); arg != nil {
		return arg.(string)
	}
	return defaultValue
}

func extractNames(ctx context.Context) client.DivineOption {
	deityName := extractOptionalString(ctx, "deity-name", "iGod")
	humanName := extractOptionalString(ctx, "human-name", "Human")
	optionalValue := ctx.Value(ServiceDivineOptions)
	switch typedValue := optionalValue.(type) {
	case client.DivineOption:
		return typedValue
	case string:
		return client.WithNames(deityName, typedValue)
	default:
		return client.WithNames(deityName, humanName)
	}
}

func (god *iGodService) gpt(ctx context.Context) client.SpeakerFunc {
	alg := extractOptionalString(ctx, "openai-engine", "")
	secret := extractOptionalString(ctx, "openai-key", "")
	gpt := gpt3.NewClient(secret, gpt3.WithHTTPClient(god.client))
	return func(ctx context.Context, prompt string) client.Edict {
		var risky float32 = 0.5 // TODO: vary with request? (and limit tokens)
		c, err := gpt.CompletionWithEngine(ctx, alg, gpt3.CompletionRequest{
			MaxTokens:   gpt3.IntPtr(1000),
			Prompt:      []string{strings.TrimSpace(prompt)},
			Stop:        []string{"."},
			Temperature: &risky,
		})
		if err != nil {
			return client.FailedEdict(err)
		}
		return client.SimpleEdict(c.Choices[0].Text)
	}
}

func (god *iGodService) ListenAndServe(ctx context.Context, addr string) error {
	speaker := god.gpt(ctx)
	god.AddSpeaker(speaker)
	return god.engine.Run(addr)
}

func (god *iGodService) AddSpeaker(fn client.SpeakerFunc) Service {
	god.edicts.Add(fn)
	return god
}

func (god *iGodService) Test(ctx context.Context) ServiceEdict {
	return god.edicts.Speak(ctx, "Hello, are you there?")
}

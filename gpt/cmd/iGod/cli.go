package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/PullRequestInc/go-gpt3"
	"github.com/boynton/repl"
	dotenv "github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

func envString(propertyName, defaultValue string) string {
	if val, ok := os.LookupEnv(propertyName); ok {
		return val
	}
	return defaultValue
}

func seed(ai gpt3.Client, options ...divineOption) *iGod {
	i := newDeity(options...)
	// the REPL calls this with context, and expects an edict
	i.Speak = func(ctx context.Context, prompt string) edict {
		var risky float32 = 0.5 // 0 = most conservative
		input := strings.TrimSpace(prompt)
		req := gpt3.CompletionRequest{
			//Echo:             false,
			//FrequencyPenalty: 0,
			//LogProbs:         nil,
			//N:                nil,
			//PresencePenalty:  0,
			//Stream:           false,
			//TopP:             nil,
			MaxTokens:   gpt3.IntPtr(1000),
			Prompt:      []string{input},
			Stop:        []string{".", "!", "?"},
			Temperature: &risky,
		}
		//log.Println(req)
		c, err := ai.Completion(ctx, req)
		//log.Println(c)
		if err != nil {
			log.Println(err)
			return failedEdict(errors.New("please try again"))
		}
		first := c.Choices[0]
		log.Println(first.FinishReason, first.Text)
		output := strings.TrimSpace(first.Text)
		output = strings.ReplaceAll(output, "\n", " ")
		output = strings.ReplaceAll(output, "  ", ". ")
		return simpleEdict(fmt.Sprintf("%s.", strings.TrimSuffix(output, ".")))
	}
	return i
}

func die(err error) {
	if err != nil {
		//panic(err)
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func setup() (*iGod, error) {
	home := "/"
	if dir, err := os.UserHomeDir(); err != nil || len(dir) > 0 {
		home = envString("AI_HOME_DIR", dir)
	}
	log.Println(home) // purple?
	_ = dotenv.Load(path.Join(home, ".iGod"))
	_ = dotenv.Overload() // will use .env if present
	ans := struct{ Deity, Engine, Human, Key, Org, URL string }{
		Deity:  envString("DEITY_NAME", "iGod"),
		Engine: envString("AI_ENGINE_ID", "davinci-instruct-beta"),
		Human:  envString("HUMAN_NAME", "Human"),
		Key:    envString("OPEN_AI_KEY", ""),
		//Org:    cmd.EnvString("OPEN_AI_ORG", ""),
		//URL:    cmd.EnvString("OPEN_AI_URL", ""),
	}

	// prompt for text and anything missing
	qs := make([]*survey.Question, 0, 2)
	if len(ans.Key) > 0 {
		fmt.Println("Your eager offering pleases", ans.Deity)
	} else {
		qs = append(qs, &survey.Question{
			Name: "Key",
			Prompt: &survey.Password{
				Message: "What is the divine secret?",
				Help:    "All mystery aside, we need the OpenAI engine to power iGod. (your API key, please)",
			},
			Validate: survey.Required,
		})
	}
	qs = append(qs, &survey.Question{
		Name: "Human",
		Prompt: &survey.Input{
			Default: ans.Human,
			Help:    "a humble moniker",
			Message: "What is your name?",
		},
		Transform: survey.Title,
		Validate:  survey.Required,
	})
	die(survey.Ask(qs, &ans))
	fmt.Println(ans.Deity, "is almost ready to speak with you,", ans.Human) // yellow?

	// test that GTP will work
	if len(ans.Key) == 0 {
		return nil, errors.New("missing secret")
	}
	gpt := gpt3.NewClient(ans.Key, gpt3.WithDefaultEngine(ans.Engine))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := gpt.Engine(ctx, ans.Engine); err != nil {
		return nil, fmt.Errorf("fatal: %v", err)
	}
	return seed(gpt, withNames(ans.Deity, ans.Human)), nil
}

func main() {
	app := &cli.App{}
	app.Setup()
	god, err := setup()
	die(err)
	die(repl.REPL(god))
}

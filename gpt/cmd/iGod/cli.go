package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/PullRequestInc/go-gpt3"
	"github.com/boynton/repl"
	"github.com/hagemt/bijection/gpt/cmd/iGod/client"
	"github.com/hagemt/bijection/gpt/cmd/iGod/server"
	dotenv "github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var iGodVersion string

type gptEngine struct {
	apiClient  gpt3.Client
	apiTimeout time.Duration
	textEcho   bool
	userAgent  string
}

func seed(ai *gptEngine, i client.Speaker) client.Speaker {
	// the REPL calls this with context, and expects an edict
	i.Add(func(ctx context.Context, prompt string) client.Edict {
		// TODO: if the user seems hostile or confused, adjust "temperature"
		var risky float32 = 0.5 // 0 = most conservative (sent as temperature)
		// TODO: aggregate back-and-forth so that "AI has more context" b/t reqs
		input := strings.TrimSpace(prompt)
		if strings.HasPrefix(input, ".echo ") {
			b := strings.TrimPrefix(input, ".echo ")
			c, err := strconv.ParseBool(b)
			if err != nil {
				// TODO: print help text instead of error
				return client.FailedEdict(err)
			}
			ai.textEcho = c
			return client.SimpleEdict(fmt.Sprintf(".echo %t", ai.textEcho))
		}

		// pure I/O to GTP3 is okay, but it requires a network connection
		req := gpt3.CompletionRequest{
			//Echo:             false,
			//FrequencyPenalty: 0,
			//LogProbs:         nil,
			//N:                nil,
			//PresencePenalty:  0,
			//TopP:             nil,
			MaxTokens:   gpt3.IntPtr(1000),
			Prompt:      []string{input},
			Stop:        []string{"."},
			Temperature: &risky,
		}
		//log.Println(req)
		c, err := ai.apiClient.Completion(ctx, req)
		//log.Println(c)
		if err != nil {
			log.Println(err)
			return client.FailedEdict(errors.New("please try again"))
		}
		// just pick the first choice (probably best option)
		first := c.Choices[0]
		if ai.textEcho {
			// use ".echo true" or false to toggle log
			log.Println(first.FinishReason, first.Text)
		}

		// output transformations make it a better conversation partner
		output := strings.TrimSpace(first.Text)
		output = strings.ReplaceAll(output, "\n", " ")
		output = strings.ReplaceAll(output, "  ", ". ")
		output = strings.ReplaceAll(output, "!.", "!")
		output = strings.ReplaceAll(output, "?.", "?")
		output = strings.ReplaceAll(output, "..", ".")
		// FIXME: some of these transforms don't always make sense
		if strings.HasSuffix(output, "?") {
			// questions from the AI are always good for conversation
			return client.SimpleEdict(output)
		}
		if len(output) < 2 {
			// probably hit "stop" because the input didn't end in punctuation
			return client.SimpleEdict("How rude; please ask questions!")
		}
		// make the deity more likely to sound like it's stating facts:
		output = fmt.Sprintf("%s.", strings.TrimSuffix(output, "."))
		return client.SimpleEdict(output)
	})
	return i
}

func setup(c *cli.Context) (client.Speaker, error) {
	ans := struct{ Debug, Deity, Engine, Human, Key, Org, URL string }{
		Debug:  c.String("debug"),
		Deity:  c.String("deity-name"),
		Engine: c.String("openai-engine"),
		Human:  c.String("human-name"),
		Key:    c.String("openai-key"),
		Org:    c.String("openai-org"),
		URL:    c.String("openai-url"),
	}

	// prompt for text and anything missing
	qs := make([]*survey.Question, 0, 2)
	if len(ans.Key) > 0 {
		fmt.Println("Your eager offering pleases the", ans.Deity)
	} else {
		qs = append(qs, &survey.Question{
			Name: "Key",
			Prompt: &survey.Password{
				Message: "What is the divine secret?",
				Help:    fmt.Sprintf("All mystery aside, we need the %s OpenAI engine. (API key)", ans.Engine),
			},
			Validate: survey.Required,
		})
	}
	if ans.Human != "Human" {
			// skip prompt if known already
	} else {
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
	}
	if err := survey.Ask(qs, &ans); err != nil {
		return nil, err
	}
	fmt.Println(ans.Deity, "is almost ready to speak with you,", ans.Human) // yellow?

	// test that GTP will work
	if len(ans.Key) == 0 {
		return nil, errors.New("missing secret")
	}
	gpt := &gptEngine{
		textEcho:   isDebug(ans.Debug, "echo"),
		apiTimeout: c.Duration("openai-sla"),
		userAgent:  c.String("user-agent"),
	}
	if gpt.apiTimeout == 0 {
		gpt.apiTimeout = time.Second * 60
	}
	if gpt.userAgent == "" {
		gpt.userAgent = fmt.Sprintf("iGod/%s", iGodVersion)
	}
	// TODO: use org option, http Client?
	gpt.apiClient = gpt3.NewClient(
		ans.Key,
		gpt3.WithBaseURL(ans.URL),
		gpt3.WithDefaultEngine(ans.Engine),
		gpt3.WithTimeout(gpt.apiTimeout),
		gpt3.WithUserAgent(gpt.userAgent))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if _, err := gpt.apiClient.Engine(ctx, ans.Engine); err != nil {
		//log.Println("fatal:", gpt, err)
		return nil, err
	}
	god := seed(gpt, client.NewDeity(client.WithNames(ans.Deity, ans.Human)))
	return god, nil
}

var debugFlags map[string][]string

func isDebug(flag, key string) bool {
	if debugFlags != nil {
		// TODO: what is a better way to do this?
		// --debug=* means echo AND loud (as does true)
		// --debug=false means neither (opposite of ^^)
		// either --debug=echo and --debug=loud
		// how to support --debug=echo,loud?
	} else {
		debugFlags = make(map[string][]string, 6)
		debugFlags["*"] = []string{"echo", "loud"}
		debugFlags["echo"] = []string{"echo"}
		debugFlags["false"] = []string{}
		debugFlags["loud"] = []string{"loud"}
		debugFlags["none"] = debugFlags["false"]
		debugFlags["true"] = debugFlags["*"]
	}
	if values, ok := debugFlags[key]; ok {
		for _, value := range values {
			if value == flag {
				return true
			}
		}
	}
	return false
}

func main() {
	if iGodVersion == "" {
		iGodVersion = "development"
	}
	home := "/" // TODO: consider using CLI context to load .env
	if dir, _ := os.UserHomeDir(); len(dir) > 0 {
		home = dir // ignores err
	}
	log.Println(home) // purple?
	_ = dotenv.Load(path.Join(home, ".iGod"))
	_ = dotenv.Overload() // will use .env if present
	deityBrain := "davinci-instruct-beta"
	deityName := server.ServiceDeity

	// run client or server, resp.
	app := &cli.App{
		Usage: "speak with AI in a simple REPL (read, eval, print loop) or via HTTP requests",
		Name:  deityName,
		Flags: []cli.Flag{
			&cli.StringFlag{
				DefaultText: deityName,
				EnvVars:     []string{"DEITY_NAME"},
				Name:        "deity-name",
				Usage:       "for blasphemers to override the holy moniker",
				Value:       deityName,
			},
			&cli.StringFlag{
				DefaultText: "false",
				EnvVars:     []string{"IGOD_DEBUG"},
				Name:        "debug",
				Usage:       "set --debug=true for verbose logging",
				Value:       "none",
			},
			&cli.StringFlag{
				DefaultText: "none",
				EnvVars:     []string{"HTTP_ADDR"},
				Name:        "http-addr",
				Usage:       "network port and/or address for HTTP (vs. REPL in shell)",
			},
			&cli.StringFlag{
				DefaultText: "will prompt",
				EnvVars:     []string{"HUMAN_NAME"},
				Name:        "human-name",
				Usage:       "specify a user's name upfront",
				Value:       "Human",
			},
			&cli.StringFlag{
				DefaultText: deityBrain,
				EnvVars:     []string{"OPENAI_ENGINE"},
				Name:        "openai-engine",
				Usage:       "specify a given OpenAI engine; optional",
				Value:       deityBrain,
			},
			&cli.StringFlag{
				DefaultText: "none",
				EnvVars:     []string{"OPENAI_KEY"},
				Name:        "openai-key",
				Usage:       "specify your OpenAI API key; required",
			},
			&cli.StringFlag{
				DefaultText: "none",
				EnvVars:     []string{"OPENAI_ORG"},
				Name:        "openai-org",
				Usage:       "specify your OpenAI organization ID; optional",
			},
			&cli.StringFlag{
				DefaultText: "none",
				EnvVars:     []string{"OPENAI_URL"},
				Name:        "openai-url",
				Usage:       "specify a base URL, for OpenAI APIs; optional",
				Value:       "https://api.openai.com/v1",
			},
		},
		Action: func(c *cli.Context) error {
			i, err := setup(c)
			if err != nil {
				return err
			}
			if addr := c.String("http-addr"); len(addr) > 0 {
				return server.ListenAndServe(c.Context, addr)
			}
			return repl.REPL(i)
		},
	}
	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

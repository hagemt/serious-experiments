package textsynth

import (
	"os"
	"time"
)

var Defaults = struct {
	BaseURL     string
	EngineName  string
	Key         string
	MaxWaitTime time.Duration
	UserAgent   string
}{
	BaseURL:     envString(envKeyBase, "https://api.textsynth.com"),
	EngineName:  envString(envKeyDefaultEngineName, "gptj_6B"),
	Key:         envString(envKey, ""),
	MaxWaitTime: envParseDuration(envKeyMaxTime, time.Minute),
	UserAgent:   envString(envKeyUserAgent, "go-textsynth/v0.1"),
	/*
	   gptj_6B: GPT-J is a language model with 6 billion parameters trained on the Pile (825 GB of text data) published by EleutherAI. Its main language is English but it is also fluent in several other languages. It is also trained on several computer languages.
	   boris_6B: Boris is a fine tuned version of GPT-J for the French language. Use this model is you want the best performance with the French language.
	   fairseq_gpt_13B: Fairseq GPT 13B is an English language model with 13 billion parameters. Its training corpus is less diverse than GPT-J but it has better performance at least on pure English language tasks.
	   gptneox_20B: GPT-NeoX-20B is the largest publically available English language model with 20 billion parameters. It was trained on the same corpus as GPT-J.
	   codegen_6B_mono: CodeGen-6B-mono is a 6 billion parameter model specialized to generate source code. It was mostly trained on Python code.

	   // ^ general, vs. specialized engines:

	   m2m100_1_2B: M2M100 1.2B is a 1.2 billion parameter language model specialized for translation. It supports multilingual translation between 100 languages. See the translate endpoint.
	   stable_diffusion: Stable Diffusion is a 1 billion parameter text to image model trained to generate 512x512 pixel images from English text (sd-v1-4.ckpt checkpoint). See the text_to_image endpoint. There are specific use restrictions associated with this model.
	*/
}

func envString(name, defaultValue string) string {
	if s, ok := os.LookupEnv(name); ok {
		return s
	}
	return defaultValue
}

func envParseDuration(envVarName string, defaultValue time.Duration) time.Duration {
	if s := envString(envVarName, ""); s != "" {
		t, err := time.ParseDuration(s)
		if err != nil {
			panic(err)
		}
		return t
	}
	return defaultValue
}

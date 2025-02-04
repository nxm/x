package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"git.jakub.app/jakub/X/cmd/layla/modules/discord"
	"git.jakub.app/jakub/X/internal/env"
	"git.jakub.app/jakub/X/internal/llm"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

const (
	PROMPT_FORMAT = `Translate to Polish and return JSON. For sentences (has spaces): {"en":"input","pl":"translation","type":"sentence"}. For single words (no spaces): {"en":"word","pl":"translation","type":"noun|verb|etc","example":"Usage example","phonetic":"IPA","use_frequency":0-1,"difficulty":"easy|medium|hard"}. Input: %INPUT%`
	OPENAI_MODEL  = "gpt-3.5-turbo"
)

var (
	OPENAI_API_KEY  = env.GetEnv("OPENAI_API_KEY", "")
	NOCODB_API_KEY  = env.GetEnv("NOCODB_API_KEY", "")
	NOCODB_BASE_URL = env.GetEnv("NOCODB_BASE_URL", "")
)

type svc struct {
	discordModule *discord.Discord
	llmClient     *llm.Client
}

func run() (*svc, error) {
	var err error
	svc := svc{}
	svc.discordModule, err = discord.New()
	if err != nil {
		log.Error().Err(err).Msg("can't load discord module")
		return nil, err
	}

	openAIProvider := llm.NewOpenAIProvider(OPENAI_API_KEY, OPENAI_MODEL)
	svc.llmClient = llm.NewClient(openAIProvider)

	return &svc, nil
}

func main() {
	svc, err := run()
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello from vocabulary memorizer")
	})

	e.POST("/translate", func(c echo.Context) error {
		type TranslateRequest struct {
			Input string `json:"input"`
		}
		var req TranslateRequest
		if err := c.Bind(&req); err != nil {
			log.Error().Err(err).Msg("can't bind request")
			return c.JSON(http.StatusBadRequest, "invalid request")
		}

		finalPrompt := strings.Replace(PROMPT_FORMAT, "%INPUT%", req.Input, -1)
		completionRequest := llm.CompletionRequest{
			Messages: []llm.Message{
				{Role: "user", Content: finalPrompt},
			},
			MaxTokens:   100,
			Temperature: 0.7,
		}

		resp, err := svc.llmClient.Complete(context.Background(), completionRequest)
		if err != nil {
			log.Error().Err(err).Msgf("can't send the completion from llm client")
			return c.JSON(http.StatusInternalServerError, "internal server error")
		}

		err = nocoDbUpload([]byte(resp.Content))
		if err != nil {
			log.Error().Err(err).Msg("can't insert a new entry to nocodb")
			return c.JSON(http.StatusInternalServerError, "internal server error")
		}

		return c.String(http.StatusOK, resp.Content)
	})

	if err := e.Start(":1323"); err != nil {
		panic(err)
	}
}

func nocoDbUpload(data []byte) error {
	u, err := url.Parse(NOCODB_BASE_URL)
	if err != nil {
		log.Error().Err(err).Msg("can't parse url")
		return err
	}

	type Payload struct {
		En           string  `json:"en"`
		Pl           string  `json:"pl"`
		Type         string  `json:"type"`
		Example      string  `json:"example"`
		Phonetic     string  `json:"phonetic"`
		UseFrequency float64 `json:"use_frequency"`
		Difficulty   string  `json:"difficulty"`
	}

	var payload Payload
	err = json.Unmarshal(data, &payload)
	if err != nil {
		return err
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("xc-token", NOCODB_API_KEY)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("nocodb returned %d status code", resp.StatusCode)
    }

	return nil
}

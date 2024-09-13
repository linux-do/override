package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"golang.org/x/net/http2"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const DefaultInstructModel = "gpt-3.5-turbo-instruct"

const StableCodeModelPrefix = "stable-code"

const DeepSeekCoderModel = "deepseek-coder"

type config struct {
	Bind                 string            `json:"bind"`
	ProxyUrl             string            `json:"proxy_url"`
	Timeout              int               `json:"timeout"`
	CodexApiBase         string            `json:"codex_api_base"`
	CodexApiKey          string            `json:"codex_api_key"`
	CodexApiOrganization string            `json:"codex_api_organization"`
	CodexApiProject      string            `json:"codex_api_project"`
	CodexMaxTokens       int               `json:"codex_max_tokens"`
	CodeInstructModel    string            `json:"code_instruct_model"`
	ChatApiBase          string            `json:"chat_api_base"`
	ChatApiKey           string            `json:"chat_api_key"`
	ChatApiOrganization  string            `json:"chat_api_organization"`
	ChatApiProject       string            `json:"chat_api_project"`
	ChatMaxTokens        int               `json:"chat_max_tokens"`
	ChatModelDefault     string            `json:"chat_model_default"`
	ChatModelMap         map[string]string `json:"chat_model_map"`
	ChatLocale           string            `json:"chat_locale"`
	AuthToken            string            `json:"auth_token"`
}

func readConfig() *config {
	content, err := os.ReadFile("config.json")
	if nil != err {
		log.Fatal(err)
	}

	_cfg := &config{}
	err = json.Unmarshal(content, &_cfg)
	if nil != err {
		log.Fatal(err)
	}

	v := reflect.ValueOf(_cfg).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tag := t.Field(i).Tag.Get("json")
		if tag == "" {
			continue
		}

		value, exists := os.LookupEnv("OVERRIDE_" + strings.ToUpper(tag))
		if !exists {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(value)
		case reflect.Bool:
			if boolValue, err := strconv.ParseBool(value); err == nil {
				field.SetBool(boolValue)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
				field.SetInt(intValue)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if uintValue, err := strconv.ParseUint(value, 10, 64); err == nil {
				field.SetUint(uintValue)
			}
		case reflect.Float32, reflect.Float64:
			if floatValue, err := strconv.ParseFloat(value, field.Type().Bits()); err == nil {
				field.SetFloat(floatValue)
			}
		}
	}
	if _cfg.CodeInstructModel == "" {
		_cfg.CodeInstructModel = DefaultInstructModel
	}

	if _cfg.CodexMaxTokens == 0 {
		_cfg.CodexMaxTokens = 500
	}

	if _cfg.ChatMaxTokens == 0 {
		_cfg.ChatMaxTokens = 4096
	}

	return _cfg
}

func getClient(cfg *config) (*http.Client, error) {
	transport := &http.Transport{
		ForceAttemptHTTP2: true,
		DisableKeepAlives: false,
	}

	err := http2.ConfigureTransport(transport)
	if nil != err {
		return nil, err
	}

	if "" != cfg.ProxyUrl {
		proxyUrl, err := url.Parse(cfg.ProxyUrl)
		if nil != err {
			return nil, err
		}

		transport.Proxy = http.ProxyURL(proxyUrl)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(cfg.Timeout) * time.Second,
	}

	return client, nil
}

func abortCodex(c *gin.Context, status int) {
	c.Header("Content-Type", "text/event-stream")

	c.String(status, "data: [DONE]\n")
	c.Abort()
}

func closeIO(c io.Closer) {
	err := c.Close()
	if nil != err {
		log.Println(err)
	}
}

type ProxyService struct {
	cfg    *config
	client *http.Client
}

func NewProxyService(cfg *config) (*ProxyService, error) {
	client, err := getClient(cfg)
	if nil != err {
		return nil, err
	}

	return &ProxyService{
		cfg:    cfg,
		client: client,
	}, nil
}
func AuthMiddleware(authToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if token != authToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s *ProxyService) InitRoutes(e *gin.Engine) {
	e.GET("/_ping", s.pong)
	e.GET("/models", s.models)
	e.GET("/v1/models", s.models)
	authToken := s.cfg.AuthToken // replace with your dynamic value as needed
	if authToken != "" {
		// 鉴权
		v1 := e.Group("/:token/v1/", AuthMiddleware(authToken))
		{
			v1.POST("/chat/completions", s.completions)
			v1.POST("/engines/copilot-codex/completions", s.codeCompletions)

			v1.POST("/v1/chat/completions", s.completions)
			v1.POST("/v1/engines/copilot-codex/completions", s.codeCompletions)
		}
	} else {
		e.POST("/v1/chat/completions", s.completions)
		e.POST("/v1/engines/copilot-codex/completions", s.codeCompletions)

		e.POST("/v1/v1/chat/completions", s.completions)
		e.POST("/v1/v1/engines/copilot-codex/completions", s.codeCompletions)
	}
}

type Pong struct {
	Now    int    `json:"now"`
	Status string `json:"status"`
	Ns1    string `json:"ns1"`
}

func (s *ProxyService) pong(c *gin.Context) {
	c.JSON(http.StatusOK, Pong{
		Now:    time.Now().Second(),
		Status: "ok",
		Ns1:    "200 OK",
	})
}

func (s *ProxyService) models(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": []gin.H{
			{
				"capabilities": gin.H{
					"family":    "gpt-3.5-turbo",
					"limits":    gin.H{"max_prompt_tokens": 12288},
					"object":    "model_capabilities",
					"supports":  gin.H{"tool_calls": true},
					"tokenizer": "cl100k_base",
					"type":      "chat",
				},
				"id":      "gpt-3.5-turbo",
				"name":    "GPT 3.5 Turbo",
				"object":  "model",
				"version": "gpt-3.5-turbo-0613",
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-3.5-turbo",
					"limits":    gin.H{"max_prompt_tokens": 12288},
					"object":    "model_capabilities",
					"supports":  gin.H{"tool_calls": true},
					"tokenizer": "cl100k_base",
					"type":      "chat",
				},
				"id":      "gpt-3.5-turbo-0613",
				"name":    "GPT 3.5 Turbo",
				"object":  "model",
				"version": "gpt-3.5-turbo-0613",
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-4",
					"limits":    gin.H{"max_prompt_tokens": 20000},
					"object":    "model_capabilities",
					"supports":  gin.H{"tool_calls": true},
					"tokenizer": "cl100k_base",
					"type":      "chat",
				},
				"id":      "gpt-4",
				"name":    "GPT 4",
				"object":  "model",
				"version": "gpt-4-0613",
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-4",
					"limits":    gin.H{"max_prompt_tokens": 20000},
					"object":    "model_capabilities",
					"supports":  gin.H{"tool_calls": true},
					"tokenizer": "cl100k_base",
					"type":      "chat",
				},
				"id":      "gpt-4-0613",
				"name":    "GPT 4",
				"object":  "model",
				"version": "gpt-4-0613",
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-4-turbo",
					"limits":    gin.H{"max_prompt_tokens": 20000},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "tool_calls": true},
					"tokenizer": "cl100k_base",
					"type":      "chat",
				},
				"id":      "gpt-4-0125-preview",
				"name":    "GPT 4 Turbo",
				"object":  "model",
				"version": "gpt-4-0125-preview",
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-4o",
					"limits":    gin.H{"max_prompt_tokens": 20000},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":      "gpt-4o",
				"name":    "GPT 4o",
				"object":  "model",
				"version": "gpt-4o-2024-05-13",
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-4o",
					"limits":    gin.H{"max_prompt_tokens": 20000},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":      "gpt-4o-2024-05-13",
				"name":    "GPT 4o",
				"object":  "model",
				"version": "gpt-4o-2024-05-13",
			},
			{
				"capabilities": gin.H{
					"family":    "gpt-4o",
					"limits":    gin.H{"max_prompt_tokens": 20000},
					"object":    "model_capabilities",
					"supports":  gin.H{"parallel_tool_calls": true, "tool_calls": true},
					"tokenizer": "o200k_base",
					"type":      "chat",
				},
				"id":     "gpt-4-o-preview",
				"name":   "GPT 4o",
				"object": "model",
			},
			{
				"capabilities": gin.H{
					"family":    "text-embedding-ada-002",
					"limits":    gin.H{"max_inputs": 256},
					"object":    "model_capabilities",
					"supports":  gin.H{},
					"tokenizer": "cl100k_base",
					"type":      "embeddings",
				},
				"id":      "text-embedding-ada-002",
				"name":    "Embedding V2 Ada",
				"object":  "model",
				"version": "text-embedding-ada-002",
			},
			{
				"capabilities": gin.H{
					"family":    "text-embedding-3-small",
					"limits":    gin.H{"max_inputs": 256},
					"object":    "model_capabilities",
					"supports":  gin.H{"dimensions": true},
					"tokenizer": "cl100k_base",
					"type":      "embeddings",
				},
				"id":      "text-embedding-3-small",
				"name":    "Embedding V3 small",
				"object":  "model",
				"version": "text-embedding-3-small",
			},
			{
				"capabilities": gin.H{
					"family":    "text-embedding-3-small",
					"object":    "model_capabilities",
					"supports":  gin.H{"dimensions": true},
					"tokenizer": "cl100k_base",
					"type":      "embeddings",
				},
				"id":      "text-embedding-3-small-inference",
				"name":    "Embedding V3 small (Inference)",
				"object":  "model",
				"version": "text-embedding-3-small",
			},
		},
		"object": "list",
	})
}

func (s *ProxyService) completions(c *gin.Context) {
	ctx := c.Request.Context()

	body, err := io.ReadAll(c.Request.Body)
	if nil != err {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	model := gjson.GetBytes(body, "model").String()
	if mapped, ok := s.cfg.ChatModelMap[model]; ok {
		model = mapped
	} else {
		model = s.cfg.ChatModelDefault
	}
	body, _ = sjson.SetBytes(body, "model", model)

	if !gjson.GetBytes(body, "function_call").Exists() {
		messages := gjson.GetBytes(body, "messages").Array()
		for i, msg := range messages {
			toolCalls := msg.Get("tool_calls").Array()
			if len(toolCalls) == 0 {
				body, _ = sjson.DeleteBytes(body, fmt.Sprintf("messages.%d.tool_calls", i))
			}
		}
		lastIndex := len(messages) - 1
		if !strings.Contains(messages[lastIndex].Get("content").String(), "Respond in the following locale") {
			locale := s.cfg.ChatLocale
			if locale == "" {
				locale = "zh_CN"
			}
			body, _ = sjson.SetBytes(body, "messages."+strconv.Itoa(lastIndex)+".content", messages[lastIndex].Get("content").String()+"Respond in the following locale: "+locale+".")
		}
	}

	body, _ = sjson.DeleteBytes(body, "intent")
	body, _ = sjson.DeleteBytes(body, "intent_threshold")
	body, _ = sjson.DeleteBytes(body, "intent_content")

	if int(gjson.GetBytes(body, "max_tokens").Int()) > s.cfg.ChatMaxTokens {
		body, _ = sjson.SetBytes(body, "max_tokens", s.cfg.ChatMaxTokens)
	}

	proxyUrl := s.cfg.ChatApiBase + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, proxyUrl, io.NopCloser(bytes.NewBuffer(body)))
	if nil != err {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.ChatApiKey)
	if "" != s.cfg.ChatApiOrganization {
		req.Header.Set("OpenAI-Organization", s.cfg.ChatApiOrganization)
	}
	if "" != s.cfg.ChatApiProject {
		req.Header.Set("OpenAI-Project", s.cfg.ChatApiProject)
	}

	resp, err := s.client.Do(req)
	if nil != err {
		if errors.Is(err, context.Canceled) {
			c.AbortWithStatus(http.StatusRequestTimeout)
			return
		}

		log.Println("request conversation failed:", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer closeIO(resp.Body)

	if resp.StatusCode != http.StatusOK { // log
		body, _ := io.ReadAll(resp.Body)
		log.Println("request completions failed:", string(body))

		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	c.Status(resp.StatusCode)

	contentType := resp.Header.Get("Content-Type")
	if "" != contentType {
		c.Header("Content-Type", contentType)
	}

	_, _ = io.Copy(c.Writer, resp.Body)
}

func (s *ProxyService) codeCompletions(c *gin.Context) {
	ctx := c.Request.Context()

	time.Sleep(200 * time.Millisecond)
	if ctx.Err() != nil {
		abortCodex(c, http.StatusRequestTimeout)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if nil != err {
		abortCodex(c, http.StatusBadRequest)
		return
	}

	body = ConstructRequestBody(body, s.cfg)

	proxyUrl := s.cfg.CodexApiBase + "/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, proxyUrl, io.NopCloser(bytes.NewBuffer(body)))
	if nil != err {
		abortCodex(c, http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.CodexApiKey)
	if "" != s.cfg.CodexApiOrganization {
		req.Header.Set("OpenAI-Organization", s.cfg.CodexApiOrganization)
	}
	if "" != s.cfg.CodexApiProject {
		req.Header.Set("OpenAI-Project", s.cfg.CodexApiProject)
	}

	resp, err := s.client.Do(req)
	if nil != err {
		if errors.Is(err, context.Canceled) {
			abortCodex(c, http.StatusRequestTimeout)
			return
		}

		log.Println("request completions failed:", err.Error())
		abortCodex(c, http.StatusInternalServerError)
		return
	}
	defer closeIO(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Println("request completions failed:", string(body))

		abortCodex(c, resp.StatusCode)
		return
	}

	c.Status(resp.StatusCode)

	contentType := resp.Header.Get("Content-Type")
	if "" != contentType {
		c.Header("Content-Type", contentType)
	}

	_, _ = io.Copy(c.Writer, resp.Body)
}

func ConstructRequestBody(body []byte, cfg *config) []byte {
	body, _ = sjson.DeleteBytes(body, "extra")
	body, _ = sjson.DeleteBytes(body, "nwo")
	body, _ = sjson.SetBytes(body, "model", cfg.CodeInstructModel)

	if int(gjson.GetBytes(body, "max_tokens").Int()) > cfg.CodexMaxTokens {
		body, _ = sjson.SetBytes(body, "max_tokens", cfg.CodexMaxTokens)
	}

	if strings.Contains(cfg.CodeInstructModel, StableCodeModelPrefix) {
		return constructWithStableCodeModel(body)
	} else if strings.HasPrefix(cfg.CodeInstructModel, DeepSeekCoderModel) {
		if gjson.GetBytes(body, "n").Int() > 1 {
			body, _ = sjson.SetBytes(body, "n", 1)
		}
	}

	if strings.HasSuffix(cfg.ChatApiBase, "chat") {
		// @Todo  constructWithChatModel
		// 如果code base以chat结尾则构建chatModel，暂时没有好的prompt
	}

	return body
}

func constructWithStableCodeModel(body []byte) []byte {
	suffix := gjson.GetBytes(body, "suffix")
	prompt := gjson.GetBytes(body, "prompt")
	content := fmt.Sprintf("<fim_prefix>%s<fim_suffix>%s<fim_middle>", prompt, suffix)

	// 创建新的 JSON 对象并添加到 body 中
	messages := []map[string]string{
		{
			"role":    "user",
			"content": content,
		},
	}
	return constructWithChatModel(body, messages)
}

func constructWithChatModel(body []byte, messages interface{}) []byte {

	body, _ = sjson.SetBytes(body, "messages", messages)

	// fmt.Printf("Request Body: %s\n", body)
	// 2. 将转义的字符替换回原来的字符
	jsonStr := string(body)
	jsonStr = strings.ReplaceAll(jsonStr, "\\u003c", "<")
	jsonStr = strings.ReplaceAll(jsonStr, "\\u003e", ">")
	return []byte(jsonStr)
}

func main() {
	cfg := readConfig()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	proxyService, err := NewProxyService(cfg)
	if nil != err {
		log.Fatal(err)
		return
	}

	proxyService.InitRoutes(r)

	err = r.Run(cfg.Bind)
	if nil != err {
		log.Fatal(err)
		return
	}

}

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/linux-do/tiktoken-go"
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

const INSTRUCT_MODEL = "gpt-3.5-turbo-instruct"

type GPTMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type StreamResponse struct {
	Response string `json:"response"`
}
type Message struct {
	Role    string  `json:"role,omitempty"`
	Content any     `json:"content,omitempty"`
	Name    *string `json:"name,omitempty"`
}
type ChatCompletionsStreamResponseChoice struct {
	Index        int     `json:"index"`
	Delta        Message `json:"delta"`
	FinishReason *string `json:"finish_reason,omitempty"`
}

type ChatCompletionsStreamResponse struct {
	Id      string                                `json:"id"`
	Object  string                                `json:"object"`
	Created int64                                 `json:"created"`
	Model   string                                `json:"model"`
	Choices []ChatCompletionsStreamResponseChoice `json:"choices"`
}
type CustomEvent struct {
	Event string
	Id    string
	Retry uint
	Data  interface{}
}
type stringWriter interface {
	io.Writer
	writeString(string) (int, error)
}

type stringWrapper struct {
	io.Writer
}

var dataReplacer = strings.NewReplacer(
	"\n", "\ndata:",
	"\r", "\\r")
var contentType = []string{"text/event-stream"}
var noCache = []string{"no-cache"}

func (w stringWrapper) writeString(str string) (int, error) {
	return w.Writer.Write([]byte(str))
}
func checkWriter(writer io.Writer) stringWriter {
	if w, ok := writer.(stringWriter); ok {
		return w
	} else {
		return stringWrapper{writer}
	}
}
func encode(writer io.Writer, event CustomEvent) error {
	w := checkWriter(writer)
	return writeData(w, event.Data)
}
func writeData(w stringWriter, data interface{}) error {
	dataReplacer.WriteString(w, fmt.Sprint(data))
	if strings.HasPrefix(data.(string), "data") {
		w.writeString("\n\n")
	}
	return nil
}
func (r CustomEvent) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	return encode(w, r)
}

func (r CustomEvent) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	header["Content-Type"] = contentType

	if _, exist := header["Cache-Control"]; !exist {
		header["Cache-Control"] = noCache
	}
}

func GetTimestamp() int64 {
	return time.Now().Unix()
}
func StreamResponseCloudflare2OpenAI(cloudflareResponse *StreamResponse) *ChatCompletionsStreamResponse {
	var choice ChatCompletionsStreamResponseChoice
	choice.Delta.Content = cloudflareResponse.Response
	choice.Delta.Role = "assistant"
	openaiResponse := ChatCompletionsStreamResponse{
		Object:  "chat.completion.chunk",
		Choices: []ChatCompletionsStreamResponseChoice{choice},
		Created: GetTimestamp(),
	}
	return &openaiResponse
}
func StreamResponse2OpenAI(cloudflareResponse *StreamResponse) *ChatCompletionsStreamResponse {
	var choice ChatCompletionsStreamResponseChoice
	choice.Delta.Content = cloudflareResponse.Response
	choice.Delta.Role = "assistant"
	openaiResponse := ChatCompletionsStreamResponse{
		Object:  "chat.completion.chunk",
		Choices: []ChatCompletionsStreamResponseChoice{choice},
		Created: GetTimestamp(),
	}
	return &openaiResponse
}
func SetEventStreamHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
}

const (
	RequestIdKey = "X-Oneapi-Request-Id"
)

func GetResponseID(c *gin.Context) string {
	logID := c.GetString(RequestIdKey)
	return fmt.Sprintf("chatcmpl-%s", logID)
}

type config struct {
	Bind                 string            `json:"bind"`
	ProxyUrl             string            `json:"proxy_url"`
	Timeout              int               `json:"timeout"`
	CodexApiBase         string            `json:"codex_api_base"`
	CodexApiKey          string            `json:"codex_api_key"`
	CodexApiOrganization string            `json:"codex_api_organization"`
	CodexApiProject      string            `json:"codex_api_project"`
	CodexMaxTokens       int               `json:"codex_max_tokens"`
	CodexModelDefault    string            `json:"codex_model_default"`
	ChatApiBase          string            `json:"chat_api_base"`
	ChatApiKey           string            `json:"chat_api_key"`
	ChatApiOrganization  string            `json:"chat_api_organization"`
	ChatApiProject       string            `json:"chat_api_project"`
	ChatModelDefault     string            `json:"chat_model_default"`
	ChatModelMap         map[string]string `json:"chat_model_map"`
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
	cfg       *config
	client    *http.Client
	tokenizer *tiktoken.Tiktoken
}

func NewProxyService(cfg *config) (*ProxyService, error) {
	client, err := getClient(cfg)
	if nil != err {
		return nil, err
	}

	tokenizer, err := tiktoken.EncodingForModel(INSTRUCT_MODEL)
	if nil != err {
		return nil, err
	}

	return &ProxyService{
		cfg:       cfg,
		client:    client,
		tokenizer: tokenizer,
	}, nil
}

func (s *ProxyService) InitRoutes(e *gin.Engine) {
	e.POST("/v1/chat/completions", s.completions)
	e.POST("/v1/engines/copilot-codex/completions", s.codeCompletions)
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
	body, _ = sjson.DeleteBytes(body, "intent")

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

	time.Sleep(100 * time.Millisecond)
	if ctx.Err() != nil {
		abortCodex(c, http.StatusRequestTimeout)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if nil != err {
		abortCodex(c, http.StatusBadRequest)
		return
	}

	prompt := gjson.GetBytes(body, "prompt").String()
	suffix := gjson.GetBytes(body, "suffix").String()
	inputTokens := len(s.tokenizer.Encode(prompt, nil, nil))
	suffixTokens := len(s.tokenizer.Encode(suffix, nil, nil))
	outputTokens := int(gjson.GetBytes(body, "max_tokens").Int())

	totalTokens := inputTokens + suffixTokens + outputTokens
	if totalTokens > s.cfg.CodexMaxTokens { // reduce
		left, right := 0, len(prompt)
		for left < right {
			mid := (left + right) / 2
			subPrompt := prompt[mid:]
			subInputTokens := len(s.tokenizer.Encode(subPrompt, nil, nil))
			totalTokens = subInputTokens + suffixTokens + outputTokens
			if totalTokens > s.cfg.CodexMaxTokens {
				left = mid + 1
			} else {
				right = mid
			}
		}

		body, _ = sjson.SetBytes(body, "prompt", prompt[left:])
	}

	body, _ = sjson.DeleteBytes(body, "extra")
	body, _ = sjson.DeleteBytes(body, "nwo")
	var model string
	proxyUrl := s.cfg.CodexApiBase
	if s.cfg.CodexModelDefault == "" || s.cfg.CodexModelDefault == INSTRUCT_MODEL {
		model = INSTRUCT_MODEL
		proxyUrl = proxyUrl + "/completions"
	} else {
		model = s.cfg.CodexModelDefault
	}
	body, _ = sjson.SetBytes(body, "model", model)
	if model == "deepseek-coder" {
		message := gjson.GetBytes(body, "prompt").String()
		body, _ = sjson.DeleteBytes(body, "prompt")
		msg := make([]GPTMessage, 0)
		msg = append(msg, GPTMessage{Role: "system", Content: "You are a helpful assistant"})
		msg = append(msg, GPTMessage{Role: "user", Content: message})
		body, _ = sjson.SetBytes(body, "messages", msg)
		body, _ = sjson.DeleteBytes(body, "n")
	} else if strings.HasPrefix(model, "@") {
		proxyUrl = s.cfg.CodexApiBase
		message := gjson.GetBytes(body, "prompt").String()
		body, _ = sjson.DeleteBytes(body, "prompt")
		msg := make([]GPTMessage, 0)
		msg = append(msg, GPTMessage{Role: "system", Content: ""})
		msg = append(msg, GPTMessage{Role: "user", Content: message})
		body, _ = sjson.SetBytes(body, "messages", msg)
		body, _ = sjson.DeleteBytes(body, "n")
	}
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

	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	dataChan := make(chan string)
	stopChan := make(chan bool)
	go func() {
		for scanner.Scan() {
			data := scanner.Text()
			if len(data) < len("data: ") {
				continue
			}
			data = strings.TrimPrefix(data, "data: ")
			dataChan <- data
		}
		stopChan <- true
	}()
	SetEventStreamHeaders(c)
	id := GetResponseID(c)
	responseModel := c.GetString("original_model")
	var responseText string
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			// some implementations may add \r at the end of data
			data = strings.TrimSuffix(data, "\r")
			var codeResponse StreamResponse
			err := json.Unmarshal([]byte(data), &codeResponse)
			if err != nil {
				if data == "[DONE]" {
					return true
				}
				log.Println("error unmarshalling stream response: ", err.Error())
				return true
			}
			if model != INSTRUCT_MODEL {
				response := StreamResponseCloudflare2OpenAI(&codeResponse)
				if response == nil {
					return true
				}
				responseText += codeResponse.Response
				response.Id = id
				response.Model = responseModel
				jsonStr, err := json.Marshal(response)
				if err != nil {
					log.Println("error marshalling stream response: ", err.Error())
					return true
				}
				c.Render(-1, CustomEvent{Data: "data: " + string(jsonStr)})
			} else {
				c.Render(-1, CustomEvent{Data: "data:" + string(data)})
			}
			return true
		case <-stopChan:
			c.Render(-1, CustomEvent{Data: "data: [DONE]"})
			return false
		}
	})
	_ = resp.Body.Close()
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

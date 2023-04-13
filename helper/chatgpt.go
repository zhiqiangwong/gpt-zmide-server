/*
 * @Author: Wzq
 * @Date: 2023-04-06
 * @FilePath: /gpt-zmide-server/helper/chatgpt.go
 */
package helper

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

// 构造请求体
type ChatRequest struct {
	Model            string         `json:"model"`
	Messages         []*ChatMessage `json:"messages"`
	User             string         `json:"user"`
	Stream           bool           `json:"stream"`
	Temperature      float64        `json:"temperature,omitempty"`
	MaxTokens        int            `json:"max_tokens,omitempty"`
	FrequencyPenalty float64        `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64        `json:"presence_penalty,omitempty"`
	TopP             float64        `json:"top_p,omitempty"`
	Raw              bool           `json:"-"`
	RawBody          []byte         `json:"-"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatChoices struct {
	Message        *ChatMessage `json:"message"`
	Index          int          `json:"index"`
	FinishReason   string       `json:"finish_reason"`
	CompletionTime int64        `json:"completion_time"`
}

// openAi 返回结构体
type OpenAIResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Usage   struct {
		PromptTokens     int64 `json:"prompt_tokens"`
		CompletionTokens int64 `json:"completion_tokens"`
		TotalTokens      int64 `json:"total_tokens"`
	} `json:"usage"`
	Choices []*ChatChoices `json:"choices"`
	Raw     string         `json:"-"`
}

// openAi 返回结构体(Stream)
type OpenAIResponseStream struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Raw string `json:"-"`
}

// 调用 chatGPT
func ChatGptAsk(req ChatRequest, streamCall ...func(line *OpenAIResponseStream)) (res *OpenAIResponse, err error) {
	client, err := Config.GetOpenAIHttpClient()
	if err != nil {
		return
	}

	c := client.R()

	// 流式返回时指定参数
	if len(streamCall) > 0 {
		req.Stream = true
		c = c.SetDoNotParseResponse(true)
	}

	var bodyStr []byte

	if req.RawBody == nil {
		bodyStr, err = json.Marshal(req)
		if err != nil {
			return
		}
	} else {
		bodyStr = req.RawBody
	}

	// 发送请求
	resp, err := c.SetBody(bodyStr).Post("/v1/chat/completions")

	if len(streamCall) == 0 {
		if err = json.Unmarshal(resp.Body(), &res); err != nil {
			return
		}

		res.Raw = string(resp.Body())
	} else {
		// 逐行读取响应结果
		reader := bufio.NewReader(resp.RawBody())
		defer resp.RawBody().Close()

		res = &OpenAIResponse{}
		message := &ChatMessage{
			Role:    "",
			Content: "",
		}
		var choices = []*ChatChoices{}

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					err = nil
					break
				}
				break
			}

			var resStream *OpenAIResponseStream

			// 是否指定原样返回
			if !req.Raw {
				// 去掉 data: 前缀
				jsonData := strings.TrimSpace(strings.TrimPrefix(string(line), "data:"))

				if jsonData == "[DONE]" {
					break
				}

				err = json.Unmarshal([]byte(jsonData), &resStream)

				if resStream != nil {
					if resStream.Choices[0].Delta.Role != "" {
						message.Role = resStream.Choices[0].Delta.Role
					}

					if resStream.Choices[0].Delta.Content != "" {
						message.Content += resStream.Choices[0].Delta.Content
					}

					res.Raw += jsonData
					res.ID += resStream.ID
					res.Model += resStream.Model
					res.Object += resStream.Object
					res.Created += resStream.Created
				}
			} else {
				resStream = &OpenAIResponseStream{}
				resStream.Raw = string(line)
			}

			if resStream != nil {
				streamCall[0](resStream)
			}
		}

		if res.ID != "" {
			choices = append(choices, &ChatChoices{
				Message:      message,
				Index:        0,
				FinishReason: "stop",
			})
			res.Choices = choices
		}
	}

	return
}

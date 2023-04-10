/*
 * @Author: Bin
 * @Date: 2023-03-06
 * @FilePath: /gpt-zmide-server/models/chat.go
 */
package models

import (
	"errors"
	"fmt"
	"gpt-zmide-server/helper"
)

type Chat struct {
	ID          uint             `gorm:"primaryKey" json:"id"`
	AppID       uint             `json:"-"`
	Remark      string           `json:"remark"`
	Messages    []*Message       `gorm:"foreignKey:ChatID" json:"messages"`
	Application *ChatApplication `gorm:"foreignKey:AppID" json:"app"`
	Model       string           `json:"model"`
	MessageChan chan *Message    `gorm:"-" json:"-"`
	BaseModel
}

type ChatApplication struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	AppSecret        string `json:"-"`
	AppKey           string `json:"-"`
	Status           uint   `json:"-"`
	EnableFixLongMsg uint   `json:"-"`
}

func (chat *Chat) QueryChatGPT(stream bool) (msg *Message, err error) {

	model := chat.Model
	secret_key := helper.Config.OpenAI.SecretKey
	if model == "" || secret_key == "" {
		return nil, errors.New("OpenAI model 未设置或 secret_key 未设置")
	}

	if len(chat.Messages) < 1 {
		return nil, errors.New("chat messages 处理异常")
	}

	// 处理会话数据
	var msgsTmp = []*helper.ChatMessage{}
	msgCount := 0
	// 倒序遍历消息记录
	for i := len(chat.Messages) - 1; i >= 0; i-- {
		item := chat.Messages[i]
		contextCount := msgCount + len(item.Content)
		// 避免消息上下文超过 4600 字数限制
		if contextCount > 4500 {
			// 判断应用是否需要修复长消息
			DB.Preload("Application").Find(chat)
			if chat.Application != nil && chat.Application.EnableFixLongMsg != 1 {
				continue
			} else {
				return nil, errors.New("消息上下文超过 4600 字数限制")
			}
		}
		msgCount = contextCount
		msgsTmp = append(msgsTmp, &helper.ChatMessage{
			Role:    item.Role,
			Content: item.Content,
		})
	}

	// 修正消息顺序
	var msgs = []*helper.ChatMessage{}
	for i := len(msgsTmp) - 1; i >= 0; i-- {
		msgs = append(msgs, msgsTmp[i])
	}

	chatReq := helper.ChatRequest{
		Model:    model,
		Messages: msgs,
		User:     helper.Config.SiteName,
	}

	var res *helper.OpenAIResponse

	if !stream {
		// 请求 openAi
		res, err = helper.ChatGptAsk(chatReq)
	} else {
		chat.MessageChan = make(chan *Message)
		// 以 stream 模式进行请求
		res, err = helper.ChatGptAsk(chatReq, func(line *helper.OpenAIResponseStream) {
			message := &Message{
				ID:      0,
				ChatID:  chat.ID,
				Role:    line.Choices[0].Delta.Role,
				Content: line.Choices[0].Delta.Content,
			}
			// 将消息推送到MessageChannel中
			chat.MessageChan <- message
		})

		close(chat.MessageChan)
	}

	if err != nil {
		return nil, err
	}

	if len(res.Choices) < 1 {
		return nil, errors.New("openai api callback choices data error")
	}

	choiceFirst := res.Choices[0]
	msg = &Message{
		ChatID:  chat.ID,
		Raw:     res.Raw,
		Role:    choiceFirst.Message.Role,
		Content: choiceFirst.Message.Content,
	}

	if err = DB.Create(msg).Error; err != nil {
		fmt.Println("message create error " + err.Error())
	}

	return msg, nil
}

/*
 * @Author: Bin
 * @Date: 2023-03-09
 * @FilePath: /gpt-zmide-server/controllers/apis/open.go
 */
package apis

import (
	"encoding/json"
	"gpt-zmide-server/helper"
	"gpt-zmide-server/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Open struct {
	Controller
}

func (ctl *Open) Index(c *gin.Context) {
	app := c.MustGet(helper.MiddlewareAuthAppKey)
	ctl.Success(c, app)
}

func (ctl *Open) Query(c *gin.Context) {
	appTmp := c.MustGet(helper.MiddlewareAuthAppKey)
	bodyMap := c.GetStringMap(helper.PostBodyKey)

	if appTmp == nil {
		ctl.Fail(c, "应用异常")
		return
	}

	var app *models.Application
	var ok bool
	if app, ok = appTmp.(*models.Application); !ok || app == nil || app.Status != 1 {
		ctl.Fail(c, "应用异常")
		return
	}

	var content, p_chat_id, p_remark, model string

	if bodyMap != nil {
		content = bodyMap["content"].(string)

		if chatID, ok := bodyMap["chat_id"].(string); ok {
			p_chat_id = chatID
		}

		if remark, ok := bodyMap["remark"].(string); ok {
			p_remark = remark
		}

		if p_model, ok := bodyMap["model"].(string); ok {
			model = p_model
		}
	} else {
		content = c.PostForm("content")
		p_chat_id = c.PostForm("chat_id")
		p_remark = c.PostForm("remark")
		model = c.PostForm("model")
	}

	// content 参数为必传
	if content == "" {
		ctl.Fail(c, "参数异常1")
		return
	}

	// 当 model 不存在时，使用配置的默认 model
	if model == "" {
		model = helper.Config.OpenAI.Model
	}

	chat := &models.Chat{}

	chat.Model = model

	if p_chat_id != "" {
		if id, err := strconv.Atoi(p_chat_id); err == nil && id != 0 {
			// 当 chat_id 合法时，去数据库查找 chat
			chat.ID = uint(id)
			if err := models.DB.First(chat).Error; err != nil || chat.AppID != app.ID {
				ctl.Fail(c, "chat_id 不合法")
				return
			}
			chat.AppID = app.ID
		}
	}

	if chat.AppID == 0 {
		chat.AppID = app.ID
		if err := models.DB.Create(chat).Error; err != nil {
			ctl.Fail(c, "chat 处理异常")
			return
		}
	}

	// 当 remark 参数存在时更新 chat remark
	if p_remark != "" {
		chat.Remark = p_remark
		models.DB.Updates(chat)
	}

	message := &models.Message{
		ChatID:  chat.ID,
		Role:    "user",
		Content: content,
		Raw:     "",
	}

	if err := models.DB.Create(message).Error; err != nil {
		ctl.Fail(c, "消息处理失败")
		return
	}

	// 刷新 chat Messages
	models.DB.Preload("Messages").Find(chat)

	callback, err := chat.QueryChatGPT(false)
	if err != nil {
		ctl.Fail(c, err.Error())
		return
	}

	ctl.Success(c, callback)
}

func (ctl *Open) Chat(c *gin.Context) {
	appTmp := c.MustGet(helper.MiddlewareAuthAppKey)
	bodyMap := c.GetStringMapString(helper.PostBodyKey)

	if appTmp == nil {
		ctl.Fail(c, "应用异常")
		return
	}

	var app *models.Application
	var ok bool
	if app, ok = appTmp.(*models.Application); !ok || app == nil || app.Status != 1 {
		ctl.Fail(c, "应用异常")
		return
	}

	content := c.DefaultPostForm("content", bodyMap["content"])
	p_chat_id := c.DefaultPostForm("chat_id", bodyMap["chat_id"])
	p_remark := c.DefaultPostForm("remark", bodyMap["remark"])
	model := c.DefaultPostForm("model", bodyMap["model"])

	// content 参数为必传
	if content == "" {
		ctl.Fail(c, "参数异常")
		return
		// content = "你好"
	}

	// 当 model 不存在时，使用配置的默认 model
	if model == "" {
		model = helper.Config.OpenAI.Model
	}

	chat := &models.Chat{}

	chat.Model = model

	if p_chat_id != "" {
		if id, err := strconv.Atoi(p_chat_id); err == nil && id != 0 {
			// 当 chat_id 合法时，去数据库查找 chat
			chat.ID = uint(id)
			if err := models.DB.First(chat).Error; err != nil || chat.AppID != app.ID {
				ctl.Fail(c, "chat_id 不合法")
				return
			}
			chat.AppID = app.ID
		}
	}

	if chat.AppID == 0 {
		chat.AppID = app.ID
		if err := models.DB.Create(chat).Error; err != nil {
			ctl.Fail(c, "chat 处理异常")
			return
		}
	}

	// 当 remark 参数存在时更新 chat remark
	if p_remark != "" {
		chat.Remark = p_remark
		models.DB.Updates(chat)
	}

	message := &models.Message{
		ChatID:  chat.ID,
		Role:    "user",
		Content: content,
		Raw:     "",
	}

	if err := models.DB.Create(message).Error; err != nil {
		ctl.Fail(c, "消息处理失败")
		return
	}

	// 刷新 chat Messages
	models.DB.Preload("Messages").Find(chat)

	// 设置流式响应头
	c.Header("Content-Type", "text/event-stream;charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // 禁用nginx缓冲

	go func() {
		for data := range chat.MessageChan {
			jsonStr, _ := json.Marshal(gin.H{"status": "ok", "code": 200, "data": data})
			c.Writer.WriteString("data: " + string(jsonStr) + "\n")
			c.Writer.Flush()
		}
	}()

	callback, err := chat.QueryChatGPT(true)
	if err != nil {
		ctl.Fail(c, err.Error())
		return
	}

	//最后一次输出不需要输出完整内容
	callback.Content = ""
	jsonStr, _ := json.Marshal(gin.H{"status": "ok", "code": 200, "data": callback})
	c.Writer.WriteString("data: " + string(jsonStr) + "\n")
	c.Writer.WriteString("data: " + "[DONE]" + "\n")
}

// 按 openai 原格式
func (ctl *Open) ChatRaw(c *gin.Context) {
	appTmp := c.MustGet(helper.MiddlewareAuthAppKey)

	if appTmp == nil {
		ctl.Fail(c, "应用异常")
		return
	}

	var app *models.Application
	var ok bool
	if app, ok = appTmp.(*models.Application); !ok || app == nil || app.Status != 1 {
		ctl.Fail(c, "应用异常")
		return
	}

	var reqBody []byte
	var err error
	var bodyMap map[string]any

	// 兼容加密数据
	bodyMap = c.GetStringMap(helper.PostBodyKey)
	if bodyMap != nil {
		reqBody, err = json.Marshal(bodyMap)
	} else {
		//获取请求内容
		rawBody, err := c.GetRawData()
		if err != nil {
			ctl.Fail(c, "请求参数异常")
			return
		}

		json.Unmarshal(rawBody, &bodyMap)
	}

	if bodyMap == nil {
		ctl.Fail(c, "请求参数异常")
		return
	}

	if _, ok := bodyMap["token"]; ok {
		delete(bodyMap, "token")
	}

	reqBody, err = json.Marshal(bodyMap)

	chatReq := &helper.ChatRequest{
		RawBody: reqBody,
		Raw:     true, // 指定结果原样返回
	}

	// 设置流式响应头
	c.Header("Content-Type", "text/event-stream;charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // 禁用nginx缓冲
	c.Header("Access-Control-Allow-Origin", "*")

	// 以 stream 模式进行请求
	_, err = helper.ChatGptAsk(*chatReq, func(line *helper.OpenAIResponseStream) {
		c.Writer.WriteString(line.Raw)
		c.Writer.Flush()
	})

	if err != nil {
		ctl.Fail(c, "请求参数异常")
		return
	}
}

package gpt_3_5_turbo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Params struct {
	// Это будет удалено из запроса перед отправкой в API. Необходимый.
	// This will be stripped from the request before sending to the API. Required.
	API_TOKEN          string      `json:"api_token,omitempty"`
	// Если установлено, API будет удалять новые строки с начала сгенерированного текста. Необязательно, по умолчанию false.
	// If set, the API will strip newlines from the beginning of the generated text. Optional, defaults to false.
	StripNewline       bool        `json:"strip_newline,omitempty"`
	// Тело запроса для отправки в API. Необходимый.
	// The request body to send to the API. Required.
	Request            ChatRequest `json:"request,omitempty"`
	// Если установлено, история сообщений будет храниться в ответе. Необязательно, по умолчанию false.
	// If set, the message history will be kept in the response. Optional, defaults to false.
	KeepMessageHistory bool        `json:"keep_message_history,omitempty"`
	// Используемая история сообщений. Необязательный, по умолчанию равен нулю.
	// The message history to use. Optional, defaults to null.
	MessageHistory     []Message   `json:"message_history,omitempty"`
}

type ChatRequest struct {
	// ID используемой модели. В настоящее время поддерживаются только gpt-3.5-turbo и gpt-3.5-turbo-0301. Необходимый.
	// ID of the model to use. Currently, only gpt-3.5-turbo and gpt-3.5-turbo-0301 are supported. Required.
	Model            string             `json:"model"`
	// Сообщения, для которых генерируются завершения чата, в формате чата. Необходимый.
	// The messages to generate chat completions for, in the chat format. Required.
	Messages         []Message          `json:"messages"`
	// Какую температуру отбора проб использовать, от 0 до 2. Необязательно, по умолчанию 1.
	// What sampling temperature to use, between 0 and 2. Optional, defaults to 1.
	Temperature      float64            `json:"temperature,omitempty"`
	// Альтернатива выборке с температурой, когда модель учитывает результаты токенов с вероятностной массой top_p. Необязательно, по умолчанию 1.
	// An alternative to sampling with temperature, where the model considers the results of the tokens with top_p probability mass. Optional, defaults to 1.
	TopP             float64            `json:"top_p,omitempty"`
	// Сколько вариантов завершения чата генерировать для каждого входного сообщения. Необязательно, по умолчанию 1.
	// How many chat completion choices to generate for each input message. Optional, defaults to 1.
	N                int                `json:"n,omitempty"`
	// *** НЕ РЕАЛИЗОВАНО *** Если установлено, будут отправлены частичные дельты сообщения. Необязательно, по умолчанию false.
	// *** NOT IMPLEMENTED *** If set, partial message deltas will be sent. Optional, defaults to false.
	Stream           bool               `json:"stream,omitempty"`
	// До 4 последовательностей, в которых API перестанет генерировать новые токены. Необязательный, по умолчанию равен нулю.
	// Up to 4 sequences where the API will stop generating further tokens. Optional, defaults to null.
	Stop             interface{}        `json:"stop,omitempty"`
	// Максимальное количество токенов, разрешенное для сгенерированного ответа. Необязательно, по умолчанию инф.
	// The maximum number of tokens allowed for the generated answer. Optional, defaults to inf.
	MaxTokens        int                `json:"max_tokens,omitempty"`
	// Число от -2,0 до 2,0. Положительные значения штрафуют новые токены в зависимости от того, появляются ли они в тексте до сих пор. Необязательно, по умолчанию 0.
	// Number between -2.0 and 2.0. Positive values penalize new tokens based on whether they appear in the text so far. Optional, defaults to 0.
	PresencePenalty  float64            `json:"presence_penalty,omitempty"`
	// Число от -2,0 до 2,0. Положительные значения штрафуют новые токены в зависимости от их существующей частоты в тексте на данный момент. Необязательно, по умолчанию 0.
	// Number between -2.0 and 2.0. Positive values penalize new tokens based on their existing frequency in the text so far. Optional, defaults to 0.
	FrequencyPenalty float64            `json:"frequency_penalty,omitempty"`
	// Изменить вероятность появления указанных токенов в завершении. Необязательный, по умолчанию равен нулю.
	// Modify the likelihood of specified tokens appearing in the completion. Optional, defaults to null.
	LogitBias        map[string]float64 `json:"logit_bias,omitempty"`
	// Уникальный идентификатор, представляющий вашего конечного пользователя. Необязательный.
	// A unique identifier representing your end-user. Optional.
	User             string             `json:"user,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Choices []Choice `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (params *Params) Query(msg string) ([]Choice, error) {
	// Если история сохранена, копируем все сообщения в запрос
	// If history is kept, copy all messages to request
	if params.KeepMessageHistory {
		params.Request.Messages = params.MessageHistory
	}

	// Добавляем сообщение к запросу
	// Append the message to the request
	params.Request.Messages = append(params.Request.Messages, Message{
		Role:    "user",
		Content: msg,
	})

	// Преобразование тела запроса в JSON
	// Convert the request body to JSON
	jsonBody, err := json.Marshal(params.Request)
	if err != nil {
		return nil, err
	}

	// Определяем объект запроса
	// Define the request object
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	// Устанавливаем заголовки
	// Set the headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+params.API_TOKEN)

	// Отправляем запрос
	// Send the request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Если ответ не 200 OK, возвращаем ошибку
	// If response is not 200 OK, return an error
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OpenAI API returned status code %d", resp.StatusCode)
	}

	// Декодируем тело ответа в структуру
	// Decode the response body into a struct
	var responseStruct ChatCompletionResponse
	err = json.NewDecoder(resp.Body).Decode(&responseStruct)
	if err != nil {
		return nil, err
	}

	// Добавляем ответ к сообщению
	// Append response to message
	for i, choice := range responseStruct.Choices {
		// Если StripNewline равно true, удалить 2 последовательных символа новой строки
		// If StripNewline is true, remove the 2 consecutive newlines
		if params.StripNewline {
			if len(choice.Message.Content) < 2 {
				continue
			}

			// Если первые 2 символа являются новой строкой, удаляем их
			// If the first 2 characters are newlines, remove them
			if choice.Message.Content[:2] == "\n\n" {
				choice.Message.Content = choice.Message.Content[2:]
			}
		}

		// Сохраняем измененное сообщение обратно в ответ
		// Store modified message back in response
		responseStruct.Choices[i].Message = choice.Message
	}

	// Если история сохраняется, добавляем ответ в историю сообщений
	// If history is kept, append the response to the message history
	if params.KeepMessageHistory {
		params.MessageHistory = append(params.MessageHistory, Message{
			Role:    "user",
			Content: msg,
		})

		for _, choice := range responseStruct.Choices {
			params.MessageHistory = append(params.MessageHistory, choice.Message)
		}
	}

	return responseStruct.Choices, nil
}

func Init(userParams Params) (*Params, error) {
	params := &Params{}

	// Проверяем, установлен ли токен API
	// Check if API token is set
	if userParams.API_TOKEN == "" {
		return nil, fmt.Errorf("API_TOKEN is not set")
	} else {
		params.API_TOKEN = userParams.API_TOKEN
	}

	// Проверяем, установлена ли модель, если не установлена, то по умолчанию
	// Check if model is set, if not set to default
	if userParams.Request.Model == "" {
		params.Request.Model = "gpt-3.5-turbo"
	} else {
		params.Request.Model = userParams.Request.Model
	}

	// Проверяем, установлен ли KeepMessageHistory, если он не установлен по умолчанию
	// Check if KeepMessageHistory is set, if not set to default
	if userParams.KeepMessageHistory {
		params.KeepMessageHistory = true
	}

	// Проверяем, установлена ли температура, если она не установлена по умолчанию
	// Check if temperature is set, if not set to default
	if userParams.StripNewline {
		params.StripNewline = true
	}

	return params, nil
}

func (params *Params) ClearHistory(msg string) {
	// Просматриваем все params.MessageHistory и удаляем все сообщения
	// Go over all params.MessageHistory and remove all messages
	for i := 0; i < len(params.MessageHistory); i++ {
		params.MessageHistory = append(params.MessageHistory[:i], params.MessageHistory[i+1:]...)
	}
}

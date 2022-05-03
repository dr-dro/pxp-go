package pxp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Hub struct {
	sync.RWMutex
	sendUrl    string
	recieveUrl string
}

func NewHub(sendUrl, recieveUrl string) *Hub {
	return &Hub{
		sendUrl:    sendUrl,
		recieveUrl: recieveUrl,
	}
}

//Структура запроса
//для отправки сообщения
type sendParams struct {
	//отправитель
	From string `json:"from"`
	//адресат
	To string `json:"to"`
	//данные
	Data string `json:"data"`
	//контрольная сумма
	Sign string `json:"sign"`
}

//Структура запроса
//для получения сообщения
type receiveParams struct {
	//адресат
	Recipient string `json:"recipient"`
	//предыдущий токен
	Token string `json:"token"`
	//контрольная сумма
	Sign string `json:"sign"`
}

//Сообщение
type Message struct {
	//номер сообщения в очереди
	Offset int `json:"i"`
	//отправитель сообщения
	From string `json:"from"`
	//тело сообщения
	Data string `json:"data"`
}

//Структура ответа
type resultJSON struct {
	//флаг успешной операции
	Success bool `json:"success"`
	//дополнительная информация
	Info string `json:"info"`
	//цепочка токенов
	Token string `json:"token"`
	//сообщения
	List []Message `json:"list"`
}

//Отправка сообщения "data" от пользователя "from" к пользователю "to".
//Подписывается секретным словом "secret".
//Блокируется до получения ответа.
func (h *Hub) SendMessage(from, to, data, secret string) error {

	h.RLock()
	url := h.sendUrl
	h.RUnlock()

	//сообщение
	var Message = sendParams{
		From: from,
		To:   to,
		Data: data,
		Sign: checkSum(from, to, data, secret),
	}

	//json
	content, _ := json.Marshal(Message)
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(content))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	//отправить post запрос
	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		return error
	}
	defer response.Body.Close()

	if response.StatusCode != 201 {

		//код ответа должен быть 201
		return errors.New("Wrong response status code: " + strconv.Itoa(response.StatusCode))

	} else {

		//тело ответа
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}

		//распарсить JSON
		var reply resultJSON
		err = json.Unmarshal(body, &reply)
		if err != nil {
			return err
		}

		if !reply.Success {
			//хаб не принял запрос
			return errors.New(reply.Info)
		}

	}

	//отправка сообщения подтверждена
	return nil

}

//Отправка сообщения "data" от пользователя "from" к пользователю "to".
//Подписывается секретным словом "secret"
func (h *Hub) ReceiveMessage(ctx context.Context, recipient, secret string) (chan Message, chan error) {

	h.RLock()
	url := h.recieveUrl
	h.RUnlock()

	dispatch := make(chan Message, 1)
	on_error := make(chan error, 1)

	//поток чтения цепочки
	go func() {
		token := ""
		for {
			func() {
				//сообщение
				var Message = receiveParams{
					Recipient: recipient,
					Token:     token,
					Sign:      checkSum(recipient, token, secret),
				}

				//json request
				content, _ := json.Marshal(Message)
				request, _ := http.NewRequest("POST", url, bytes.NewBuffer(content))
				request.Header.Set("Content-Type", "application/json; charset=UTF-8")

				//отправить post запрос
				client := &http.Client{}
				response, error := client.Do(request)
				if error != nil {
					on_error <- error
					return
				}
				defer response.Body.Close()

				if response.StatusCode != 201 {

					//код ответа должен быть 201 Created
					on_error <- errors.New("Wrong response status code: " + strconv.Itoa(response.StatusCode))
					return

				} else {

					//тело ответа
					body, err := ioutil.ReadAll(response.Body)
					if err != nil {
						on_error <- err
						return
					}

					//распарсить JSON
					var reply resultJSON
					err = json.Unmarshal(body, &reply)
					if err != nil {
						on_error <- err
						return
					}

					//обновить токен
					token = reply.Token

					if !reply.Success {
						//хаб не принял запрос
						on_error <- errors.New(reply.Info)
						return
					}

					for _, mes := range reply.List {
						dispatch <- mes
					}
				}
			}()

			pause, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
			select {
			case <-pause.Done():
				//продолжаем читать цепочку сообщений
				cancel()
			case <-ctx.Done():
				//завершение снаружи
				cancel()
				return
			}
		}
	}()

	return dispatch, on_error
}

func checkSum(v ...string) string {
	h := sha256.New()
	for _, s := range v {
		h.Write([]byte(s))
	}
	return hex.EncodeToString(h.Sum(nil))
}

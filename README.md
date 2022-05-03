# pxp-go
Клиент PxP хаба для Go


```go

//Адрес хаба
hub_host = "https://localhost:8000"

//Хаб обмена сообщениями
hub := pxp.NewHub(
  //url отправки сообщений
  hub_host+"/send",
  //url чтения сообщений
  hub_host+"/receive",
)

//Отправка сообщения "Hello!" от user1 к user2
//блокируется до получения ответа
err := hub.SendMessage("user1", "user2", "Hello!", "secretA")

//Создание каналов чтения сообщений и ошибок
dispatch, on_error := hub.ReceiveMessage(ctx, "user2", "secret2")

for {

  select {
  case m := <-dispatch:
    //новое сообщение
    fmt.Println(m.Offset, m.From, m.Data)
  case err := <-on_error:
    fmt.Println("error: ", err.Error())
  case <-ctx.Done():
    return
  }

}
```

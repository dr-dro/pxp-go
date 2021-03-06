# Клиент PxP хаба для Go

```
go get github.com/dr-dro/pxp-go
```

### Создание подключения к хабу
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
```

### Отправка сообщения
```go
//Отправка сообщения "Hello!" от user1 к user2
//блокируется до получения ответа
err := hub.SendMessage("user1", "user2", "Hello!", "secretA")
if err == nil {
  //Сообщение точно не потеряется и 
  //рано или поздно дойдет до адресата
  fmt.Println("done")
}
```

### Чтение сообщений
```go

//контекст для остановки чтения сообщений
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()
  
//Создание каналов чтения сообщений и ошибок
dispatch, on_error := hub.ReceiveMessage(ctx, "user2", "secret2")

//цикл ожидания сообщений
for {

  select {
  case mes := <-dispatch:
    //новое сообщение
    fmt.Println(mes.Offset, mes.From, mes.Data)
  case err := <-on_error:
    //ошибка  
    fmt.Println("error: ", err.Error())
  case <-ctx.Done():
    //завершение  
    return
  }

}
```

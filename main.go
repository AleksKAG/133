package main

import (
	"fmt"
	"net/http"
	"os"

	"main.go/base"
)

/* //еще один возможный вариант
func mainHandle(res http.ResponseWriter, req *http.Request) {
	var filePath string

	if req.URL.Path == "/" {
		filePath = filepath.Join("web", "index.html")
	} else {
		filePath = filepath.Join("web", req.URL.Path)
	}
	http.ServeFile(res, req, filePath)

}*/

const webDir = "./web"

func main() {
	fmt.Println("Запускаем сервер")
	//http.HandleFunc(`/`, mainHandle)
	// Устанавливаем обработчик для корневого URL
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	envPort := os.Getenv("TODO_PORT")
	if envPort == "" {
		envPort = "7540"
	}
	envDBFILE := os.Getenv("TODO_DBFILE")
	base.CreateTable(envDBFILE)
	err := http.ListenAndServe(":"+envPort, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Завершаем работу")
}

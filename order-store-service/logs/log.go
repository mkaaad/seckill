package logs

import (
	"log"
	"order-store/model"
	"os"
)

var file *os.File

func OpenFile() {
	oFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("无法打开日志文件%v\n", err)
	}
	log.SetFlags(log.LstdFlags)
	file = oFile
}
func WriteLog(message error) {
	log.SetOutput(file)
	log.Println(message)
	log.SetOutput(os.Stdout)
}
func WriteData(order model.Order) {
	log.SetOutput(file)
	log.Println(order)
	log.SetOutput(os.Stdout)

}

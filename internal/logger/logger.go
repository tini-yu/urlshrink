package logger

import (
	"log"

	"go.uber.org/zap"
)

var L *zap.Logger
var S *zap.SugaredLogger

//Init(true) если продакшн
func Init(prod bool) {
    var err error
    if prod {
        L, err = zap.NewProduction()
    } else {
        L, err = zap.NewDevelopment()
    }
    if err != nil {
        log.Fatalf("Не получилось инициализировать zap: %v", err)
    }
    S = L.Sugar()
}

func Sync() {
    _ = L.Sync()
}
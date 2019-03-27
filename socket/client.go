package socket

import (
	"log"
	"os/user"
)

func InitClient() Message {
	current_user, err := user.Current()
	if err != nil {
		log.Fatal("ユーザの取得に失敗")
	}
	return Message{
		Name: current_user.Username,
	}
}

package main

import (
	"context"
	"fmt"

	"github.com/leetcode-golang-classroom/golang-order-api/internal/application"
)

func main() {
	app := application.New()
	err := app.Start(context.TODO())
	if err != nil {
		fmt.Println("failed to start app:", err)
	}
}

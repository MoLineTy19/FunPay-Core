package main

import (
	"FunPay-Core/internal/engine"
	"FunPay-Core/internal/fp"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	goldenKey := os.Getenv("FP_GOLDEN_KEY")
	sessionID := os.Getenv("FP_PHPSESSID")
	goldenSeal := os.Getenv("FP_GOLDEN_SEAL")

	fmt.Println("FunPay-Core Starting...")
	ctx := context.Background()
	client := fp.NewClient(goldenKey, sessionID, goldenSeal, 800*time.Millisecond, 600*time.Millisecond)
	account, err := client.GetAccount(ctx)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println(account)

	userID := os.Getenv("FP_USER_ID")
	csrfToken := os.Getenv("FP_CSRF_TOKEN")

	objectTypes := []string{"orders_counters", "chat_bookmarks"}

	runner := fp.NewRunner(client, userID, csrfToken, objectTypes)

	if err := runner.Init(ctx); err != nil {
		fmt.Println("Runner Init error: ", err)
		return
	}

	buf := engine.NewBuffer()

	for {
		ev, err := runner.Poll(ctx)
		if err != nil {
			fmt.Println("Poll error:", err)
			return
		}

		events := engine.WrapEvents(ev)
		buf.Push(events)
		buf.EvictExpired(time.Now())

		for i, e := range events {
			fmt.Printf("[%d] %+v\n", i, e)
		}

		time.Sleep(2 * time.Second)
	}

	//
	//resp, err := runner.Poll(ctx)
	//if err != nil {
	//	fmt.Println("Poll error:", err)
	//	return
	//}
	//
	//fmt.Printf("response=%v, objects=%d\n", resp.Response, len(resp.Objects))
	//for i, raw := range resp.Objects {
	//	fmt.Printf("[%d] %s\n", i, raw)
	//}
}

package worker

import (
	"BE2_API_Shop/shop/internal/data/postgres"
	"BE2_API_Shop/shop/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type worker struct {
	repo *postgres.ProductRepository
}

func NewWorker(repo *postgres.ProductRepository) *worker {
	return &worker{
		repo: repo,
	}
}

func (w *worker) Work(ch *amqp091.Channel, queue amqp091.Queue) {
	ticker := time.NewTicker(1 * time.Second)

	for range ticker.C {
		uuidList, err := w.repo.GetProductIDs(context.Background())
		if err != nil || len(uuidList) < 1 {
			fmt.Println("err in conn to db or len < 1, uuidList:", uuidList)
			continue
		}

		randID := uuidList[rand.Intn(len(uuidList))]
		event := model.ProductUpdateEvent{
			ID:             randID,
			Price:          int64(rand.Intn(1000)),
			AvailableStock: rand.Intn(20),
			Timestamp:      time.Now(),
		}
		body, err := json.Marshal(event)
		if err != nil {
			log.Fatal(err)
		}
		err = ch.PublishWithContext(context.Background(),
			"", queue.Name, false, false,
			amqp091.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp091.Persistent,
				Body:         body,
			})

		if err != nil {
			log.Println("publish:", err)
			continue
		}

		fmt.Println("Published:", string(body))

	}
}

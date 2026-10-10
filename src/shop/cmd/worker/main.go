package main

import (
	"BE2_API_Shop/shop/cfg"
	"BE2_API_Shop/shop/internal/data/postgres"
	"BE2_API_Shop/shop/worker"
	"context"
	"log"
	"os"

	"github.com/rabbitmq/amqp091-go"
)

func main() {
	conf, err := cfg.Load("shop/cfg/config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := amqp091.Dial(conf.RabbitMq.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	queue, err := ch.QueueDeclare(conf.RabbitMq.Queue,
		true, false, false, false, nil)

	dbURL := conf.Shop.Database.URL
	db, err := postgres.NewDatabase(context.Background(), dbURL)
	if err != nil {
		os.Exit(1)
	}

	productRepo := postgres.NewProductRepository(db.Pool)
	wrkr := worker.NewWorker(productRepo)

	wrkr.Work(ch, queue)
}

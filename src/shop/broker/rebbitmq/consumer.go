package rabbitmq

import (
	"BE2_API_Shop/shop/internal/model"
	"BE2_API_Shop/shop/internal/service"
	"BE2_API_Shop/shop/web/realtime/sse"
	websockethub "BE2_API_Shop/shop/web/realtime/websocket"

	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	productService service.ProductService
	conn           *amqp091.Connection
	ch             *amqp091.Channel
	queue          amqp091.Queue
	sseHub         *sse.SSEHub
	wsHub          *websockethub.WebSocketHub
}

func NewConsumer(productService service.ProductService,
	urlAddres string, queueName string, sseHub *sse.SSEHub,
	wsHub *websockethub.WebSocketHub) (*Consumer, error) {

	conn, err := amqp091.Dial(urlAddres)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	queue, err := ch.QueueDeclare(queueName,
		true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		productService: productService,
		conn:           conn,
		ch:             ch,
		queue:          queue,
		sseHub:         sseHub,
		wsHub:          wsHub,
	}, nil
}

func (c *Consumer) Consume() error {
	msg, err := c.ch.Consume(c.queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for delivery := range msg {
		var event model.ProductUpdateEvent
		err := json.Unmarshal(delivery.Body, &event)
		if err != nil {
			log.Println("unmarshal:", err)
			if err := delivery.Nack(false, false); err != nil {
				log.Println("nack:", err)
			}
			continue
		}
		err = c.productService.UpdateProduct(context.Background(), &event)

		if err != nil {
			log.Println("RepoError:", err)
			if err := delivery.Nack(false, false); err != nil {
				log.Println("nack:", err)
			}
			continue
		}
		newProduct, err := c.productService.GetByID(context.Background(), event.ID)
		if err != nil {
			log.Println("get newProduct:", err)
			if err := delivery.Nack(false, false); err != nil {
				log.Println("nack:", err)
			}
			continue
		}

		data, err := json.Marshal(newProduct)
		if err != nil {
			log.Println("update task:", err)
			if err := delivery.Nack(false, false); err != nil {
				log.Println("nack:", err)
			}
			continue
		}

		c.sseHub.Broadcast(data)
		c.wsHub.Broadcast(data)

		fmt.Println("Consumer: product update")
		if err := delivery.Ack(false); err != nil {
			log.Println("ack:", err)

		}
	}
	return nil
}

package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/opentracing/opentracing-go"
	amqp "github.com/rabbitmq/amqp091-go"

	"e-mall/consts"
	"e-mall/repository/db/dao"
	"e-mall/repository/db/model"
	"e-mall/types"
	"e-mall/utils/idgen"
	log "e-mall/utils/log"
)

var (
	consumerChOnce sync.Once
	consumerConnMu sync.Mutex
)

// getConsumerChannel 为消费者打开一个专用 channel（不和生产者共享）
func getConsumerChannel() (*amqp.Channel, error) {
	consumerConnMu.Lock()
	defer consumerConnMu.Unlock()

	if connection == nil {
		InitRabbitMQ()
	}
	if connection == nil {
		return nil, fmt.Errorf("rabbitmq connection unavailable")
	}

	ch, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("open consumer channel: %w", err)
	}
	return ch, nil
}

// StartConsumers 启动所有 RabbitMQ 消费者（在独立 goroutine 中运行）
func StartConsumers(ctx context.Context) {
	consumerChOnce.Do(func() {
		go runOrderPaidConsumer(ctx)
		go runRechargePaidConsumer(ctx)
		go runOrderTimeoutConsumer(ctx)
	})
}

// runOrderPaidConsumer 监听 rabbitmq-order-paid-queue
func runOrderPaidConsumer(ctx context.Context) {
	queue := consts.OrderPaidQueue

	ch, err := getConsumerChannel()
	if err != nil {
		log.LogrusObj.Errorf("order-paid consumer: %v", err)
		return
	}
	defer ch.Close()

	// 声明队列（幂等，生产者那边也会声明，两边都声明保证消费前队列一定存在）
	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		log.LogrusObj.Errorf("order-paid consumer declare queue: %v", err)
		return
	}

	// 每次投递后手动 Ack，不自动确认
	deliveries, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		log.LogrusObj.Errorf("order-paid consumer register: %v", err)
		return
	}

	log.LogrusObj.Infof("order-paid consumer started, waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			log.LogrusObj.Info("order-paid consumer stopped")
			return
		case delivery, ok := <-deliveries:
			if !ok {
				log.LogrusObj.Error("order-paid consumer deliveries channel closed")
				return
			}
			processOrderPaidMessage(ctx, delivery)
		}
	}
}

// processOrderPaidMessage 处理单条订单支付消息
func processOrderPaidMessage(_ context.Context, delivery amqp.Delivery) {
	// 从消息头中提取 Jaeger 追踪上下文
	consumeCtx := extractConsumerContext(delivery.Headers, "rabbitmq.consume."+consts.OrderPaidQueue)

	var event types.OrderPaidEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		log.LogrusObj.Errorf("order-paid deserialize failed: %v", err)
		// 反序列化失败属于永久性错误，不重新入队
		_ = delivery.Nack(false, false)
		return
	}

	tx := &model.PaymentTransaction{
		TransactionNo:   strconv.FormatUint(idgen.NextID(), 10),
		OrderNum:        strconv.FormatUint(event.OrderNum, 10),
		UserID:          event.UserID,
		PayeeID:         &event.BossID,
		Amount:          event.TotalAmount,
		PaymentMethod:   "balance",
		TransactionType: "order_paid",
		Status:          "success",
		PaidAt:          &event.PaidAt,
	}

	if err := dao.NewPaymentTransactionDao(consumeCtx).CreatePaymentTransaction(tx); err != nil {
		log.LogrusObj.Errorf("order-paid save transaction failed: %v", err)
		// DB 写入失败可能是临时性问题，重新入队
		_ = delivery.Nack(false, true)
		return
	}

	if err := delivery.Ack(false); err != nil {
		log.LogrusObj.Errorf("order-paid ack failed: %v", err)
	}
}

// processRechargePaidMessage 处理单条充值支付消息
func processRechargePaidMessage(_ context.Context, delivery amqp.Delivery) {
	consumeCtx := extractConsumerContext(delivery.Headers, "rabbitmq.consume."+consts.RechargePaidQueue)

	var event types.RechargePaidEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		log.LogrusObj.Errorf("recharge-paid deserialize failed: %v", err)
		_ = delivery.Nack(false, false)
		return
	}

	tx := &model.PaymentTransaction{
		TransactionNo:   strconv.FormatUint(idgen.NextID(), 10),
		OrderNum:        event.OrderNum,
		UserID:          event.UserID,
		PayeeID:         nil,
		Amount:          event.Amount,
		PaymentMethod:   event.Channel,
		TransactionType: "recharge",
		Status:          "success",
		PaidAt:          &event.PaidAt,
	}

	if err := dao.NewPaymentTransactionDao(consumeCtx).CreatePaymentTransaction(tx); err != nil {
		log.LogrusObj.Errorf("recharge-paid save transaction failed: %v", err)
		_ = delivery.Nack(false, true)
		return
	}

	if err := delivery.Ack(false); err != nil {
		log.LogrusObj.Errorf("recharge-paid ack failed: %v", err)
	}
}

// runRechargePaidConsumer 监听 rabbitmq-recharge-paid-queue
func runRechargePaidConsumer(ctx context.Context) {
	queue := consts.RechargePaidQueue

	ch, err := getConsumerChannel()
	if err != nil {
		log.LogrusObj.Errorf("recharge-paid consumer: %v", err)
		return
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		log.LogrusObj.Errorf("recharge-paid consumer declare queue: %v", err)
		return
	}

	deliveries, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		log.LogrusObj.Errorf("recharge-paid consumer register: %v", err)
		return
	}

	log.LogrusObj.Infof("recharge-paid consumer started, waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			log.LogrusObj.Info("recharge-paid consumer stopped")
			return
		case delivery, ok := <-deliveries:
			if !ok {
				log.LogrusObj.Error("recharge-paid consumer deliveries channel closed")
				return
			}
			processRechargePaidMessage(ctx, delivery)
		}
	}
}

// runOrderTimeoutConsumer 监听订单超时延迟队列
func runOrderTimeoutConsumer(ctx context.Context) {
	queue := consts.OrderTimeoutQueue

	ch, err := getConsumerChannel()
	if err != nil {
		log.LogrusObj.Errorf("order-timeout consumer: %v", err)
		return
	}
	defer ch.Close()

	// 队列由上游 PublishDelayedJSON 声明，消费侧也声明一次保证幂等
	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		log.LogrusObj.Errorf("order-timeout consumer declare queue: %v", err)
		return
	}

	deliveries, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		log.LogrusObj.Errorf("order-timeout consumer register: %v", err)
		return
	}

	log.LogrusObj.Infof("order-timeout consumer started, waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			log.LogrusObj.Info("order-timeout consumer stopped")
			return
		case delivery, ok := <-deliveries:
			if !ok {
				log.LogrusObj.Error("order-timeout consumer deliveries channel closed")
				return
			}
			processOrderTimeoutMessage(ctx, delivery)
		}
	}
}

// processOrderTimeoutMessage 处理订单超时消息
func processOrderTimeoutMessage(_ context.Context, delivery amqp.Delivery) {
	consumeCtx := extractConsumerContext(delivery.Headers, "rabbitmq.consume."+consts.OrderTimeoutQueue)

	var event types.OrderTimeoutEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		log.LogrusObj.Errorf("order-timeout deserialize failed: %v", err)
		_ = delivery.Nack(false, false)
		return
	}

	// DeleteUnpaidOrderByOrderNum 自带 WHERE type = 1 条件
	// 已支付的订单不会被删除，已删除的订单影响 0 行，均不会报错
	if err := dao.NewOrderDao(consumeCtx).DeleteUnpaidOrderByOrderNum(event.OrderNum); err != nil {
		log.LogrusObj.Errorf("order-timeout close order failed: %v", err)
		_ = delivery.Nack(false, true)
		return
	}

	if err := delivery.Ack(false); err != nil {
		log.LogrusObj.Errorf("order-timeout ack failed: %v", err)
	}
}

// extractConsumerContext 从 AMQP 消息头中提取 Jaeger 追踪上下文
func extractConsumerContext(headers amqp.Table, spanName string) context.Context {
	carrier := opentracing.TextMapCarrier{}
	for key, val := range headers {
		if str, ok := val.(string); ok {
			carrier.Set(key, str)
		}
	}

	tracer := opentracing.GlobalTracer()
	wireContext, err := tracer.Extract(opentracing.TextMap, carrier)
	if err == nil {
		span := tracer.StartSpan(spanName, opentracing.FollowsFrom(wireContext))
		return opentracing.ContextWithSpan(context.Background(), span)
	}

	span := tracer.StartSpan(spanName)
	return opentracing.ContextWithSpan(context.Background(), span)
}

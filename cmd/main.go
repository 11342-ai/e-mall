package main

import (
	"context"
	"fmt"

	conf "e-mall/config"
	"e-mall/repository/cache"
	"e-mall/repository/db/dao"

	// "e-mall/repository/db/dao"
	"e-mall/repository/es"
	"e-mall/repository/kafka"
	"e-mall/repository/rabbitmq"
	"e-mall/routes"
	"e-mall/utils/idgen"
	log "e-mall/utils/log"
	"e-mall/utils/track"

	_ "github.com/apache/skywalking-go"
)

func main() {
	loading()
	r := routes.NewRouter()
	_ = r.Run(conf.Config.System.HttpPort)
	fmt.Println("启动配置成功...")
}

func loading() {
	conf.InitConfig()
	dao.InitMysql()
	cache.InitCache()
	rabbitmq.InitRabbitMQ()
	es.InitES()
	kafka.InitKafka()
	track.InitTrack()
	idgen.Init()

	ctx := context.Background()
	go rabbitmq.StartConsumers(ctx)

	log.InitLogger()
	fmt.Println("加载配置完成...")
}

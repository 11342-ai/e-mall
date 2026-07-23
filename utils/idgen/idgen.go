package idgen

import (
	"time"

	"github.com/bwmarrin/snowflake"

	conf "e-mall/config"
)

var node *snowflake.Node

func Init() error {
	startTime, err := time.Parse("2006-01-02", conf.Config.Snowflake.StartTime)
	if err != nil {
		return err
	}
	snowflake.Epoch = startTime.UnixMilli()

	node, err = snowflake.NewNode(conf.Config.Snowflake.WorkerID)
	return err
}

func NextID() uint64 {
	return uint64(node.Generate().Int64())
}

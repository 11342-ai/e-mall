- 需求 : Kafka DLQ + 退避
- 要解决的问题 : 坏消息阻塞分区、恢复后洪峰、无死信归档
- 核心手段 : 重试上限 + 指数退避 + 死信主题

目前，他这个死信主题缺少一个消费者，用于处理死信消息。
对于这个消费者，它有这样子的几个方案：

|方案 |怎么做 |适合场景 |
| :---: | :----: | :----: |
|纯人工 |运维人员用 Kafka CLI（ kcat 、Kafka UI）查看 DLQ 消息，分析原因后手动重发 | 消息量极少，几小时才出一两条失败 |
|半自动 + 管理后台| 加一个 DLQ 消费者，把失败消息写入数据库 → 管理员后台能看到失败记录，点击"重新处理"按钮| 消息量中低，需要可观测性和控制能力 |
|全自动重试 |DLQ 消费者每隔一段时间（如 30 分钟）重新投递回 flash-sale-orders ，再经历一轮退避重试 | 失败原因是临时性的（如 DB 抖动）|

|失败原因 |举例 |重试有意义吗 |
| :---: | :---: | :----: |
|临时性| DB 连接闪断、事务冲突 |有意义，但重试 3 次 + 退避已经处理过了，到 DLQ 说明 3 次都失败 |
|永久性 |地址已被用户删除、用户ID不存在| 重试 100 次也没用|

所以到了 DLQ 的消息， 绝大多数是永久性失败 ，全自动重试没意义。建议选 方案②（半自动 + 管理后台） ：
1. 新增一个 DLQ 消费者 ，收到消息后写入 failed_flash_sale_orders 表
2. 管理员登录后台，看到失败的订单列表 + 失败原因
3. 管理员判断原因后，点击"重新处理"将消息重新投回 flash-sale-orders

涉及的文件：

|文件| 动作 |
| :---: | :----: |
|repository/db/model/flash_sale.go |新增 FailedFlashSaleOrder 结构体 |
|repository/db/dao/flash_sale.go |新增 DAO：Create / List / Retry |
|types/flash_sale.go |（如果有）新增请求/响应类型 |
|repository/kafka/common.go |新增 consumeDLQ 消费者，启动在 InitKafka 中 |
|consts/product.go |常量已加 |
|api/v1/flash_sale.go |新增管理后台查看/重试接口 |
|service/flash_sale.go |新增重试逻辑 |
|routes/router.go |注册管理员路由|

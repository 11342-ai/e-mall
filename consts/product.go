package consts

const BaseProductPageSize = 15

const FlashSaleQueues = "flash-sale-orders"

// FlashSaleDLQ 秒杀订单死信主题（消费重试耗尽后投递）
const FlashSaleDLQ = "flash-sale-orders-dlq"

const ProductBatchCreate = 1000

const (
	ProductAuditPending  = 0 // 待审核
	ProductAuditApproved = 1 // 已上架
	ProductAuditRejected = 2 // 已拒绝
)

FROM rabbitmq:3-management

# 下载并安装 delayed message exchange 插件（版本需匹配 RabbitMQ 3.x）
ADD https://github.com/rabbitmq/rabbitmq-delayed-message-exchange/releases/download/v3.13.0/rabbitmq_delayed_message_exchange-3.13.0.ez /opt/rabbitmq/plugins/
RUN rabbitmq-plugins enable --offline rabbitmq_delayed_message_exchange

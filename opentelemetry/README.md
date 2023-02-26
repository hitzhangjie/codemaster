# 可观测性

微服务可观测性面临很多挑战, 业界一般通过 metrics+logging+tracing 来解决此难题.

opentelemetry, 已经成为事实上的业界标准, 它定义了很多的规范并开发了SDK, 后端实
现则需要自己开发, 不过现在业界也有很多成熟的产品, 比如lightstep, Elasticsearch
APM等等.

这里主要是结合其提供的SDK来进行一点测试.

# goserver
基于消息队列（nats)的游戏服务器框架

# 架构

## 1,gate服
1，和客户端通信使用websocket，仅支持protobuf，数据结构：<br>

    -------------------------
    | id | protobuf message |
    -------------------------
id是消息名的Hash值，用于标记消息名，反解析数据<br>
因为websocket协议已经支持iframe分帧处理，不需要处理粘包，故包结构中无需包长度字段<br>

## 2，多服rpc通信
使用nats(消息队列)构建服务器间通信，支持send,request,call请求

# 示例
1，启动消息队列服务（https://github.com/nats-io/nats-streaming-server）<br>
2，见example目录，依次启动user/main.go,gate/main.go,client/main.go<br>

# 基于goserver的游戏服务器
avatar-fight-server https://github.com/0990/avatar-fight-server



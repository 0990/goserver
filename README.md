# goserver
简易的游戏服务器框架

# 构建计划

## 第一步，gate服务搭建
1，支持protobuf，websocket,信息包结构初步定义为

    -------------------------
    | id | protobuf message |
    -------------------------
因为websocket协议已经支持iframe分帧处理，不需要处理粘包，故包结构中无需包长度字段 

2，客户端使用cocos creator，改造通信，支持和gate服通信   

## 第二步，多服rpc通信
初步考虑使用nats(消息队列)构建服务器间通信

## 第三步，使用此框架开发一个对战游戏（进行中）
初步考虑将shootgame改造


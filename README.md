Swaggo - `v0.3.0`


### 项目修改自：github.com/teambition/swaggo
### 修改内容：
1. 项目本身支持 go mod
2. 支持使用本工具的项目是 go mod 项目，不用再放到 go path 下了
3. 个性化支持了分页响应，例如 // @Success 200 []schema.DeviceQRCodeResult "{list:列表数据,pagination:{current:页索引,pageSize:页大小,total:总数量}}" 根据描述里含有 pagination 自动转化为分页结构体 


## About

Generate API documentation from annotations in Go code. It's always used for you Go server application.
The swagger file accords to the [Swagger Spec](https://github.com/OAI/OpenAPI-Specification) and displays it using
[Swagger UI](https://github.com/swagger-api/swagger-ui)(this project dosn't provide).

## Quick Start Guide

### Install

```shell
go get -u -v github.com/teambition/swaggo
```


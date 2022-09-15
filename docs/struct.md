# 项目结构

pkg/gateway: 网关核心包

pkg/req: 请求包

## gateway

api: 网关API接口

auth: 后端认证源。目前只有ldap

data：网关需要的结构体

director: 存放了HTTP的认证方法，自定义添加修改此文件

main: gateway包入口

protect：waf实现

token： token管理

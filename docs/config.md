# 配置文件

在程序当前目录创建一个config.yml.目前是写死的。后期优化可以自定义。

```yaml
proxy:
  - app: jenkins #租户代码
    url: https://jenkins.com #租户地址
    authorization: Basic # 租户认证方式
    authorization_url: # 获取租户token的地址
    autorization_domain: @qq.com # 租户的域，一般是在user后面加@xxxx.com
ldap:
  host: # ldap主机
  port:  # ldap端口，暂时不支持加密
  bind_dn: # 管理员账号
  password:  # 管理员密码
  user_base_dn:  # 用户组dn
server:
  port: 8001 # 启动端口
auth:
  secret: xxxx  # token生成混淆字符串
  expire: 72000 # token超时时间，秒
redis:
  addr:  # redis地址，xxxx:port
```
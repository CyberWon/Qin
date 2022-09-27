# 必读说明
除了ldap认证接口，其他接口访问均需要添加header 
> Authorzation: 接口返回地token地址

# LDAP认证

## 请求地址

/auth/ldap

## 请求方法
POST

## 请求参数

|字段   | 类型  | 必填 | 说明   |
| ------------ | ------------ | ------------ | ------------ |
| username  | string  |  是 |  IAM用户名 |
| password   |  string  | 是  | IAM密码  |

## 返回

|字段   | 类型  | 说明 |
| ------------ | ------------ | ------------ |
| code  | int  |  状态码，0代表成功，非0代表异常|
| message   |  string  | 状态信息  |
| token| string| token信息，有效期2个小时|

# private token

一些系统提供的私人token，在HTTP header头部自动注入。

## 请求地址

/user/private_token

## 请求方法
POST

## 请求参数

json

|字段   | 类型  | 必填 | 说明   |
| ------------ | ------------ | ------------ | ------------ |
| app  | string  |  是 |  系统代码 |
| token   |  string  | 是  | 认证方式+private token，获取方式请自行查阅对应系统帮助  |

### token说明
由认证方式+private token组成。例如常见的Bearer，这里的token就设置成
```
Bearer private-token
```

## 返回

|字段   | 类型  | 说明 |
| ------------ | ------------ | ------------ |
| code  | int  |  状态码，0代表成功，非0代表异常|
| message   |  string  | 状态信息  |

# 用户信息

获取用户自身信息（如果需要请自行修改auth,api两个文件。）

## 请求地址

/user

## 请求方法
GET

## 请求参数
无

## 返回

| 字段      | 类型  | 说明               |
|---------| ------------ |------------------|
| code    | int  | 状态码，0代表成功，非0代表异常 |
| message |  string  | 状态信息             |
| data    |  string  | 用户信息             |
| - user  |  string  | 用户名              |
| - name  |  string  | 显示名              |
| - mail  |  string  | 邮箱地址             |





# 刷新token

## 请求地址

/auth/refreshToken

## 请求方法
get

## 返回

|字段   | 类型  | 说明 |
| ------------ | ------------ | ------------ |
| code  | int  |  状态码，0代表成功，非0代表异常|
| message   |  string  | 状态信息  |
| token| string| token信息，有效期2个小时|
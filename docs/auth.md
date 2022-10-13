# 认证方法

## Basic

使用用户名和密码直接登录。

如：jenkins

## PrivateToken
 
私人token。

例如rancher，gitlab这种。

## Token

有登录接口且能返回token的方式。需要在扩展字段里面指定。

例如：A系统登录接口需要user和pass字段。就可以按照下面这样写。返回的token是叫access
```
authorization_ext: user pass jwt access 7200
```
- 第一位: 所在系统的用户名字段。常见的都是username，不排除有一些非得写成user
- 第二位：所在系统的密码字段。常见的都是password，不排除有一些非得写成pass
- 第三位：token类型，一般常见的有Bearer，jwt等
- 第四位：登录成功后返回的token字段。这个各样的都有。
- 第5位：token有效期


## BKPaaS

蓝鲸API官方文档里面的认证方式。
优点：有API文档
缺点：API支持不丰富。

## BK

模拟实现的cookie登录，所有页面上调用的api都可以调用。如果有需要模拟cookie认证的，可以参考这种方法

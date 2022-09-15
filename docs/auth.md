# 认证方法

## Basic

使用用户名和密码直接登录。

如：jenkins

## PrivateToken
 
私人token。

例如rancher，gitlab这种。

## Archery

sql审计平台自定义的认证方法。只适合sql审计

## BKPaaS

蓝鲸API官方文档里面的认证方式。
优点：有API文档
缺点：API支持不丰富。

## BK

模拟实现的cookie登录，所有页面上调用的api都可以调用。如果有需要模拟cookie认证的，可以参考这种方法

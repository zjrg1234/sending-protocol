# 数据中心

将注单,流水数据同步到es中,并提供查询接口

## 配置说明

```
配置文件在resources目录中
dev,test,prod环境分别对应配置文件:
config-dev.yaml
config-test.yaml
config-prod.yaml
```

## 布署说明

1,开发语言:go,需要安装golang1.18或以上版本  
2,部署相关命令

``` shell
run,start,restart命令需要指定--env参数,以激活不同环境对应的配置文件

sh server.sh   #查看相关命令
sh server.sh run --env=prod  #编译并运行,等效于(build->stop->start),推荐使用,env可选参数有dev,test,prod
sh server.sh build  #编译
sh server.sh start --env=prod #启动
sh server.sh stop  #停止
sh server.sh restart  --env=prod #重启,等效于(stop->start)
```

### 正式环境布署

``` shell
#编译并重启服务
sh server.sh run --env=prod
```

### 测试环境布署

``` shell
#编译并重启服务
sh server.sh run --env=test
```

## 支持功能

* 对gin进行了二次封装,参数的绑定,验证,返回值更加简单和规范
* 扩展handler方法中Context,提供上下文traceId,log等常用组件的引用
* error错误处理封装,支持业务错误码和trace信息
* 对gorm进行了泛型封装,数据库操作变的简单易用,规避了原生gorm中常见的坑。
* 自动生成yapi文档,基于ast语法树解析提取注释,注释无须严格规范编写,更加简单易用。
* rpcx微服务支持,融合了rpcx框架,可同时提供基于tcp的远程调用
* migrate支持

### 框架开发说明


### elasticsearch开发说明


### 第三方库

* cast 类型转换
* carbon 时间库
* gin 路由框架
* gorm
* zap log日志库
* json库建议使用jsoniter,性能相比gjson,标准库提升一倍左右

```shell
cast,carbon等库由于使用了大量反射,在高性能场景不建议使用,如批量导入大量数据时
```

### 系统组件升级记录


## 框架开发说明

### 工程化相关

#### 目录说明

* handler 目录相当于controller,因为路由需要提供一个handle方法,根据语义修改.
* vo 有两层含义,视图对象或值对象,我们这里接收或返回视图层(前端)的对象
* dto 数据传输对象,service接受的对象,比如一个接口controller中同时调用了两个service接口,那么可以把vo转为两个dto传给service层,大部分情况下vo和dto相同时略去dto
* consumer 消息队列的消费者订阅的消息出口
* schedule 定时任务
* 其它的应该都懂...

#### 领域驱动模型

相对于动态语言如php,静态语言数据类型是固定的,当提交的参数,写库的参数,和返回的数据结构不一致时,需要考虑定义不同的结构体。项目参考了java中常见相关概念,可以在网上了解一下VO,DTO,BO,PO等概念

### 最佳实践

#### 小技巧
* goland:开发IDE,设置Tag自动补全功能
  https://blog.csdn.net/fly910905/article/details/127146795
* 代码都是红色,拉不下来依赖。设置IDE的GOPROXY=https://goproxy.cn,direct

#### 参数验证

使用了go-playground/validator/v10 验证,如:

```go
type User struct {
    FirstName      string     `validate:"required"`
    LastName       string     `validate:"required"`
    Age            uint8      `validate:"gte=0,lte=130"`
    Email          string     `validate:"required,email"`
    FavouriteColor string     `validate:"iscolor"` // alias for 'hexcolor|rgb|rgba|hsl|hsla'
    Addresses      []*Address `validate:"required,dive,required"` // a person can have a home and cottage...
}
```

### log相关

* 使用zap库
* 如果仅仅是在本地调试建议用logger.Debug()方法。这类log不会写入log文件,但会输出到控制台
* 在handler中使用log时可从ctx中取得。在service中使用log时this.Log,这两种方式会自动记录traceId

```go
func (this *Goods) Save(req vo.Goods) error {
    this.Log.Info("Goods.save", zap.Any("GoodsSaveReq", req))
```

### 错误处理
* 所有错误(包含业务错误和异常错误),返回类型为error
* service中使用this.error开头的相关方法,对于其它接口返回的error,需要用this.error()包装一层。这样可以在trace中追溯错误产生过程

```go
func (this *Goods) ChangeStatus(req vo.GoodsStatusReq) error {
    model, err := this.repo.FindById(req.ID)
    if err != nil {
        //此处对err包装后返回
        return this.error(err) 
    }
    
    if model.ID == 0 {
        //如果不关心错误编号,默认错误编号为5000
        return this.errorMessage("商品不存在")
    }
    
    if req.Status == model.Status {
        //需要错误编号,比如前端需要根据错误编号进行判断,或者提供给第三方调用的接口文档
        return this.errorCodeMessage(4001, "无须修改")
    }
    return nil
}
```
### 接口调试
建议用测试用例调试接口很方便,而且会自动生成文档到yapi,
如tests目录中:
```go
func TestChangeStatus(t *testing.T) {
	req := vo.GoodsStatusReq{
		ID:     1,
		Status: 1,
	}
	resp := tests.Post("/api/goods/change_status", req)
	tests.Echo(resp.Body.String())
}
```
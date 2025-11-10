package system

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"megin/app/schedule"
	meRouter "megin/library/context/router"
	"megin/library/logger"
	"megin/library/migrate"
	redisClient "megin/library/redis"
	"megin/library/rpc"
	"megin/library/validate"
	"megin/router"
	"megin/system/config"
	"megin/system/datasource"
	"megin/system/middleware"
)

// 1,初始化服务
func ServerInit(configPath string, onStart func() error) {
	fmt.Println("configPath:", configPath)
	//1,解析配置文件
	conf := config.InitConfig(configPath)
	//2,Log初始化
	logger.InitLog(logger.LogConfig{LogInConsole: true})
	//3,加载参数验证扩展
	validate.RegisterExtension()
	//4,数据库初始化
	datasource.InitDatabase(conf.Database.Dsn)
	//5,redis初始化
	redisClient.Connect(conf.Redis.Addr, conf.Redis.Password)
	//6,启动时自动执行migrate
	migrate.Install()
	//7,启动RCPX服务
	if conf.Rpcx.Enable {
		rpc.CreateRpcServer(conf.ServiceName, conf.Rpcx)
	}
	//8,业务初始化
	err := onStart()
	if err != nil {
		log.Fatalln(err)
	}
}

// 3,启动服务
func ServerRun() {
	conf := config.GetConfig()
	fmt.Println("Server Run Start")
	logger.Info("Server Run Start")
	//1,Gin框架初始化
	route := router.InitGinRouter()
	go rpc.GetRpcServer().Start()
	//2,启动定时任务
	schedule.Start()
	if route.Run(":"+conf.Port) != nil {
		logger.Error("Server Run Error")
	}
}

// 测试用例的入口
func SetupTestRouter() *gin.Engine {
	//设为release,要不然输出的东西太多，影响视线
	gin.SetMode(gin.ReleaseMode)
	ginRouter := gin.Default()

	ctxRouter := meRouter.NewRouter(ginRouter)
	ctxRouter.Use(middleware.Recover())
	ctxRouter.Use(middleware.Cors())
	ctxRouter.Use(middleware.RequestLog())

	router.GoodsRouter(ctxRouter)
	return ginRouter
}

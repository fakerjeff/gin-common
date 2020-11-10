### 使用说明
+ 在项目根目录下创建 .env文件以模拟环境变量
```shell script
# 以下为 .env 文件示例
DATABASE_USER=root
DATABASE_PASSWORD=juan0521
DATABASE_HOST=127.0.0.1
DATABASE_PORT=3306
DATABASE_NAME=gin_common
DATABASE_LOC=Asia%2FShanghai
GORM_CONTEXT_TIMEOUT=5
DATABASE_MAX_IDLE=10
DATABASE_MAX_LIFETIME=7200
GIN_LOG_LEVEL=debug
GORM_LOG_LEVEL=debug
GORM_SLOW_THRESHOLD=200
LOCALE=zh
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_DB=0
REDIS_MAX_RETRIES=20
REDIS_POOL_SIZE=20
REDIS_MIN_IDLE=5
ACCESS_TOKEN_EXPIRE=7200
SECRET_KEY=ff189145902e4618ada3cdde504175c0
RUN_ENV=dev
```
+ wires包下编写需要注入对象的provider和injector，当不存在wire_gen.go文件时，使用
```
# 使用wire命令生成注入代码 wire_gen.go
cd wires
wire
```
当存在wire_gen.go文件时，使用
```
# 使用go generate更新注入代码
cd wires
go generate
```
provider、injector写法请参考

[https://github.com/google/wire](https://github.com/google/wire "https://github.com/google/wire")

项目运行时需要将injector.go文件排除编译路径（可以使用简单的重命名处理 例如：injector.go.bak, 生成注入代码时，需将其还原）

+ 使用命令 go run server.go migrate 进行数据库迁移
+ orm相关操作参考
[https://gorm.io/](https://gorm.io/ "https://gorm.io/")
+ 其他使用方式，自行查看源码demo


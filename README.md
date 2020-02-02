# dlman
when you struggle to download across the wall, DLMAN stands with you.


# 接口
目前，可以从 handler/*.http 里面看。不过只有请求的参数和格式。

返回的参数格式现在还没有整理。

# 运行
#### 安装
1. 安装 go 1.13+
2. 安装 redis
3. 安装 gcc（sqlite3驱动有c的拓展），win比较麻烦，mac可能会好一些。

#### 编译
程序的编译运行，需要设置几个环境变量，程序运行所需要的配置项，目前是通过环境变量的方式传递的。
1. DLMAN_ENV， dev或者prod
2. DLMAN_LOCAL_VENDOR_PATH， 存放下载文件的临时路径，绝对路径
3. DLMAN_REDIS_ADDR， redis服务的地址，格式为 host:port
4. DLMAN_SQLITE3_PATH，sqlite3数据库文件的绝对路径，如/home/someone/dlman/database.db

编译：

```shell script
cd /path/to/dlman
go build

cd /path/to/dlman/worker
go build
```
会生成`/path/to/dlman/dlman`和`/path/to/dlman/worker/worker`两个可执行文件

#### 运行
```shell script
/path/to/dlman/dlman
/path/to/dlman/worker/worker
```

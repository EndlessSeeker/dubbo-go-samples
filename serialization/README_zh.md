# Dubbo-go Serialization

## 1.介绍

本示例演示如何在Dubbo-go框架中使用不同的序列化方式发起请求

## 2.序列化类型

dubbo-go支持4中序列化方式，包括hessian2、protobuf、msgpack和json，默认使用protobuf进行序列化和反序列化，如果需要使用其他序列化方式，需要进行额外的配置和声明

```go
const (
	Hessian2Serialization = "hessian2"
	ProtobufSerialization = "protobuf"
	MsgpackSerialization  = "msgpack"
	JSONSerialization     = "json"
)
```

## 3. Json

### 3.1 使用说明

如果希望使用Json的序列化方式，可以在客户端中针对client或者service来进行指定

指定client序列化方式：
```go
	cli, err := client.NewClient(
        client.WithClientSerializationJSON(),
		// Other Client Options
    )
```

指定service序列化方式:
```go
    svc, err = greet.NewGreetService(cli, client.WithSerializationJSON())
)
```
服务端无需额外指定序列化方式,会判断请求的序列化类型，使用对应的方式进行反序列化

### 3.2 案例结果

关于Json序列化的示例在serialization/json文件夹下,先在go-server/cmd中启动服务端,再在go-client/cmd中启动客户端,可以观察到客户端发起了两次调用,从客户端打印的日志可以观察到serialization=json，且返回了正确的结果

```
2024-03-07 14:22:36     INFO    logger/logging.go:42    URL specified explicitly [127.0.0.1:20000]
2024-03-07 14:22:36     INFO    logger/logging.go:42    [TRIPLE Protocol] Refer service: [tri://127.0.0.1:20000/greet.GreetService?app.version=&application=dubbo.io&async=false&bean.name=greet.GreetService&cluster=failover&config.tracing=&environment=&generic=&group=&interface=greet.GreetService&loadbalance=&metadata-type=local&module=sample&name=dubbo.io&organization=dubbo-go&owner=dubbo-go&peer=true&provided-by=&reference.filter=cshutdown&registry.role=0&release=dubbo-golang-3.2.0&remote.timestamp=&retries=&serialization=json&side=consumer&sticky=false&timestamp=1709792556&version=]
2024-03-07 14:22:36     INFO    logger/logging.go:42    Greet response: [hello world 1]
2024-03-07 14:22:36     INFO    logger/logging.go:42    URL specified explicitly [127.0.0.1:20000]
2024-03-07 14:22:36     INFO    logger/logging.go:42    [TRIPLE Protocol] Refer service: [tri://127.0.0.1:20000/greet.GreetService?app.version=&application=dubbo.io&async=false&bean.name=greet.GreetService&cluster=failover&config.tracing=&environment=&generic=&group=&interface=greet.GreetService&loadbalance=&metadata-type=local&module=sample&name=dubbo.io&organization=dubbo-go&owner=dubbo-go&peer=true&provided-by=&reference.filter=cshutdown&registry.role=0&release=dubbo-golang-3.2.0&remote.timestamp=&retries=&serialization=json&side=consumer&sticky=false&timestamp=1709792556&version=]
2024-03-07 14:22:36     INFO    logger/logging.go:42    Greet response: [hello world 2]

```

## 4. hessian2

### 4.1 使用说明

目前, hessian2序列化方式没办法通过IDL直接生成调用代码,需要手动实现。如果希望使用hessian2的序列化方式进行调用, 采用配置文件的方式可能实现更加容易

#### 4.1.1 客户端

编写Service，包含一个GetGreet方法
```go
package main

import "context"

type Greet struct {
	Name string
}

func (g *Greet) JavaClassName() string {
	return "com.apache.dubbo.sample.basic.Greet"
}

type GreetProvider struct {
	GetGreet func(ctx context.Context, req *Greet) (*Greet, error)
}
```

在conf/dubbogo.yaml中进行配置,声明调用的Service
```yaml
dubbo:
  consumer:
    references:
      GreetProvider:
        protocol: tri
        serialization: hessian2
        url: tri://127.0.0.1:20000
        interface: com.apache.dubbo.sample.basic.Greet
```

在main函数中进行初始化，并且发起调用
```go
package main

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/config"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	hessian "github.com/apache/dubbo-go-hessian2"
)

var greetProvider = new(GreetProvider)

func init() {
	config.SetConsumerService(greetProvider)
	hessian.RegisterPOJO(&Greet{})
}

func main() {
	err := config.Load()
	if err != nil {
		panic(err)
	}

	greet, err := greetProvider.GetGreet(context.TODO(), &Greet{Name: "hello world"})
	if err != nil {
		panic(err)
	}
	println(greet.Name)
}
```

#### 4.1.2 服务端

编写Service，包含一个GetGreet方法
```go
package main

import (
	"context"
)

import (
	gxlog "github.com/dubbogo/gost/log"
)

type Greet struct {
	Name string
}

func (g *Greet) JavaClassName() string {
	return "com.apache.dubbo.sample.basic.Greet"
}

type GreetProvider struct {
}

func (u *GreetProvider) GetGreet(ctx context.Context, greet *Greet) (*Greet, error) {
	gxlog.CInfo("req:%#v", greet)
	rsp := Greet{"Hello, " + greet.Name}
	gxlog.CInfo("rsp:%#v", rsp)
	return &rsp, nil
}
```

在conf/dubbogo.yaml中进行配置,声明Service
```yaml
dubbo:
  protocols:
    triple:
      name: tri
      port: 20000
  provider:
    services:
      GreetProvider:
        serialization: hessian2
        interface: com.apache.dubbo.sample.basic.Greet
```

在main函数中进行初始化，并且启动服务
```go
package main

import (
	"dubbo.apache.org/dubbo-go/v3/config"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	"fmt"
	hessian "github.com/apache/dubbo-go-hessian2"
	"github.com/dubbogo/gost/log/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	survivalTimeout = int(3 * time.Second)
)

func init() {
	//------for hessian2------
	hessian.RegisterPOJO(&Greet{})
	config.SetProviderService(new(GreetProvider))
}

func main() {
	if err := config.Load(); err != nil {
		panic(err)
	}
	initSignal()
}

func initSignal() {
	signals := make(chan os.Signal, 1)
	// It is not possible to block SIGKILL or syscall.SIGSTOP
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)
	for {
		sig := <-signals
		logger.Infof("get signal %s", sig.String())
		switch sig {
		case syscall.SIGHUP:
			// reload()
		default:
			time.Sleep(time.Second * 5)
			time.AfterFunc(time.Duration(survivalTimeout), func() {
				logger.Warnf("app exit now by force...")
				os.Exit(1)
			})

			// The program exits normally or timeout forcibly exits.
			fmt.Println("provider app exit now...")
			return
		}
	}
}
```

### 4.2 案例结果

关于Hessian2序列化的示例在serialization/hessian2文件夹下,先在go-server/cmd中启动服务端,再在go-client/cmd中启动客户端,从客户端打印的日志可以观察到serialization=hessian2，且返回了正确的结果

```
2024-03-07 17:56:27     INFO    logger/logging.go:42    The following profiles are active: [default]
2024-03-07 17:56:27     INFO    config/root_config.go:138       [Config Center] Config center doesn't start
2024-03-07 17:56:27     INFO    config/reference_config.go:142  URL specified explicitly tri://127.0.0.1:20000
2024-03-07 17:56:27     INFO    triple/triple.go:124    [TRIPLE Protocol] Refer service: tri://127.0.0.1:20000/com.apache.dubbo.sample.basic.Greet?app.version=&application=dubbo.io&async=false&bean.name=GreetProvider&cluster=failover&config.tracing=&environment=&generic=&group=&interface=com.apache.dubbo.sample.basic.Greet&loadbalance=&metadata-type=local&module=sample&name=dubbo.io&organization=dubbo-go&owner=dubbo-go&peer=true&provided-by=&reference.filter=cshutdown&registry.role=0&release=dubbo-golang-3.2.0&remote.timestamp=&retries=&serialization=hessian2&side=consumer&sticky=false&timestamp=1709805387&version=
Hello, hello world

```


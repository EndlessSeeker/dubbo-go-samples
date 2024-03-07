# Dubbo-go Serialization

## 1.Introduction

This example demonstrates how to communicate using different serializations in the Dubbo-go framework

## 2.Serialization type

Four serialization types are supported in the Dubbo-go, including hessian2, protobuf, msgpack and json. By default, protobuf is used for serialization and deserialization. If you need to use other serialization type, additional configuration and declaration are required.

```go
const (
	Hessian2Serialization = "hessian2"
	ProtobufSerialization = "protobuf"
	MsgpackSerialization  = "msgpack"
	JSONSerialization     = "json"
)
```

## 3. Json

### 3.1 Introduction

If you want to use Json serialization, you can specify it in the client for client or service scope.

client scope:
```go
	cli, err := client.NewClient(
        client.WithClientSerializationJSON(),
		// Other Client Options
    )
```

service scope:
```go
    svc, err = greet.NewGreetService(cli, client.WithSerializationJSON())
)
```
The server does not need to specify a serialization type. It will identify the serialization type of the request and deserialize it.

### 3.2 Sample Result

An example of Json serialization is in the serialization/json folder. First start the server in go-server/cmd, and then start the client in go-client/cmd. You can observe that the client made two requests. 'serialization=json' can be observed in the log printed by the client, and the correct result is returned.

```
2024-03-07 14:22:36     INFO    logger/logging.go:42    URL specified explicitly [127.0.0.1:20000]
2024-03-07 14:22:36     INFO    logger/logging.go:42    [TRIPLE Protocol] Refer service: [tri://127.0.0.1:20000/greet.GreetService?app.version=&application=dubbo.io&async=false&bean.name=greet.GreetService&cluster=failover&config.tracing=&environment=&generic=&group=&interface=greet.GreetService&loadbalance=&metadata-type=local&module=sample&name=dubbo.io&organization=dubbo-go&owner=dubbo-go&peer=true&provided-by=&reference.filter=cshutdown&registry.role=0&release=dubbo-golang-3.2.0&remote.timestamp=&retries=&serialization=json&side=consumer&sticky=false&timestamp=1709792556&version=]
2024-03-07 14:22:36     INFO    logger/logging.go:42    Greet response: [hello world 1]
2024-03-07 14:22:36     INFO    logger/logging.go:42    URL specified explicitly [127.0.0.1:20000]
2024-03-07 14:22:36     INFO    logger/logging.go:42    [TRIPLE Protocol] Refer service: [tri://127.0.0.1:20000/greet.GreetService?app.version=&application=dubbo.io&async=false&bean.name=greet.GreetService&cluster=failover&config.tracing=&environment=&generic=&group=&interface=greet.GreetService&loadbalance=&metadata-type=local&module=sample&name=dubbo.io&organization=dubbo-go&owner=dubbo-go&peer=true&provided-by=&reference.filter=cshutdown&registry.role=0&release=dubbo-golang-3.2.0&remote.timestamp=&retries=&serialization=json&side=consumer&sticky=false&timestamp=1709792556&version=]
2024-03-07 14:22:36     INFO    logger/logging.go:42    Greet response: [hello world 2]

```

## 4. Hessian2

### 4.1 Introduction

Currently, we cannot directly generate the calling code through IDL with the hessian2 serialization type and needs to be implemented manually. If you want to use hessian2 to write your code, using the configuration file may be easier.

#### 4.1.1 Client

Write a Service, including a GetGreet method
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

Declare the aimed service in conf/dubbogo.yaml
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

Initialize and make a call
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

#### 4.1.2 Server

Write a Service, including a GetGreet method
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

Declare Service in conf/dubbogo.yaml
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

Initialize in the main function and start the service
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

### 4.2 Sample Result

An example of Hessian2 serialization is in the serialization/hessian2 folder. First start the server in go-server/cmd, and then start the client in go-client/cmd. 'serialization=json' can be observed in the log printed by the client, and the correct result is returned.

```
2024-03-07 17:56:27     INFO    logger/logging.go:42    The following profiles are active: [default]
2024-03-07 17:56:27     INFO    config/root_config.go:138       [Config Center] Config center doesn't start
2024-03-07 17:56:27     INFO    config/reference_config.go:142  URL specified explicitly tri://127.0.0.1:20000
2024-03-07 17:56:27     INFO    triple/triple.go:124    [TRIPLE Protocol] Refer service: tri://127.0.0.1:20000/com.apache.dubbo.sample.basic.Greet?app.version=&application=dubbo.io&async=false&bean.name=GreetProvider&cluster=failover&config.tracing=&environment=&generic=&group=&interface=com.apache.dubbo.sample.basic.Greet&loadbalance=&metadata-type=local&module=sample&name=dubbo.io&organization=dubbo-go&owner=dubbo-go&peer=true&provided-by=&reference.filter=cshutdown&registry.role=0&release=dubbo-golang-3.2.0&remote.timestamp=&retries=&serialization=hessian2&side=consumer&sticky=false&timestamp=1709805387&version=
Hello, hello world
```


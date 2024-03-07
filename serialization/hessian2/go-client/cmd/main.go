/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
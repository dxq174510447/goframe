## mvc(springboot),aop(filter-chain),orm(mybatis),log(logback)

## 案例见 仓库下的 firstgo

## 不支持(可能是永久不支持也可能是暂时不支持)
1. 配置文件注入到struct不支持混合类型（支持内嵌）
2. sql查询当参数类型是指针struct的时候，不支持混合（支持内嵌）

## 前提
1. struct-->类   
2. 所有的controller,service,dao都是单例     
3. 所有的aop都是基于方法中调用类字段函数，类在加载到容器中会解析字段函数，在其外层在包裹一层在重新赋值给该字段（后面如果有更好的实现会替换掉，毕竟凭空多一个字段函数还是有点别扭）   
4. 在类调用中最好不要直接使用字段，方法和字段命令规则，字段在方法的后面加上"_"即可，例如方法   
5. 所有方法第一个参数必须是*context.LocalStack，类似java中的threadlocal类型  
   
### mvc(springboot)

#### 支持yaml配置文件
1. 支持将配置文件里面的值直接注入到实例中（支持map，指针struct，和基础类型 类似spring value注解的用法）
见demo configyml.go，localconfigyml.go 类似spring的配置，可以注入到容器实例中，或者从环境中获取

#### 支持http filter
见demo binduserfilter.go 实现http.Filter即可

#### 控制器支持类似spring注解  
见demo userscontroller.go   
1. 直接将控制器的方法映射到http的handle函数中，可以设置控制器方法对应的http路径和http method   
2. 直接将request中的参数映射到控制器方法中的参数，支持requestbody，query，form，head，path。(暂时不支持上传)  
4. 控制器方法可直接返回指针类，根据视图模型渲染返回的结果。默认是json序列化   
5. 后续支持根据控制器生成swagger文档   

### orm(mybatis)
见demo dao目录下的usersdao.go，usersxml.go  
1. 只需要把方法写出来即可，不需要去实现，默认会找到对应的sql并执行(类似mybatis mapper接口的用法)  
2. 将dao和sql分开管理，会有专门的go文件里面就是sql的xml内容     
3. sql支持golang模版用法，循环判断等   
4. 支持查询参数是结构体指针，基础类型，string等   
5. 根据threadlocal中的变量，失败是返回error还是panic。默认panic  
6. 内嵌数据库拦截器。daoconnectproxyfilter.go，txreadonlyproxyfilter.go，txrequirenewproxyfilter.go，txrequireproxyfilter.go。对应的是数据库连接，只读事物，新事物和线程事物拦截  
7. 包含BaseDao可以混合到其它dao中  

使用建议
1. 建议返回返回两个结果（测试案例都是返回两个结果），error放在最后面       
2. struct结构都是返回指针类型，无论是slice还是单个  
3. 熟悉mybatis的，这个只支持include标签  


### aop
需要实现ProxyTarger接口,具体见demo txrequireproxyfilter.go  


### log
支持logback配置，暂时只支持console appender,具体用法见案例。后面会添加file和rollingfile

### 在持续优化中

### 20210520-20210528 v1.1.1(完成)

### 20210531-20210611 v1.1.2(部分完成)
1. 引入单例注入(完成)
2. 引入配置值注入(完成)
3. 引入事件监听模型
4. 框架引入logger(完成)


### 20210613-20210625 v1.1.3
1. yaml项目变更，触发注入的配置实例变更
2. 引入事件监听模型(完成)
3. 优化logger，引入file，rollingfile appender
4. 修复database filter，针对返回error触发回滚（之前是panic）
5. 将http渲染组件化

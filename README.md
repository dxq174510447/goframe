## mvc(springboot),aop(filter-chain),orm(mybatis)

## 案例见 firstgo

## 前提
1. struct-->类   
2. 所有的controller,service,dao都是单例    
3. 所有的aop都是基于类字段类型是func来动态包裹一层来实现的，所以调用类的方法，在通过类的方法中去调用类的字段，然后字段在初始化的时候去设置它的实现    
4. 在类调用中最好不要直接使用字段，方法和字段命令规则，字段在方法的后面加上"_"即可，例如方法  
```
   type UsersController struct {
       Save_         func(local *context.LocalStack, param *vo.UsersAdd, self *UsersController) *vo.UsersVo
   }
   func (c *UsersController) Save(local *context.LocalStack, param *vo.UsersAdd) *vo2.JsonResult {
       result := c.Save_(local, param, c)
       return util.JsonUtil.BuildJsonSuccess(result)
   }
```
   
5. 所有方法第一个参数必须是*context.LocalStack，类似java中的threadlocal类型
   
### mvc(springboot)
见demo  
```golang
//.....

func (c *UsersController) ChangeStatus(
    local *context.LocalStack,
    id int, status int,
    requestId int,
    yid int, ystatus int) *vo2.JsonResult {
    fmt.Println(id, status, requestId, yid, ystatus)
    c.ChangeStatus_(local, id, status, c)
    return util.JsonUtil.BuildJsonSuccess(nil)
}
// 对应的方法注解

{
    Name: "ChangeStatus",
        Annotations: []*proxy.AnnotationClass{
        http.NewRestAnnotation(
        	"/change/{yid}/status/{ystatus}",  //请求url
        	"post",  //请求method
            "_,id,status,_,_,_", //方法中query(form)中的参数位置和request中的key
            "_,_,_,_,yid,ystatus", //方法中路径参数位置和对应的key
            "_,_,_,requestId,_,_", //方法中header中的参数位置和header中的key
            ""), // 返回视图模式 默认json
    },
},

```

### orm(mybatis)
1. 建议返回返回两个结果（测试案例都是返回两个结果），error放在最后面    
2. struct结构都是返回指针类型，无论是slice还是单个
3. 原有mybatis 只支持include标签
4. sql语句支持golang模版
5. 如果有多个参数，必须要指定每个参数的别名，在sql中使用别名

见案例 

```
// userdao.go
type UsersDao struct {
    FindByNameAndStatus_ func(local *context.LocalStack, name string, status int, statusList []string) ([]*po.Users, error)
}
func (c *UsersDao) FindByNameAndStatus(local *context.LocalStack, name string, status int, statusList []string) ([]*po.Users, error) {
	return c.FindByNameAndStatus_(local, name, status, statusList)
}

// userdaoxml.go 部分代码
<select id="FindByNameAndStatus">
			select * from users where   
			<include refid="conditionA1"></include>
           order by id desc 
	</select>

	<sql id="conditionA1">
		name = #{name} and status = #{status} 
		{{if .statusList}}
			and status in (
				{{range $index, $ele := $.statusList}}{{if $index}},{{end}}#{statusList[{{$index}}]}{{end}}
			)
		{{end}}
	</sql>

	<select id="FindIds">
			select id from users order by id desc 
	</select>

	<select id="FindNames">
			select distinct name from users 
			where 1=1
<![CDATA[
			{{if .NameIn}}
				and name in (
					{{range $index, $ele := $.NameIn}}{{if $index}},{{end}}#{NameIn[{{$index}}]}{{end}}
				)
			{{end}}
			and status < 1
]]>
			order by id desc 
	</select>
```


### aop
需要实现ProxyTarger接口,具体见demo

```

type ProxyTarger interface {
	ProxyTarget() *ProxyClass  //ProxyClass 类似java的注解，
}


```

### http filter
见 BindUserFilter


### controller,service,dao aop
见 TxRequireNewProxyFilter

### 在持续优化中

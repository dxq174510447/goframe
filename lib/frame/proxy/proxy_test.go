package proxy

import (
	"firstgo/frame/context"
	"firstgo/povo/po"
	"fmt"
	"testing"
)

type RoleDao struct {
	Proxy   *ProxyClass
	Save_   func(context *context.LocalStack, users *po.Users, self *RoleDao) int
	Update_ func(context *context.LocalStack, users *po.Users, self *RoleDao) int
}

func (r *RoleDao) Save(context *context.LocalStack, users *po.Users) int {
	return r.Save_(context, users, r)
}

func (r *RoleDao) Update(context *context.LocalStack, users *po.Users) int {
	return r.Update_(context, users, r)
}

func (r *RoleDao) ProxyTarget() *ProxyClass {
	return r.Proxy
}

var roleDao RoleDao = RoleDao{
	Save_: func(context *context.LocalStack, users *po.Users, self *RoleDao) int {
		fmt.Println(" save begin ", users)
		defer fmt.Println(" save end ", users)
		return users.Id
	},
	Update_: func(context *context.LocalStack, users *po.Users, self *RoleDao) int {
		fmt.Println(" update begin ", users)
		defer fmt.Println(" update end ", users)
		return 2
	},
	Proxy: &ProxyClass{
		Annotations: []*AnnotationClass{
			&AnnotationClass{
				Name: AnnotationDao,
			},
		},
	},
}

func TestName(t *testing.T) {
	context := context.NewLocalStack()
	AddClassProxy(ProxyTarger(&roleDao))

	fmt.Println(context.Get("a"))
	fmt.Println(roleDao.Save(context, &po.Users{Name: "haha1", Id: 1}))

	fmt.Println(context.Get("a"))
	fmt.Println(roleDao.Save(context, &po.Users{Name: "haha2", Id: 2}))

	fmt.Println(context.Get("a"))
	//fmt.Println(roleDao.Save(&po.Users{Name: "haha3"}))
	//fmt.Println(roleDao.Save(&po.Users{Name: "haha4"}))
	//fmt.Println(roleDao.Update(&po.Users{Name: "haha5"}))
	//fmt.Println(roleDao.Update(&po.Users{Name: "haha6"}))
	fmt.Println(roleDao.Update(context, &po.Users{Name: "haha7"}))
	//fmt.Println(roleDao.Update(context,&po.Users{Name: "haha8"}))

	//fmt.Println(roleDao.Update(nil))
	//fmt.Println(roleDao.Save(nil))

}

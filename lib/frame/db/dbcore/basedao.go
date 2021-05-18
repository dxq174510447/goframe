package dbcore

import (
	"firstgo/frame/context"
	"firstgo/frame/proxy"
	"reflect"
)

var BaseProxyTarger *proxy.ProxyClass = &proxy.ProxyClass{}

type BaseDao struct {
	Save_   func(local *context.LocalStack, entity interface{}) (int, error)
	Update_ func(local *context.LocalStack, entity interface{}) (int, error)
	Delete_ func(local *context.LocalStack, id interface{}) (int, error)
	Get_    func(local *context.LocalStack, id interface{}) (interface{}, error)
	Find_   func(local *context.LocalStack, entity interface{}) ([]interface{}, error)
	//Update_ func(entity interface{}) (int,error)
	//Delete_ func(id interface{}) (int,error)
	//Get_    func(id interface{}) (interface{},error)
	//Find_ func(target interface{}) ([]interface{},error)
	//FindIds_ func(param interface{}) ([]interface{},error)
	//FindByIds_ func(param interface{}) ([]interface{},error)
}

// Save entity类型指针
func (b *BaseDao) Save(local *context.LocalStack, entity interface{}) (int, error) {
	return b.Save_(local, entity)
}

func (b *BaseDao) Update(local *context.LocalStack, entity interface{}) (int, error) {
	return b.Update_(local, entity)
}

func (b *BaseDao) Delete(local *context.LocalStack, id interface{}) (int, error) {
	return b.Delete_(local, id)
}
func (b *BaseDao) Get(local *context.LocalStack, id interface{}) (interface{}, error) {
	return b.Get_(local, id)
}
func (b *BaseDao) Find(local *context.LocalStack, entity interface{}) ([]interface{}, error) {
	return b.Find_(local, entity)
}

func (b *BaseDao) ProxyTarget() *proxy.ProxyClass {
	return BaseProxyTarger
}

// Update entity类型数指针
//func (b *BaseDao) Update(entity interface{}) (int,error){
//	return b.Update_(entity)
//}
//
//// Delete ID主键
//func (b *BaseDao) Delete(id interface{}) (int,error){
//	return b.Delete_(id)
//}
//
//// Get ID主键 返回entity类型指针
//func (b *BaseDao) Get(id interface{}) (interface{},error){
//	return b.Get_(id)
//}
//
//// Find entity类型指针参数 返回entity指针的slice
//func (b *BaseDao) Find(param interface{}) ([]interface{},error){
//	return b.Find_(param)
//}
//
//// FindIds entity类型指针参数 返回id的slice
//func (b *BaseDao) FindIds(param interface{}) ([]interface{},error){
//	return b.FindIds_(param)
//}
//
//// FindByIds id的slice类型 返回entity指针的slice
//func (b *BaseDao) FindByIds(param interface{}) ([]interface{},error){
//	return b.FindByIds_(param)
//}

var BaseDaoType reflect.Type = reflect.TypeOf(BaseDao{})

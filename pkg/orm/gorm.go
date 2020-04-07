package orm

import "github.com/jinzhu/gorm"

//SQLExecutor sql执行器
type SQLExecutor interface {
	NewRecord(value interface{}) bool
	Create(value interface{}) *gorm.DB
	Save(value interface{}) *gorm.DB
	Delete(value interface{}, where ...interface{}) *gorm.DB

	Find(out interface{}, where ...interface{}) *gorm.DB
	Exec(sql string, values ...interface{}) *gorm.DB
	First(out interface{}, where ...interface{}) *gorm.DB
}

//DBInfo db链接信息
type DBInfo struct {
	*gorm.DB
}

//Begin 开启事物
func (dbInfo *DBInfo) Begin(trans *Transaction, objects ...TransactionObject) (*Transaction, error) {
	if trans == nil {
		_trans := dbInfo.DB.Begin()
		trans = &Transaction{DB: _trans, objects: []TransactionObject{}}
	}
	for _, o := range objects {
		trans.AddObject(o)
	}
	return trans, nil
}

//TransactionObject interface
type TransactionObject interface {
	SetTransaction(trans *Transaction)
}

//Transaction 事物维护的object
type Transaction struct {
	*gorm.DB
	objects []TransactionObject
}

//AddObject 往事物维护的object中增加object
func (trans *Transaction) AddObject(o TransactionObject) {
	trans.objects = append(trans.objects, o)
	o.SetTransaction(trans)
}

//Commit 事物提交
func (trans *Transaction) Commit() error {
	err := trans.DB.Commit().Error
	for _, o := range trans.objects {
		o.SetTransaction(nil)
	}
	return err
}

//Rollback 事物回滚
func (trans *Transaction) Rollback() error {
	err := trans.DB.Rollback().Error
	for _, o := range trans.objects {
		o.SetTransaction(nil)
	}
	return err
}

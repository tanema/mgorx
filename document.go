package mgorx

import (
  "github.com/robfig/revel"
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
  "errors"
  "reflect"
)

type Document struct {
  D           interface{}
  changes     interface{}
  LastError   error
}

func (doc *Document) Id() string {
  return reflect.ValueOf(doc.D).Elem().FieldByName("Id").String()
}

func (doc *Document) IsNew() bool {
  return !doc.IsPersisted()
}

func (doc *Document) IsPersisted() bool {
  return bson.ObjectId(doc.Id()).Valid()
}

func (doc *Document) Validate(v *revel.Validation) {
  reflect.ValueOf(doc.D).MethodByName("Validate").Call([]reflect.Value{reflect.ValueOf(v)})
}

func (doc *Document) Get(field_name string) (val interface {}) {
  return reflect.ValueOf(doc.D).Elem().FieldByName(field_name).Interface()
}

func (doc *Document) Set(field_name string, v interface{}) {
  reflect.ValueOf(doc.D).Elem().FieldByName(field_name).Set(reflect.Value(reflect.ValueOf(v)))
}

func (doc *Document) Save(v *revel.Validation) bool {
  return doc.saveChain(v)
}

func (doc *Document) Update(changes interface{}, v *revel.Validation) bool {
  doc.changes = changes
  return doc.saveChain(v)
}

func (doc *Document) Delete() bool {
  doc.BeforeDestroy()
  collection_name := collection_name_from(doc.D)
  err := with_collection(collection_name, func(c *mgo.Collection) (err error) {
    if doc.IsPersisted() {
      err = c.RemoveId(doc.Id())
    }
    doc.LastError = err
    return
  })
  doc.AfterDestroy()
  return err == nil
}

func (doc *Document) saveChain(v *revel.Validation) bool {
  if v != nil {
    doc.BeforeValidation()
    doc.Validate(v)
    if v.HasErrors() {
      return false
    }
    doc.AfterValidation()
  }
  doc.BeforeSave()
  if doc.IsNew() {
    doc.BeforeCreate()
  }else{
    doc.BeforeUpdate()
  }
  saved := doc.save()
  if doc.IsNew() {
    doc.AfterCreate()
  }else{
    doc.AfterUpdate()
  }
  doc.AfterSave()

  return saved
}

func (doc *Document) save() bool {
  collection_name := collection_name_from(doc.D)
  err :=  with_collection(collection_name, func(c *mgo.Collection) (err error) {
    if doc.IsPersisted() {
      if doc.changes != nil {
        err = c.UpdateId(doc.Id(), doc.changes)
      }else{
        err = c.UpdateId(doc.Id(), doc.D)
      }
    }else{
      err = c.Insert(doc.D)
    }
    doc.LastError = err
    return
  })
  return err == nil
}

//callbacks
func (doc *Document) BeforeValidation() bool {}
func (doc *Document) AfterValidation() bool {}
func (doc *Document) BeforeSave() bool {}
func (doc *Document) BeforeCreate() bool {}
func (doc *Document) BeforeUpdate() bool {}
func (doc *Document) AfterUpdate() bool {}
func (doc *Document) AfterCreate() bool {}
func (doc *Document) AfterSave() bool {}
func (doc *Document) BeforeDestroy() bool {}
func (doc *Document) AfterDestroy() bool {}

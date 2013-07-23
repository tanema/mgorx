package mgorx

import (
  "github.com/robfig/revel"
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
  "reflect"
  "strings"
)

type Collection struct {
  collection_name string
  init_obj        interface{}
  obj_type        reflect.Type
}

func GetCollection(c interface{}) *Collection {
  return &Collection{collection_name_from(c), c, reflect.TypeOf(c)}
}

func collection_name_from(result interface{}) string {
  type_of := reflect.TypeOf(result).String()
  name := strings.ToLower(type_of[strings.LastIndex(type_of,".")+1:]) + "s"
  return name
}

func with_collection(collection_name string, next func(*mgo.Collection) error) error {
  session := getSession()
  defer session.Close()
  col := session.DB(database_name).C(collection_name)
  return next(col)
}

func (c *Collection) New(new_obj interface{}) Document {
  doc := Document{D: new_obj}
  doc.Set("Document", doc)
  return doc
}

func (c *Collection) Create(new_obj interface{}, v *revel.Validation) bool {
  doc := c.New(new_obj)
  return doc.saveChain(v)
}

func (col *Collection) Find(result, q interface{}) error {
  err := with_collection(col.collection_name, func(c *mgo.Collection) error {
    // if the id is given just query by the id
    if query_type := reflect.TypeOf(q).Kind().String(); query_type == "string" {
      return c.Find(bson.M{"_id": bson.ObjectIdHex(reflect.ValueOf(q).String())}).One(result)
    }else{ // find one with the query
      return c.Find(q).One(result)
    }
  })
  col.New(result)
  return err
}

func (col *Collection) Where(results, q interface{}, options map[string]int) error {
  err := with_collection(col.collection_name, func(c *mgo.Collection) error {
    fn := c.Find(q)
    if skip, ok := options["skip"]; ok {
      fn = fn.Skip(skip)
    }
    if limit, ok := options["limit"]; ok {
      fn = fn.Limit(limit)
    }
    return fn.All(results)
  })

  slicev := reflect.ValueOf(results).Elem()
  for i := 0; i < slicev.Len(); i++ {
    v := slicev.Index(i).Addr().Interface()
    col.New(v)
  }
  return err
}

func (c *Collection) All(result interface{}, options map[string]int) error {
  return c.Where(result, nil, options)
}

func (col *Collection) Delete(q interface{}) error {
  return with_collection(col.collection_name, func(c *mgo.Collection) error {
    return c.Remove(q)
  })
}

func (col *Collection) DeleteAll(q interface{}) error {
  return with_collection(col.collection_name, func(c *mgo.Collection) error {
    _, err := c.RemoveAll(q)
    return err
  })
}

package kl2cache_test

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func init() {
	//cache.Init(context.Background())

	//cache.DefaultProvider
}

//func TestReflectReturnType(t *testing.T) {
//	u := &User{}
//	err := Get("x", u, func(key string) (val interface{}, err error) {
//		val = &User{
//			ID:   "xxx",
//			Name: "dada",
//		}
//		return
//	})
//	if err != nil {
//		panic(err)
//	}
//
//	err = Get("x", u, func(key string) (val interface{}, err error) {
//		val = &User{
//			ID:   "xxx",
//			Name: "dada",
//		}
//		return
//	})
//	if err != nil {
//		panic(err)
//	}
//}
//
//var mapCache = map[string]interface{}{}
//
//func Get(key string, val interface{}, fn func(key string) (val interface{}, err error)) (err error) {
//	if v, ok := mapCache[key]; ok {
//		val = v
//		return
//	}
//	val, err = fn(key)
//	if err != nil {
//		return
//	}
//	mapCache[key] = val
//	return
//}

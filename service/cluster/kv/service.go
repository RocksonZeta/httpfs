package kv

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/gomodule/redigo/redis"
)

// Service the Redis service, contains the config and the redis Pool
type Service struct {
	// // Connected is true when the Service has already connected
	// Connected bool
	// // Config the redis config for this redis
	Redis redis.Conn
}

func NewService(cli redis.Conn) *Service {
	return &Service{Redis: cli}
}

func (r *Service) Close() error {
	return r.Redis.Close()
}

// func NewService(conf *config.RedisConfig) *Service {
// 	// // }
// 	// r := &Service{Config: conf}
// 	// r.connect()
// 	// return r
// }

// PingPong sends a ping and receives a pong, if no pong received then returns false and filled error
func (r *Service) PingPong() (bool, error) {
	// c := r.Redis.Get()
	// defer c.Close()
	msg, err := r.Redis.Do("PING")
	if err != nil || msg == nil {
		return false, err
	}
	return (msg == "PONG"), nil
}

func (r *Service) SetJson(key string, value interface{}, secondsLifetime int64) (err error) {
	bs, err := r.marshal(value)
	if err != nil {
		return err
	}
	if secondsLifetime > 0 {
		_, err = r.Redis.Do("SETEX", key, secondsLifetime, bs)
	} else {
		_, err = r.Redis.Do("SET", key, bs)
	}

	return
}
func (r *Service) GetJson(key string, out interface{}) error {
	redisVal, err := r.Redis.Do("GET", key)
	if err != nil {
		return err
	}
	if redisVal == nil {
		return nil
	}
	return r.unmarshal(redisVal, out)
}
func (r *Service) GetMap(key string) (map[string]interface{}, error) {

	redisVal, err := r.Redis.Do("GET", key)
	if err != nil {
		return nil, err
	}
	if redisVal == nil {
		// return nil, service.ErrKeyNotFound.Format(key)
		return nil, nil
	}
	out := make(map[string]interface{})
	err = r.unmarshal(redisVal, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns value, err by its key
//returns nil and a filled error if something bad happened.
func (r *Service) Set(key interface{}, value interface{}, seconds int) error {
	if seconds <= 0 {
		_, err := r.Redis.Do("SET", key, value)
		return err
	} else {
		_, err := r.Redis.Do("SETEX", key, seconds, value)
		return err
	}

}
func (r *Service) Get(key interface{}) (interface{}, error) {
	return r.Redis.Do("GET", key)
	// redisVal, err := r.Redis.Do("GET", key)

	// if err != nil {
	// 	return nil, err
	// }
	// if redisVal == nil {
	// 	// return nil, service.ErrKeyNotFound.Format(key)
	// 	return nil, nil
	// }
	// return redisVal
	// return r.unmarshal(redisVal)
}

func (r *Service) HSet(key, field string, value interface{}, secondsLifetime int64) (err error) {
	bs, err := r.marshal(value)
	if err != nil {
		return err
	}
	_, err = r.Redis.Do("HSET", key, field, bs)
	if nil != err {
		return err
	}
	// if has expiration, then use the "EX" to delete the key automatically.
	if secondsLifetime > 0 {
		err = r.Expire(key, secondsLifetime)
	}
	return err
}
func (r *Service) HGet(key, field string, result interface{}) error {
	bs, err := r.Redis.Do("HGET", key, field)
	if err != nil {
		return err
	}
	if bs == nil {
		return nil
	}
	return r.unmarshal(bs, result)
}
func (r *Service) HGetAsMap(key, field string) (map[string]interface{}, error) {
	if r.Redis.Err() != nil {
		return nil, r.Redis.Err()
	}
	bs, err := r.Redis.Do("HGET", key, field)
	if err != nil {
		return nil, err
	}
	if bs == nil {
		return nil, nil
	}
	out := make(map[string]interface{})
	err = r.unmarshal(bs, &out)
	return out, err
}

// result: *map[string]interface{}
func (r *Service) HMGetAll(key string, result interface{}) error {
	if r.Redis.Err() != nil {
		return r.Redis.Err()
	}
	v, err := r.Redis.Do("HGETALL", key)
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	vs := v.([]interface{})
	length := len(vs)
	mType := reflect.TypeOf(result).Elem()
	vType := mType.Elem()
	m := reflect.MakeMapWithSize(mType, length/2)
	for i := 0; i < len(vs); i += 2 {
		ele := reflect.New(vType)
		if bs, ok := vs[i+1].([]byte); ok {
			r.unmarshal(bs, ele.Interface())
		}
		m.SetMapIndex(reflect.ValueOf(string(vs[i].([]byte))), ele.Elem())
	}
	reflect.ValueOf(result).Elem().Set(m)
	return nil
}

func (r *Service) HMGet(key string, result interface{}, fields ...string) error {
	fs := make([]interface{}, len(fields)+1)
	fs[0] = key
	for i := range fields {
		fs[i+1] = fields[i]
	}
	bs, err := r.Redis.Do("HMGET", fs...)
	if err != nil {
		return err
	}
	if bs == nil {
		return nil
	}
	vs := bs.([]interface{})
	length := len(fields)
	sliceType := reflect.TypeOf(result).Elem()
	eleType := sliceType.Elem()
	slice := reflect.MakeSlice(sliceType, length, length)

	for i := 0; i < len(vs); i++ {
		ele := reflect.New(eleType)
		if nil != vs[i] {
			r.unmarshal(vs[i].([]byte), ele.Interface())
		}
		slice.Index(i).Set(ele.Elem())
	}
	reflect.ValueOf(result).Elem().Set(slice)
	return err
}

func (r *Service) HMSet(key string, kvs map[string]interface{}, secondsLifetime int64) error {
	args := make([]interface{}, len(kvs)*2+1)
	args[0] = key
	i := 1
	for k, v := range kvs {
		args[i] = k
		value, err := r.marshal(v)
		if err != nil {
			return err
		}
		args[i+1] = value
		i += 2
	}
	_, err := r.Redis.Do("HMSET", args...)
	if nil != err {
		return err
	}
	if secondsLifetime > 0 {
		err = r.Expire(key, secondsLifetime)
	}
	return err
}

// objs : a list , eg. []xxx
func (r *Service) LSet(key string, objs interface{}, secondsLifetime int64) error {

	arr := reflect.ValueOf(objs)
	length := arr.Len()
	args := make([]interface{}, length+1)
	args[0] = key
	for i := 1; i <= length; i++ {
		args[i], _ = r.marshal(arr.Index(i - 1).Interface())
	}
	_, err := r.Redis.Do("RPUSH", args...)
	if nil != err {
		return err
	}
	if secondsLifetime > 0 {
		err = r.Expire(key, secondsLifetime)
	}

	return err
}
func (r *Service) LGet(key string, index int, value interface{}) error {
	bs, err := r.Redis.Do("LINDEX", key, index)
	if err != nil {
		return err
	}
	if bs1, ok := bs.([]byte); ok {
		if len(bs1) <= 0 {
			return nil
		}
		r.unmarshal(bs1, value)
	}
	return nil
}

// result : is a list point ,eg result = &a ; a is var a []Obj
func (r *Service) LGetAll(key string, result interface{}) error {
	v, err := r.Redis.Do("LRANGE", key, 0, -1)
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	vs := v.([]interface{})
	length := len(vs)
	sliceType := reflect.TypeOf(result).Elem()
	slice := reflect.MakeSlice(sliceType, length, length)
	elemType := sliceType.Elem()

	for i := 0; i < length; i++ {
		elemValue := reflect.New(elemType)
		err := r.unmarshal(vs[i].([]byte), elemValue.Interface())
		if err != nil {
			return err
		}
		slice.Index(i).Set(elemValue.Elem())
	}
	reflect.ValueOf(result).Elem().Set(slice)

	return nil

}
func (r *Service) LLen(key string) (int, error) {
	return redis.Int(r.Redis.Do("LLEN", key))
}

func (r *Service) Publish(channel string, msg interface{}) (int, error) {
	bs, err := r.marshal(msg)
	if err != nil {
		return 0, err
	}
	return redis.Int(r.Redis.Do("PUBLISH", channel, bs))
}
func (r *Service) Subscribe(channel string) (*redis.PubSubConn, error) {
	// defer c.Close()
	psc := redis.PubSubConn{Conn: r.Redis}
	err := psc.Subscribe(channel)
	if err != nil {
		return nil, err
	}
	return &psc, nil
	// for {
	// 	switch v := psc.Receive().(type) {
	// 	case redis.Message:
	// 		fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
	// 	case redis.Subscription:
	// 		fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
	// 	case error:
	// 		return v
	// 	}
	// }

}

// TTL returns the seconds to expire, if the key has expiration and error if action failed.
// Read more at: https://redis.io/commands/ttl
func (r *Service) TTL(key string) (seconds int64, hasExpiration bool, found bool) {
	redisVal, err := r.Redis.Do("TTL", key)
	if err != nil {
		return -2, false, false
	}
	seconds = redisVal.(int64)
	// if -1 means the key has unlimited life time.
	hasExpiration = seconds > -1
	// if -2 means key does not exist.
	found = !(r.Redis.Err() != nil || seconds == -2)
	return
}

func (r *Service) Expire(key string, newSecondsLifeTime int64) error {
	_, err := r.Redis.Do("EXPIRE", key, newSecondsLifeTime)
	return err
}
func (r *Service) Expires(seconds int64, keys ...string) error {
	var err error
	for _, k := range keys {
		err = r.Expire(k, seconds)
	}
	return err
}
func (r *Service) Exists(key string) (bool, error) {
	return redis.Bool(r.Redis.Do("EXISTS", key))
}
func (r *Service) AllExist(keys ...string) (bool, error) {
	length := len(keys)
	kks := make([]interface{}, length)
	for i, k := range keys {
		kks[i] = k
	}

	count, err := redis.Int(r.Redis.Do("EXISTS", kks...))
	return count == length, err
}

// GetAll returns all redis entries using the "SCAN" command (2.8+).
func (r *Service) GetAll() (interface{}, error) {
	redisVal, err := r.Redis.Do("SCAN", 0) // 0 -> cursor

	if err != nil {
		return nil, err
	}

	if redisVal == nil {
		return nil, err
	}

	return redisVal, nil
}

func (r *Service) getKeysConn(c redis.Conn, prefix string) ([]string, error) {
	if err := c.Send("SCAN", 0, "MATCH", prefix+"*", "COUNT", 9999999999); err != nil {
		return nil, err
	}

	if err := c.Flush(); err != nil {
		return nil, err
	}

	reply, err := c.Receive()
	if err != nil || reply == nil {
		return nil, err
	}

	// it returns []interface, with two entries, the first one is "0" and the second one is a slice of the keys as []interface{uint8....}.

	if keysInterface, ok := reply.([]interface{}); ok {
		if len(keysInterface) == 2 {
			// take the second, it must contain the slice of keys.
			if keysSliceAsBytes, ok := keysInterface[1].([]interface{}); ok {
				keys := make([]string, len(keysSliceAsBytes), len(keysSliceAsBytes))
				for i, k := range keysSliceAsBytes {
					keys[i] = fmt.Sprintf("%s", k)
				}

				return keys, nil
			}
		}
	}

	return nil, nil
}
func (r *Service) marshal(v interface{}) ([]byte, error) {
	bs, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

// func (r *Service) unmarshal(v interface{}) (interface{}, error) {
// 	if bs, ok := v.([]byte); ok {
// 		var obj interface{}
// 		err := json.Unmarshal(bs, &obj)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return obj, nil
// 	}
// 	return nil, errors.New("type cast error:redis value of []byte type")
// }
func (r *Service) unmarshal(v interface{}, out interface{}) error {
	if bs, ok := v.([]byte); ok {
		err := json.Unmarshal(bs, out)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("type cast error:redis value of []byte type")
}

// GetKeys returns all redis keys using the "SCAN" with MATCH command.
// Read more at:  https://redis.io/commands/scan#the-match-option.
// func (r *Service) GetKeys(prefix string) ([]string, error) {

// 	return r.getKeysConn(c, prefix)
// }

// GetBytes returns value, err by its key
// you can use utils.Deserialize((.GetBytes("yourkey"),&theobject{})
//returns nil and a filled error if something wrong happens
func (r *Service) GetBytes(key string) ([]byte, error) {
	redisVal, err := r.Redis.Do("GET", key)
	if err != nil {
		return nil, err
	}
	if redisVal == nil {
		return nil, errors.New("no such key:" + key)
	}
	return redis.Bytes(redisVal, err)
}

// Delete removes redis entry by specific key
func (r *Service) Delete(key string) error {
	_, err := r.Redis.Do("DEL", key)
	return err
}

func dial(network string, addr string, pass string) (redis.Conn, error) {
	if network == "" {
		network = "tcp"
	}
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	c, err := redis.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	if pass != "" {
		if _, err = c.Do("AUTH", pass); err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, err
}

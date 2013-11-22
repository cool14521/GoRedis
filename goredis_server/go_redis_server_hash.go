package goredis_server

import (
	. "../goredis"
	"./libs/leveltool"
	// . "./storage"
)

// 获取Hash，不存在则自动创建
func (server *GoRedisServer) hashByKey(key string) (hash *leveltool.LevelHash) {
	server.levelMutex.Lock()
	defer server.levelMutex.Unlock()
	var exist bool
	hash, exist = server.hashtable[key]
	if !exist {
		hash = leveltool.NewLevelHash(server.datasource.DB(), "__hash:"+key)
		server.hashtable[key] = hash
	}
	return
}

func (server *GoRedisServer) OnHGET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	field, _ := cmd.ArgAtIndex(2)
	hash := server.hashByKey(key)
	val := hash.Get(field)
	if val == nil {
		reply = BulkReply(nil)
	} else {
		reply = BulkReply(val)
	}
	return
}

func (server *GoRedisServer) OnHSET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	hash := server.hashByKey(key)
	field, _ := cmd.ArgAtIndex(2)
	value, _ := cmd.ArgAtIndex(3)
	hash.Set(field, value)
	return IntegerReply(1)
}

func (server *GoRedisServer) OnHGETALL(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	hash := server.hashByKey(key)
	elems := hash.GetAll(1000)
	keyvals := make([]interface{}, 0, len(elems)*2)
	for _, elem := range elems {
		keyvals = append(keyvals, elem.Key)
		keyvals = append(keyvals, elem.Value)
	}
	reply = MultiBulksReply(keyvals)
	return
}

func (server *GoRedisServer) OnHMGET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	hash := server.hashByKey(key)
	fields := cmd.Args[2:]
	keyvals := make([]interface{}, 0, len(fields)*2)
	for _, field := range fields {
		val := hash.Get(field)
		keyvals = append(keyvals, field)
		keyvals = append(keyvals, val)
	}
	reply = MultiBulksReply(keyvals)
	return
}

func (server *GoRedisServer) OnHMSET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	keyvals := cmd.Args[2:]
	if len(keyvals)%2 != 0 {
		reply = ErrorReply("Bad field/value paires")
		return
	}
	hash := server.hashByKey(key)
	hash.Set(keyvals...)
	reply = StatusReply("OK")
	return
}

func (server *GoRedisServer) OnHLEN(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	hash := server.hashByKey(key)
	length := hash.Count()
	reply = IntegerReply(length)
	return
}

func (server *GoRedisServer) OnHDEL(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	hash := server.hashByKey(key)
	fields := cmd.Args[2:]
	n := hash.Remove(fields...)
	reply = IntegerReply(n)
	return
}

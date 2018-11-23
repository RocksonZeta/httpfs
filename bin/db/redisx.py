import redis



pool= redis.ConnectionPool(host="test.iqidao.com",port=50002,decode_responses=True)

def getRedis():
	return redis.Redis(connection_pool=pool)
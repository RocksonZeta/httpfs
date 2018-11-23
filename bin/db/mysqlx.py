from __future__ import print_function
import pymysql
import time
from functools import wraps
from db import env

class ConnectionEx(pymysql.connections.Connection):
	"""Mysql Connection"""
	def __init__(self , *args, **kwargs):
		pymysql.connections.Connection.__init__(self,*args, **kwargs)
		self.pageSize = 1000
	def q(self , *args):
		with self.cursor() as cursor:
			cursor.execute(*args)
			return cursor.fetchall()
	def q1(self , *args):
		with self.cursor() as cursor:
			cursor.execute(*args)
			return cursor.fetchone()
	def qy(self , *args):
		with self.cursor() as cursor:
			cursor.execute(*args)
			while True:
				rs = cursor.fetchmany(size=self.pageSize)
				if len(rs) <=0:
					break
				b = yield rs
				if bool == False or len(rs)<self.pageSize :
					break
	def qv(self , *args):
		with self.cursor() as cursor:
			cursor.execute(*args)
			r = cursor.fetchone()
			if r is None :
				return None
			for (_,v) in r.items():
				return v
	def truncate(self , *table):
		for t in table :
			self.q("truncate `"+t+"`")
		return self
	def maxId(self ,table , idFieldName = "id") :
		return self.qv("select ifnull(max("+idFieldName+"),0) from `"+ table+"`") 
	def count(self ,table , idFieldName = "id") :
		return self.qv("select count(*) from `"+ table+"`") 
	def disableFk(self):
		self.q('SET FOREIGN_KEY_CHECKS=0')
		return self
	def enableFk(self):
		self.q('SET FOREIGN_KEY_CHECKS=1')
		return self
	def getCols(self , table):
		s = "select column_name,data_type  from information_schema.columns where table_schema='"+self.db+"' and table_name='"+table+"'"
		return self.q(s)
	def hasTable(self , table):
		return len(self.q("select * from information_schema.tables where table_schema='"+self.db+"' and table_name='"+table+"' limit 1")) >0

	def update(self, table, values , idValue , idName="id"):
		usql = "update `"+table+"` set " 
		for k in values :
			usql += k+"='{"+k+"}',"
		usql = usql[0:len(usql)-1]
		usql += " where "+idName+"="
		if isinstance(idValue, int):
			usql += str(idValue)
		else:
			usql += "'"+str(idValue)+"'"
		usql = usql.format(**values)
		return self.q(usql)
	

def ds(dataSource=None , conName='con'):
	if not dataSource :
		dataSource = env.dataSource
	def mysqli(func ):
		@wraps(func)
		def wrapper(*args , **kwargs):
			# st = time.time()
			varnames = func.__code__.co_varnames
			con = None
			if conName in varnames :
				con = dataSource.createConnection()
				kwargs[conName] = con
			try:
				result = func(*args , **kwargs)
			finally:
				if con is not None :
					con.commit()
					con.close()
			# et = time.time()
			# print(func.__name__, str(et-st)+" sec")
			return result
		return wrapper
	return mysqli
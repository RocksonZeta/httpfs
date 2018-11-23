from __future__ import print_function
import pymysql
import time
import types
from functools import wraps

from db import env
from db import mysqlx


class DataSource:
    def __init__(self,host=None,port=3306,user='root',password=None,db=None,charset='utf8',cursorclass=pymysql.cursors.DictCursor , **kwargs):
        self.host= host
        self.port= port
        self.user= user
        self.password= password
        self.db= db
        self.charset= charset
        self.cursorclass= cursorclass
        self.pymysqlKwargs = kwargs
        
    def createConnection(self):
        """create mysql connection """
        con = mysqlx.ConnectionEx(
						host=self.host,
                        user=self.user,
                        password=self.password,
                        db=self.db,
                        port=self.port,
                        charset=self.charset,
						autocommit=True,
                        cursorclass=self.cursorclass,**self.pymysqlKwargs
						)
        return con
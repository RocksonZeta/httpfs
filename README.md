# httpfs
direct fs base on http

# Installation
1. ffmpeg, x264
2. python3, pip3 install pymysql redis


# Basic Idea
1. store file on fs directly.
2. store file meta info in boltdb.
3. store cluster info in redis.
4. proccess file directly on fs. eg. resize image,compress video etc.
5. using restful style to access files.
6. easy to work with nginx.
7. easy use.

# Roadmap
- [x] distribute storage,write and read.
- [x] basic file handler. image crop and resize ,read zip file directly. compress video.
- [ ] Auto backup.

## dependencies
```
#web
govendor fetch -tree github.com/kataras/iris
#zipx
govendor fetch github.com/RocksonZeta/zipx/^
#validator
govendor fetch github.com/asaskevich/govalidator
#mysql
govendor fetch github.com/go-sql-driver/mysql
#orm
govendor fetch github.com/go-gorp/gorp
#sql null
govendor fetch github.com/guregu/null
#fuse
#github.com/hanwen/go-fuse
#govendor fetch github.com/hanwen/go-fuse/fuse/^
#command parser
govendor fetch gopkg.in/alecthomas/kingpin.v2
#request
govendor fetch github.com/mozillazg/request
#image
govendor fetch github.com/disintegration/imaging
#cmd
govendor fetch github.com/go-cmd/cmd
```

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
import videos
import time
import threading,time,json , sys,os , argparse

import db

import redis

vbrUppers = {
	1080: 120*1024*8,
	720: 80*1024*8,
	480: 40*1024*8,
}
abrUpper = 20 * 1024 * 8
psizes =[[1920,1080],[1280,720],[640,480]]

redisAddr = ""
redisPassword=""
redisDb=0
redisKey=""

args =None

# 0：未上传
# 1：上传中
# 2：上传完成待审核
# 3：审核完成转码中
# 4：转码结束

# 11：上传失败
# 12：审核失败
# 13：转码失败

# 20：资源已删除，不存在

state_mp4ing = 2
state_mp4_ok = 3
state_compressing = 4
state_compress_ok = 5

state_video_review=1

transOk = 4
transFail = 13


class Timer(threading.Thread):
	def __init__(self , m):
		super(Timer ,self).__init__(daemon=True)
		self.m = m
		self.stop= False
		# if args.ra:
		# 	i = args.ra.find(":")
		# 	self.pool= redis.ConnectionPool(host=args.ra[0:i],port=int(args.ra[i+1:]), db=int(args.rd) , password=args.rp,decode_responses=True)

		
	def run(self):
		while not self.stop:
			time.sleep(3)
			progress = str(self.m.progress)[0:5]
			# if args.ra:
			r=db.getRedis()
			r.set(args.rk, progress)
			# print("progress:"+)

def relPath(absPath):
	i = absPath.find(args.root)
	if i ==0 :
		return  absPath[len(args.root):]
	return absPath
def hfLink(absPath):
	if not absPath:
		return None
	return args.cs+relPath(absPath)


def handle(videoFile):
	m = videos.Mp4(videoFile , psizes , vbrUppers, abrUpper)
	timer = Timer(m)
	timer.start()
	m.compressDash()
	# m.snapshot(3)
	timer.stop =True
	# print("result:" , json.dumps(m.result))
	return m

def parseArgs() :
	parser = argparse.ArgumentParser(description='Video compress args.')
	parser.add_argument("-m",required=True , help="method name") # file path
	parser.add_argument("-f",required=True , help="relative file path") # file path
	# parser.add_argument("-ra")
	# parser.add_argument("-rp",default="")
	parser.add_argument("-rk",required=True , help="progress redis key") # redis key
	parser.add_argument("-cs",required=True , help="cluster:server,eg:'static:s1'")# cluster server id.eg: static:s1
	parser.add_argument("-root",required=True,help="httpfs root dir")# root dir of server
	parser.add_argument("-vid" ,required=True , help="video id")# video id
	# parser.add_argument("-rd",type=int,default=0)
	global args 
	args = parser.parse_args()

@db.ds()
def start(vid ,state ,con=None ):
	v = {"state":state,"tsStartTime":int(time.time())}
	con.update("VideoMedia",v ,vid)

@db.ds()
def finishMp4( videoFile , vid ,con=None ):
	v = {"state":state_mp4_ok}
	m = videos.Mp4(videoFile , psizes , vbrUppers, abrUpper)
	rawMp4 = m.toMp4()
	m.captureMp4(3)
	capture1 = m.result.get("capture1")
	capture2 = m.result.get("capture2")
	capture3 = m.result.get("capture3")
	v['hfCapture1']=hfLink(capture1)
	v['hfCapture2']=hfLink(capture2)
	v['hfCapture3']=hfLink(capture3)
	v['hfRawMp4Review'] = hfLink(rawMp4)
	v['tsEndTime'] = int(time.time())
	con.update("VideoMedia",v ,vid)
	con.q("update Video set duration=%s,state=%s where mediaId=%s" ,(m.duration,state_video_review, vid))
@db.ds()
def finishCompress(m , vid ,con=None ):
	'''
	VideoMedia state,duration,rawHeight,rawWidth,fileName,rawSize,Mp41080Size,Mp4720Size,Mp4480Size
	hfRaw,hfRawMp4,hfMp41080,720,480,Mpd,hfCapture1,2,3,
	r =con.q("select * from User limit 1")
	redis  = db.getRedis()
	print(r)
	'''
	print("result:" , m.result)
	v = {}
	if m.isResultOk():
		v["state"]=state_compress_ok
		v["duration"] = m.duration
		v['rawHeight'] = m.height
		v['rawWidth'] = m.width
		raw = m.getResultRaw()
		p1080 = m.result.get(1080)
		p720 = m.result.get(720)
		p480 = m.result.get(480)
		mpd = m.result.get("mpd")
		# capture1 = m.result.get("capture1")
		# capture2 = m.result.get("capture2")
		# capture3 = m.result.get("capture3")
		
		v['hfRawMp4'] = hfLink(raw)
		v['hfMp41080'] = hfLink(p1080)
		v['hfMp4720'] = hfLink(p720)
		v['hfMp4480'] = hfLink(p480)
		if raw :
			v['mp4RawSize'] = os.path.getsize(raw)
		if p1080:
			v['mp41080Size'] = os.path.getsize(p1080)
		if p720:
			v['mp4720Size'] = os.path.getsize(p720)
		if p480:
			v['mp4480Size'] = os.path.getsize(p480)
		v['hfMpd']=hfLink(mpd)
		# v['hfCapture1']=hfLink(capture1)
		# v['hfCapture2']=hfLink(capture2)
		# v['hfCapture3']=hfLink(capture3)
		v['tsEndTime'] = int(time.time())
	else :
		v["state"]=transFail
		v["tsFail"] = m.lastLine[0:100]
	print("videoId:" , vid)
	print("video:" , v)
	con.update("VideoMedia",v ,vid)

def CompressDash(vid , vfile ) :
	start(vid,state_compressing)
	m = handle(vfile)
	finishCompress(m, vid)
def Mp4(vid , vfile ) :
	start(vid,state_mp4ing)
	finishMp4(vfile, vid)

if '__main__' == __name__:
	parseArgs()
	# argv = sys.argv
	if not args.f:
		print("please entry you file.")
		sys.exit(1)
	# fileName = argv[1]
	print(os.path.join(str(args.root),str(args.f)))
	vfile = os.path.normpath(args.root+args.f)
	print('video file:',vfile)
	if not os.path.exists(vfile):
		print("no such file:"+vfile)
		sys.exit(2)
	if args.m == "CompressDash":
		CompressDash(args.vid ,vfile)
	elif args.m=="Mp4":
		Mp4(args.vid ,vfile)
	else:
		print("no such method:" + args.m)

	#python3 bin/mp4.py -m Mp4 -f /video/0/0/9o39m9wuvi/4uie3br1wj.mp4 -root /Users/ququ/projects/go/src/httpfs/testfs -rk video1/progress -cs static:s1 -vid 1
	
	
	# print(args.f)
	
	


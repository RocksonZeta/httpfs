import os,subprocess
import random,json,re
import time
from threading import Thread


'''
依赖：
compress:ffmpeg,x264
dash:mp4fragment,mp4dash
考虑：
vcodec,acodec,vbr,abr,width,height
1.h264,aac OK
vbr,abr在合理范围内，直接resize,音频copy
2.abr>标准
直接压缩音频
直接使用ffmpeg
=====
3.分离音视频，启用x264,压缩视频，压缩音频if
合并
'''

class Mp4(object):
	def __init__(self,media , sizes=None , vbrUppers=None, abrUpper=131584, brLower=600000 , heightLower = 480,stdoutCb=None,stderrCb=None):
		self.media = media
		self.sizes = sizes
		self.vbrUppers = vbrUppers
		self.vbrUppersList = sorted(list(vbrUppers.items()) ,reverse=True )
		self.abrUpper = abrUpper
		self.brLower = brLower
		self.heightLower =heightLower
		self.stdoutCbRaw = stdoutCb
		self.stderrCbRaw = stderrCb
		self.result= {}
		self.lastLine = ""
		self.init()
	def init(self):
		self.streams,self.format = getVideoInfo(self.media)
		# print(self.streams)
		# print(self.format)
		# print(self.streams['video'],self.format)
		
		self.dirpath , self.fileName = os.path.split(self.media)
		self.name,self.ext = os.path.splitext(self.fileName)
		self.fileWithoutExt =os.path.join(self.dirpath , self.name)
		self.ext = self.ext.lower()

		self.aac = self.fileWithoutExt+".aac"
		self.wav = ""

		self.height = int(self.streams['video']['height'])
		self.width = int(self.streams['video']['width'])
		self.br = int(self.format['bit_rate'])
		self.vbr = int(self.streams['video']['bit_rate'])
		self.abr = int(self.streams['audio']['bit_rate'])
		self.fileSize = int(self.format['size'])
		self.duration = float(self.format['duration'])
		v = self.streams['video']
		if 'nb_frames' in v:
			self.nb_frames = int(v['nb_frames'])
		else :
			if 'avg_frame_rate' in v:
				fri = v['avg_frame_rate'].find("/")
				self.nb_frames = int(self.duration)*int(v['avg_frame_rate'][0:fri])
			else:
				self.nb_frames = int(self.duration)*25
		self.setHeight()
		self.setProgresses()
		self.progressIndex = 0
		self.progress = 0
		self.stdoutCb = self.stdoutCbWrap
		self.stderrCb = self.stderrCbWrap


	def isH264(self):
		return 'h264' == self.streams['video']['codec_name']
	def isAac(self):
		return 'aac' == self.streams['audio']['codec_name']
	def videoNeedResize(self):
		return int(self.streams['video']['width']) > self.heightLower
	def audioNeedCompress(self):
		return int(self.streams['audio']['bit_rate']) > self.abrUpper
	# def videoNeedCompress(self , i ):
	# 	return int(self.streams['video']['bit_rate']) > self.sizesVideoBrLower[i]

	def getAcc(self,nero=False):
		if self.isAac():
			cmd = "ffmpeg -i {media} -vn -y "
			if self.audioNeedCompress() :
				cmd += " -acodec aac -vbr 3 -ab "+str(self.abrUpper)+" {aac}"
			else :
				cmd += " -acodec copy "
			cmd += " {aac}"
			runCb(cmd.format(media=self.media,aac=self.aac) ,self.stdoutCb, self.stderrCb)
		else:
			if not nero :
				cmd = "ffmpeg -i {media} -vn -y -acodec aac -vbr 3  {aac}"
				runCb(cmd.format(media=self.media,aac=self.aac) ,self.stdoutCb, self.stderrCb)
			else :
				self.wav = self.fileWithoutExt+".wav"
				cmd1 = 'ffmpeg -i {media}  -vn -c:a pcm_s16le -y -f wav {wav}'
				cmd2 = 'neroAacEnc -q 0.8 -ignorelength -2pass -if {wav} -of {aac}'
				runCb(cmd.format(media=self.media,aac=self.aac) ,self.stdoutCb, self.stderrCb)
	# def filterSizes(self):
	# 	resizes =[]
	# 	for size in self.sizes :
	# 		if size[1] < self.height :
	# 			resizes.append(size)
	# 	if len(resizes) == 0:
	# 		resizes.append([self.width ,self.height])
	# 	if self.height > resizes[0][1] :
	# 		resizes.insert(0,[self.width ,self.height])
	# 	return resizes
	def vbrOk(self):
		if not self.vbrUppers or len(self.vbrUppers) <=0 :
			return False
		ups  = list(self.vbrUppers.items())
		ups.sort()
		ups = list(reversed(ups))
		if self.height > ups[0][0] :
			return self.vbr <= ups[0][1]
		for i in ups :
			if self.height >= i[0]:
				return self.vbr<= i[1] * (self.height/i[0]) * 1.2
		return self.vbr <= ups[-1][1]
	def preprocess(self):
		if self.ext.lower() == ".mov":
			toMp4(self.media , self.streams,self.stdoutCb,self.stderrCb)
			self.media = self.fileWithoutExt+".mp4"
			self.init()
	def compress(self):
		#  vbr is ok ,quick compress
		if self.vbrOk():
			qr = self.quickCompress()
			if qr is not None:
				self.progress = 100
				return qr
		# other ,split media into video and audio
		# video -x264-> mkvs 
		# audio -aac-> aac 
		# merge into mp4s

		
		self.preprocess()
		videoSuffix = ".mkv"
		r = []
		self.getAcc()
		tmpVideos= []
		if self.sizes and len(self.sizes)>0 :
			curVideo = self.media
			for i, size in enumerate(self.sizes):
				self.progressIndex = i
				mp4Video = self.fileWithoutExt+"_tmp_"+formatSize(size)+videoSuffix
				x264Compress(curVideo,mp4Video,size,vbr=self.maxVbr(size[1]),stdoutCb=self.stdoutCb,stderrCb=self.stderrCb)
				tmpVideos.append(mp4Video)
				out = self.fileWithoutExt +"_"+str(i+1)+".mp4"
				merge(mp4Video,self.aac,out)
				r.append(out)
				curVideo = mp4Video
				self.result[size[1]] = out
		else :
			mp4Video = self.fileWithoutExt+"_tmp"+videoSuffix
			x264Compress(self.media,mp4Video,stdoutCb=self.stdoutCb,stderrCb=self.stderrCb)
			tmpVideos.append(mp4Video)
			out = self.fileWithoutExt +"_1.mp4"
			merge(mp4Video,self.aac,out)
			r.append(out)
			self.result[self.height] = out
		if self.wav:
			os.remove(self.wav)
		if self.aac:
			os.remove(self.aac)
		for v in tmpVideos:
			os.remove(v)
		self.progress = 100
		return r
	
	def maxVbr(self , height):
		if height >= self.vbrUppersList[0][0]:
			return self.vbrUppersList[0][1]
		if height <= self.vbrUppersList[-1][0]:
			return self.vbrUppersList[-1][1]
		for i in self.vbrUppersList :
			if height >= i[0]:
				return i[1]
		return 0
	def needCompress(self):
		return self.height >= 1080 and self.br > 100*1024*8 \
		or self.height < 1080 and self.height >= 720 and self.br > 60*1024*8 \
		or self.height < 720 and self.height >= 480 and self.br > 40*1024*8 \
		or self.height <  self.br > 40*1024*8

	def quickCompress(self):
		# if not self.videoNeedResize() and not self.audioNeedCompress():
		# return [toMp4(self.media,self.streams)]
		# if self.needCompress():
		return self.mp4Resize()
		# return []
	def setHeight(self ):
		if not self.sizes :
			return
		hs =[]
		stanardHeight = False
		for size in self.sizes:
			if self.height == size[1]:
				stanardHeight = True
			if self.height >= size[1] :
				hs.append(size)
		if not stanardHeight and len(hs)>0 and self.height> hs[0][1] :
			hs.insert(0 ,[self.width,self.height])
		if not hs :
			hs =[[self.width,self.height]]
		self.sizes = hs

	def setProgresses(self):
		sizeLen = len(self.sizes)
		if sizeLen == 1:
			self.progresses = [100]
		if sizeLen ==2 :
			self.progresses = [60,40]
		if sizeLen ==3 :
			self.progresses = [50,30,20]
		if sizeLen ==4 :
			self.progresses = [45,25,20,10]
		if sizeLen ==5 :
			self.progresses = [45,25,15,10,5]
		if sizeLen >=6 :
			self.progresses = [40,20,10]
			for i in range(sizeLen-3):
				self.progresses.append(30/(sizeLen-3))

	def stdoutCbWrap(self,line):
		self.calcProgress(line)
		self.lastLine = line
		if self.stdoutCbRaw :
			self.stdoutCbRaw(line)
	def stderrCbWrap(self,line):
		self.calcProgress(line)
		self.lastLine = line
		if self.stderrCbRaw :
			self.stderrCbRaw(line)
	def calcProgress(self,line):
		# print("line:",line)
		if not line or not hasattr(self,"progresses"):
			return
		rate = 0
		if "frame=" == line[0:6]:
			#frame=  123 fps=12
			end = line.find("fps=")
			frameNo = int(line[7:end])
			rate = frameNo / self.nb_frames
		if rate==0 and '['==line[0] :
			#[0.3%]
			m = re.match(r'^\[(\d*\.\d*)\%\].*', line[0:10])
			if m :
				rate = float(m.group(1))/100
		if rate == 0 :
			# 123 frames
			m = re.match("^(\d+)\s+frames" , line[0:20])
			if m :
				rate = int(m.group(1))/self.nb_frames
		progress = sum(self.progresses[0:self.progressIndex]) + self.progresses[self.progressIndex]*rate
		if progress >self.progress :
			self.progress = progress

	def mp4Resize(self):
		cmd1 = 'ffmpeg -i {media} -y'
		if self.sizes:
			cmd1 += ' -s {size}'
		cmd2 = cmd1 +' {output}'
		# if not self.isAac():
		# 	cmd1 += " -acodec aac"
		if self.audioNeedCompress() :
			cmd1 += " -acodec aac -vbr 3 -ab "+str(self.abrUpper)
		else :
			cmd1 += " -acodec copy "
		# else :
		# 	cmd1 += " -acodec copy"
		if not self.isH264():
			cmd1 += " -vcodec h264"
		# else :
		# 	cmd1 += " -vcodec copy"
		cmd1 += ' {output}'
		cmd2 = cmd1
		r = []
		if self.sizes is not None:
			# print("sizes:",self.sizes , self.height)
			cur = self.media
			for i,size in enumerate(self.sizes):
				self.progressIndex = i
				if size[1] >= self.height and self.ext==".mp4":
					if size[1] == self.height:
						r.append(self.media)
						self.result[size[1]] = self.media
					continue
				output = self.fileWithoutExt+"_"+str(i+1)+".mp4"
				if i ==0:
					runCb(cmd1.format(media=cur,size=formatSize(size) , output=output) ,self.stdoutCb, self.stderrCb)
				else :
					runCb(cmd2.format(media=cur,size=formatSize(size) , output=output) ,self.stdoutCb, self.stderrCb)
				cur = output
				r.append(output)
				self.result[size[1]] = output
		else :
			output = self.fileWithoutExt+"_1.mp4"
			runCb(cmd1.format(media=self.media , output=output) ,self.stdoutCb, self.stderrCb)
			r.append(output)
			self.result[self.height] = output
		return r

	def dash(self , *mp4s):
		out = self.name+".mpd"
		dash(self.dirpath , out , *mp4s)
		self.result["mpd"] = os.path.join(self.dirpath , out)

	def compressDash(self):
		self.dash(*self.compress())

	def snapshot(self , n):
		v = self.result.get(720)
		if not v :
			v = self.result.get(self.sizes[-1][1])
		if not v :
			return
			
		for i in range(1,n+1) :
			img = os.path.join(self.dirpath , self.name+str(i)+".jpg")
			snapshot(v , img , random.randint(0,int(self.duration)))
			self.result["capture"+str(i)] =img

	def getResultRaw(self):
		for k in self.result :
			if k == self.height:
				return self.result[k]
		return None

	def isResultOk(self):
		for k in self.result :
			if isinstance(k , int) and self.result[k]:
				if not os.path.exists(self.result[k]) :
					return False
		return True


def getVideoInfo(media):
	prob='ffprobe -v quiet -print_format json -show_format -show_streams {}'
	c = prob.format(media)
	state, stdout,stderr = run(c)
	video_info = json.loads(stdout)
	streams = {s['codec_type'] : s for s in video_info['streams']}
	format = video_info['format']

	return streams,format

def printVideoInfo(media):
	prob='ffprobe -v quiet -show_format -show_streams {}'
	c = prob.format(media)
	state, stdout,stderr = run(c)
	return stdout,state


def readStream(stream , cb=None) :
	line = ''
	while True :
		b = stream.read(1)
		if not b :
			if cb :
				cb(line)
			break
		if '\r' == b or '\n' ==b :
			# output.append(line.strip())
			if cb :
				cb(line)
			line = ''
			continue
		line +=b


def runCb(cmd  , stdoutCb=None , stderrCb=None, timeout=None,throwError=True,cwd=None,cmdStrCb=None) :
	'''execute cmd callback stdout,stderr line by line
	'''
	cmd = cmd
	if cmdStrCb:
		cmdStrCb(cmd)
	# print("run:"+cmd)
	p = subprocess.Popen(cmd,stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE , universal_newlines=True ,shell=True,cwd=cwd)
	# stdout = []
	# stderr = []
	stdoutThread = Thread(target=readStream,daemon=False,args=(p.stdout,stdoutCb))
	stderrThread = Thread(target=readStream,daemon=False,args=(p.stderr,stderrCb))
	
	stdoutThread.start()
	stderrThread.start()
	state =0
	# try :
	state = p.wait(timeout=timeout) 
	if state!=0 and throwError:
		raise(Exception('Execute cmd:"'+cmd+'" failed. state:'+str(state)))
	# except:
	# 	return state
	stdoutThread.join()
	stderrThread.join()
	return state


def run(cmd,throwError=True, cwd=None):
	stdout ,stderr = [],[]
	state = runCb(cmd,lambda l:stdout.append(l) , lambda l:stderr.append(l) , throwError=throwError,cwd=cwd)
	return state ,'\n'.join(stdout) , '\n'.join(stderr)

def appendFileName(filename , s):
	i = filename.rfind('.')
	if -1==i :
		return filename+s
	return filename[0:i]+s+filename[i:]
def formatSize(size):
	if len(size)<1:
		return ""
	if len(size)<2:
		return size[0]+""
	return str(size[0])+"x"+str(size[1])

def x264CopyResize(media , output ,size,stdoutCb=None,stderrCb=None):
	cmd = 'ffmpeg -i {media} -y -an -s {size} {output}'
	c = cmd.format(media=media , output=output , size=formatSize(size))
	runCb(c, stdoutCb,stderrCb)
def x264Compress(media,video,size=None,vbr = 0,stdoutCb=None,stderrCb=None):
	cmd = 'x264 --threads auto --crf 26 --preset 6 --subme 10 --ref 9 --bframes 14 --b-adapt 2 --qcomp 0.55 --psy-rd 0:0 --keyint 360 --min-keyint 1 --aq-strength 0.9 --aq-mode 3'
	if vbr >0 :
		cmd += " -B "+str(vbr//1024)
	cmd +=' -o {video} {media}'
	# cmd='x264 --threads auto --crf 26 --preset medium --me umh --tune film -o {video} {media}'
	# --vf resize:960,720,,,,lanczos
	cmd3Resize = ' --vf resize:{width},{height},,,,lanczos'
	# cmd3Resize = ' --vf resize:width={width},height={height},method=spline'
	c = cmd
	if size and len(size)==2:
		c += cmd3Resize
		c = c.format(media=media , video=video ,width=size[0] , height=size[1])
	else :
		c = c.format(media=media , video=video)
	runCb(c , stdoutCb, stderrCb)

def merge(video,audio,output):
	cmd = 'ffmpeg -i {video} -i {audio} -c copy -y {output}'
	# cmd='ffmpeg -i {video} -i {audio} -vcodec copy -acodec copy -y {output}'
	run(cmd.format(video=video,audio=audio,output=output))

def resize(src_file , dst_file , size , threads=8):
	cmd = 'ffmpeg -i {} -y -s {} -threads {} {}'
	run(cmd.format(src_file,size,threads,dst_file))

def dash(output_dir,mpd,*src_files):
	# cwd = os.path.dirname(src_files[0])
	name ,_=os.path.splitext(mpd)
	# {in} {out}
	cmd1 = 'mp4fragment --fragment-duration 10000 {} {}'
	frags = []
	fragbases = []
	for sf in src_files :
		frag = appendFileName(sf , "_frag")
		frags.append(frag)
		fragbases.append(os.path.basename(frag))
		run(cmd1.format(sf , frag))
	cmd2 = 'mp4dash -f --mpd-name={mpd} --subtitles --exec-dir={output_dir} --media-prefix={name} --no-split --profiles=on-demand -o {output_dir} {mp4s}'
	run(cmd2.format(mpd=os.path.basename(mpd),name=name,output_dir=output_dir,mp4s=' '.join(fragbases)) , cwd=output_dir)
	for f in frags:
		os.remove(f)
	# exe_cmd('mp4dash --help')


def clip(src_file ,dst_file,ss='00:00:00',duration=30):
	cmd = 'ffmpeg -ss {ss} -t {duration} -accurate_seek -i {src_file} -codec copy -y -avoid_negative_ts 1 {dst_file}'
	r = run(cmd.format(src_file=src_file,dst_file=dst_file,ss=ss,duration=duration))
	return r

def snapshot(src_file,dst_file ,ss=30):
	# ss = ss + random.randint(0,60)
	cmd='ffmpeg -ss {ss} -i {src_file} -v quiet -y -f image2  -vframes 1 {dst_file}'
	r = run(cmd.format(src_file=src_file,dst_file=dst_file,ss=ss))
	return r


def toMp4(media ,streams,stdoutCb=None,stderrCb=None):
	name,ext = os.path.splitext(media)
	cmd = "ffmpeg -i {media} "
	if streams['video']['codec_name']!='h264' :
		cmd += " -vcodec h264"
	# else :
	# 	cmd +=" -vcodec copy"
	if streams['audio']['codec_name']!='aac' :
		cmd +=" -acodec aac"
	# else :
	# 	cmd +=" -vcodec copy"
	if ext==".mp4":
		return media
	cmd += " -y {output}"
	c = cmd.format(media=media,output=name+".mp4")
	runCb(c,stdoutCb,stderrCb)
	return name+".mp4"


# if '__main__' == __name__:

# 	psizes =[[1920,1080],[1280,720],[640,480]]
# 	m = Mp4("/Users/ququ/Movies/test1/1.mp4" , psizes)
# 	m.compress()
	# m.compressDash()
	# m.dash("/Users/ququ/Movies/test4/mal052_lec01_1.mp4","/Users/ququ/Movies/test4/mal052_lec01_2.mp4","/Users/ququ/Movies/test4/mal052_lec01_3.mp4")

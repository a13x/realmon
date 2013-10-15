import socket
import psutil
import platform
import realmon_pb2 as realmon
import uuid
import threading

MCAST_GRP = '239.255.13.0'
MCAST_PORT = 9000


class RealMon(threading.Thread):
    def __init__(self, service_name, frequency=10, timeout=30):
        self.service_name = service_name
        self.id = uuid.uuid4().hex
        self.frequency = frequency
        self.timeout = timeout
        super(RealMon, self).__init__(name="%s-realmon" % service_name)
        self.stop_event = threading.Event()

    def run(self):
        while not self.stop_event.is_set():
            realmon_stats = gather_realmon(self.id, self.service_name, self.frequency, self.timeout)
            sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM, socket.IPPROTO_UDP)
            sock.setsockopt(socket.IPPROTO_IP, socket.IP_MULTICAST_TTL, 2)
            stats = realmon_stats.SerializeToString()
            sock.sendto(stats, (MCAST_GRP, MCAST_PORT))
            print("sleeping a bit")
            self.stop_event.wait(self.frequency)

    def stop(self):
    	print("stopping realmon")
    	self.stop_event.set()


def gather_realmon(uuid, service, frequency, timeout):
	cpu_percent = psutil.cpu_times_percent()
	pb_cpu_percent = realmon.CpuPercent(
		user=cpu_percent.user,
		nice=cpu_percent.nice,
		system=cpu_percent.system,
		idle=cpu_percent.idle,
		)
	virt_mem = psutil.virtual_memory()
	pb_memory = realmon.Memory(
		total=virt_mem.total,
		available=virt_mem.available,
		percent=virt_mem.percent,
		used=virt_mem.used,
		free=virt_mem.free,
		active=virt_mem.active,
		inactive=virt_mem.inactive,
		wired=virt_mem.wired
		)
	pb_runtime = realmon.Runtime(
		platform=platform.platform(),
		release=platform.release(),
		version=platform.version(),
		num_cpus=psutil.NUM_CPUS,
		cpu_times_percent=pb_cpu_percent,
		memory=pb_memory
		)
	pb_runtime.cpu_percent.extend(psutil.cpu_percent(percpu=True))
	pb_report = realmon.Report(
		frequency=frequency,
		timeout=timeout,
		service=service,
		uuid=uuid,
		status_url="something",
		runtime=pb_runtime
		)
	return pb_report

import sys
import time
import mon

def main():
	service = sys.argv[1] if len(sys.argv) in [2,3] else "test_service"
	sleep = int(sys.argv[2]) if len(sys.argv) == 3 else 1
	print "Service: ", service
	print "Interval: ", sleep
	rm_thread = mon.RealMon(service_name=service, frequency=sleep, timeout=sleep*3)
	rm_thread.start()
	while 1:
		try:
			time.sleep(1)
		except KeyboardInterrupt:
			rm_thread.stop()
			break

if __name__ == '__main__':
	main()

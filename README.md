# Real Monitoring

Instead of actively monitoring your service instances, why not letting them
"phone home" instead, giving you all the metrics you really need, in real
time.

Realmon uses UDP multicasts to "phone home", which includes runtime information
about the system its running on. The frequency of phoning home is controlled
by frequency parameter of the RealMon class.

# How it works?

When service is started and RealMon wired in, it will starts multicasting with
a specified frequency. The timeout parameter specifies what's acceptable time
for a service not to phone home. This can be tied with some monitoring hooks
like sending alerts if a timeout is reached and service hasn't phone in.

### Usage

See test_client.py on how to run realmon as part of your python service.

### Visualisation

See gathersvc.go, a simple Golang program that collects UDP multicasts and
streams it via SSE to a browser.

### TODO

Make visualisation bit prettier :)


# Traffic Control Docker

[![Docker pulls](https://img.shields.io/docker/pulls/codyguo/tc-docker.svg?label=docker+pulls)](https://hub.docker.com/r/codyguo/tc-docker)
[![Docker stars](https://img.shields.io/docker/stars/codyguo/tc-docker.svg?label=docker+stars)](https://hub.docker.com/r/codyguo/tc-docker)

**Traffic Control Docker** - network download rate/ceil limiting of network packets using only container labels.

## Running

First run Traffic Control Docker daemon in Docker. The container needs `privileged` capability and the `host` network mode to manage network interfaces on the host system, `/var/run/docker.sock` and `/var/run/docker/netns` volume allows to observe Docker events and query container details.

```bash
docker run -d \
    --name tc-docker \
    --network host \
    --privileged \
    --restart always \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /var/run/docker/netns:/var/run/docker/netns:shared \
    codyguo/tc-docker
```

> You can also pass `DOCKER_HOST` and `DOCKER_API_VERSION` environment variables, which default to `unix:///var/run/docker.sock` and `1.40`.

This repository contains `docker-compose.yml` file in root directory, you can use it instead of manually running `docker run` command. Newest version of image will be pulled automatically and the container will run in daemon mode.

```bash
git clone https://github.com/CodyGuo/tc-docker.git
cd tc-docker
docker-compose up -d
```

## Usage

After the daemon is up it scans all running containers and starts listening for `container:start` events triggered by Docker Engine. When a new container is up and contains `org.label-schema.tc.enabled` label set to `1`, Traffic Control Docker starts applying network traffic rules according to the rest of the labels from `org.label-schema.tc` namespace it finds.

Traffic Control Docker recognizes the following labels:

* `org.label-schema.tc.enabled` - when set to `1` the container network rules will be set automatically, any other value or if the label is not specified - the container will be ignored
* `org.label-schema.tc.rate` - Bandwidth or rate limit for the container, Maximum rate this class and all its children are guaranteed, accepts a floating point number, followed by a unit, or a percentage value of the device's speed (e.g. 70.5%). Following units are recognized:
    * `bit`, `kbit`, `mbit`, `gbit`, `tbit`
    * `bps`, `kbps`, `mbps`, `gbps`, `tbps`
    * to specify in IEC units, replace the SI prefix (k-, m-, g-, t-) with IEC prefix (ki-, mi-, gi- and ti-) respectively
* `org.label-schema.tc.ceil` - Bandwidth or ceil limit for the container, Maximum rate at which a class can send, if its parent has bandwidth to spare. Defaults to the configured rate, accepts a floating point number, followed by a unit, or a percentage value of the device's speed (e.g. 70.5%). Following units are recognized:
    * `bit`, `kbit`, `mbit`, `gbit`, `tbit`
    * `bps`, `kbps`, `mbps`, `gbps`, `tbps`
    * to specify in IEC units, replace the SI prefix (k-, m-, g-, t-) with IEC prefix (ki-, mi-, gi- and ti-) respectively

> Read the [tc command manual](http://man7.org/linux/man-pages/man8/tc.8.html) to get detailed information about parameter types and possible values.

Run the `ubuntu` container, specify all possible labels and try to `iperf -c 127.0.0.1 -i 1 -n 100M -p 5001`.

```bash
docker run -it \
    -p 5001:5001 \
	--label "org.label-schema.tc.enabled=1" \
	--label "org.label-schema.tc.rate=1mbps" \
	--label "org.label-schema.tc.ceil=10mbps" \
    ubuntu sh -c " \
    apt-get update \
    && apt-get install iperf \
    && iperf -s"
```

You should see output similar to shown below, `iperf -s` The download bandwidth is limited to between 1Mbps and 10Mbps.

```
------------------------------------------------------------
Server listening on TCP port 5001
TCP window size: 85.3 KByte (default)
------------------------------------------------------------
[  4] local 172.17.0.7 port 5001 connected with 127.0.0.1 port 39406
[ ID] Interval       Transfer     Bandwidth
[  4]  0.0-46.6 sec  42.4 MBytes  7.64 Mbits/sec
[  4] local 172.17.0.7 port 5001 connected with 127.0.0.1 port 39442
[  4]  0.0- 3.6 sec  3.25 MBytes  7.64 Mbits/sec
```

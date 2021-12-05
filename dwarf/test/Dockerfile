FROM hitzhangjie/linux101:latest

USER root

ENV	GOPROXY=https://goproxy.cn,direct

RUN mkdir /root/dwarftest && \
	yum install -y libdwarf-tools.x86_64 && \
	yum install -y binutils.x86_64

WORKDIR /root/dwarftest

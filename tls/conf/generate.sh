#!/bin/bash -e

# 生成CA的key和自签名证书
openssl genrsa -out ca.key 4096
openssl req -new -x509 -sha256 -days 730 -key ca.key -out ca.crt
openssl x509 -noout -text -in ca.crt

# 生成server的key和证书签名请求
openssl genrsa -out server.key 2048
openssl req -new -sha256 -key server.key -out server.csr

# 生成client的key和证书签名请求
openssl genrsa -out client.key 2048
openssl req -new -sha256 -key client.key -out client.csr

# 为证书签名请求生成证书
openssl x509 -req -days 365 -sha256 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 1 -out server.crt
openssl x509 -req -days 365 -sha256 -in client.csr -CA ca.crt -CAkey ca.key -set_serial 1 -out client.crt

# 将key文件转换成keystore pem格式的
openssl pkcs8 -topk8 -inform pem -in server.key -outform pem -nocrypt -out server.pem
openssl pkcs8 -topk8 -inform pem -in client.key -outform pem -nocrypt -out client.pem




#!/bin/bash -e

rm -rf certs
mkdir certs
cd certs

# 生成CA的key和自签名证书
echo "generate ca.key, ca.crt, then check"
openssl genrsa -out ca.key 4096
openssl req -new -x509 -sha256 -days 365 -key ca.key -subj "/C=CN/ST=Guangdong/L=Shenzhen/O=Tencent/CN=MyRootCA" -out ca.crt
openssl x509 -noout -text -in ca.crt

# 生成server的key和证书签名请求
echo "generate server.key, server.csr"
openssl genrsa -out server.key 2048
openssl req -new -sha256 -key server.key -out server.csr -subj "/C=CN/ST=Guangdong/L=Shenzhen/O=Tencent/OU=DFM/CN=localhost"

# 生成client的key和证书签名请求
echo "generate client.key, client.csr"
openssl genrsa -out client.key 2048
openssl req -new -sha256 -key client.key -out client.csr -subj "/C=CN/ST=Guangdong/L=Shenzhen/O=Tencent/OU=DFM/CN=localhost"

# 为证书签名请求生成证书
echo "generate cert for server.csr/client.csr"
openssl x509 -req -days 365 -sha256 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 1 -out server.crt -extfile <(printf "subjectAltName=DNS:localhost1,DNS:localhost")
openssl x509 -req -days 365 -sha256 -in client.csr -CA ca.crt -CAkey ca.key -set_serial 1 -out client.crt -extfile <(printf "subjectAltName=DNS:localhost2,DNS:localhost")

# 将crt文件转换成keystore pem格式的
echo "convert server.key/client.key to server.pem/client.pem"
openssl pkcs8 -topk8 -inform pem -in server.key -outform pem -nocrypt -out server.pem
openssl pkcs8 -topk8 -inform pem -in client.key -outform pem -nocrypt -out client.pem

cd -
all: server.go
	# see: “x509: certificate relies on legacy Common Name field” error
	# https://jfrog.com/knowledge-base/general-what-should-i-do-if-i-get-an-x509-certificate-relies-on-legacy-common-name-field-error/ 
	GODEBUG=x509ignoreCN=0 go build -v && ./server

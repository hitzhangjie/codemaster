all: server.go
	# error: “x509: certificate relies on legacy Common Name field” error
	# see: https://jfrog.com/knowledge-base/general-what-should-i-do-if-i-get-an-x509-certificate-relies-on-legacy-common-name-field-error/ 
	#
	# GODEBUG=x509ignoreCN=0 go build -v && ./server 
	#
	# but go1.17 deprecates the GODEBUG=x509ignoreCN, so we must use SANs instead,
	# see: https://github.com/virtualmin/virtualmin-gpl/issues/462#issuecomment-1273719619
	go build -v && ./server
	
	

CGIFILENAME=hello.cgi
DEPLOY_DESTINATION=naruh@naruh.sakura.ne.jp:www/hello/

deploy: $(CGIFILENAME)
	scp $(CGIFILENAME) $(DEPLOY_DESTINATION)

$(CGIFILENAME): *.go
	GOOS=freebsd go build -o $(CGIFILENAME)

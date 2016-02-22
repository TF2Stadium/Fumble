static:
	go build -a -tags netgo -v

docker: 
	go build -a -tags netgo -v -o fumble
	docker build -t tf2stadium/fumble .

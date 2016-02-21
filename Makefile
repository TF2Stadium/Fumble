static:
	go build -ldflags "-w -linkmode external -extldflags -static" -v -o fumble

docker: 
	go build -ldflags "-w -linkmode external -extldflags -static" -v -o fumble
	docker build -t tf2stadium/fumble .

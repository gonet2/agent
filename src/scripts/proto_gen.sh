#!/bin/sh

##################################################
###   client proto & api
##################################################
printf "package client_handler\n" > proto.go
gawk -f proto.awk proto.txt >> proto.go 
gawk -f proto_func.awk proto.txt >> proto.go 

printf "package client_handler\n" > api.go
printf "\n" >> api.go
printf "import \"misc/packet\"\n" >> api.go
printf "import . \"types\"\n" >> api.go
printf "\n" >> api.go

gawk -f api.awk api.txt >> api.go 
gawk -f api_rcode.awk api.txt >> api.go

printf "var Handlers map[int16]func(*Session, *packet.Packet) []byte\n" >> api.go
printf "func init() {\n" >> api.go
printf "Handlers = map[int16]func(*Session, *packet.Packet) []byte {\n" >> api.go
gawk -f api_bind_req.awk api.txt >> api.go 
printf "}" >> api.go
printf "}" >> api.go

mv -f proto.go ../client_handler
mv -f api.go ../client_handler
go fmt ../client_handler

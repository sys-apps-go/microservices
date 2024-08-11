./compile.sh
go run authserver/authserver.go &
sleep 2
go run catalogserver/catalogserver.go &
sleep 2
go run apiserver/apiserver.go &
sleep 2
curl "http://localhost:50061/getProduct?id=1"
echo "\n"
go run client/client.go

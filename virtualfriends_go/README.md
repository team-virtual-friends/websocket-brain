## using protobuf message

1. install protobuf
```sh
> brew install protobuf
> protoc --version
libprotoc 24.2
```

2. compile proto
```sh
cd ./virtualfriends/web_socket/virtualfriends_proto
protoc --csharp_out=./ ./ws_message.proto
protoc --python_out=./ ./ws_message.proto
# go will generate the folder virtualfriends_go we should use ../
protoc --go_out=../ ./ws_message.proto
```
package wire

// Install before go generate:                                                                                                                                                                              
//  wget https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protoc-3.6.1-linux-x86_64.zip                                                                                                 
//  unzip protoc-3.6.1-linux-x86_64.zip -x readme.txt -d /usr/local/                                                                                                                                        
//  go get -u github.com/golang/protobuf/protoc-gen-go                                                                                                                                                      

//go:generate protoc --go_out=plugins=grpc:./ event.proto service.proto

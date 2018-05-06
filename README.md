# iofactory
golang io reader, writer extensions

This package includes useful ioutil extensions that can be used to manipulate or optimize reading and writing of I/O.  The project was inspired by a need to create a readahead io.Reader to better proxy network traffic.  The readahead allows the network proxy to read the full contents of an HTTP request from a backend server into memory and then slowly write it out to the client over its slower network connection.  This pattern frees the backend servers from having to deal with slow client network connections.  Addionally, there are plans to create a multiplexer / demultiplexer to allow many socket connections to be tunneled over a single socket connection while allowing for priority traffic management.

# usage 

To get the package use `go get -u github.com/npart/iofactory`.  The tests have pretty extensive use cases for these objects, which should be fairly self explanatory.

### MaxSizeReader

The MaxSizeReader will create a reader where each resulting read will be limited to a maximum size.  ioutil.ReadAll(), for example, may call Read() with a very large buffer, but the MaxSizeReader will limit each read to N bytes.  This is useful if the goal is to limit reads to smaller chunks.

```Go
# File
input, _ := os.Open("input.txt")
chunkedInput := NewMaxSizeReader(input, 1024) // read the whole file, but only read 1024 bytes at a time
io.Copy(os.Stdout, chunkedInput)

# Network (echo back to client)
ln, err := net.Listen("tcp", ":8080")
for {
  conn, err := ln.Accept()
  go func() {
    defer conn.Close()
    chunkedInput := NewMaxSizeReader(conn, 16) // echo full input stream, but a maximum of 16 bytes at a time  
    io.Copy(conn, chunkedInput)
  }()
}
```

### RandomSizeReader

The RandomSizeReader will create a reader where each resulting read will be of a random size (maximum of the size of the input buffer).  Calling this reader with a buffer of length 1024, for example, will result in a read that is between 1 and 1024 bytes in size.  This is particularly useful to add some randomness to test buffered readers that may work perfectly fine if every read is of the same, known size, but may break when reads of random sizes are requested.  This reader is particularly interesting when combined with the MaxSizeReader.

```Go
# Bytes
input, _ := os.Open("input.txt")
randomSizeReader := NewRandomSizeReader(input)
for {
  buf := make([]byte, 1024)
  n, err := randomSizeReader.Read(buf) // each read will be between 1 and 1024 bytes
}

# Bytes with maximum read size
input, _ := os.Open("input.txt")
randomSizeReader := NewMaxSizeReader(NewRandomSizeReader(input), 128)
// the whole file will be read to buf, but each read will be between 1 and 128 bytes 
// even though ioutil.ReadAll() will call Read() with a much larger buffer size
buf, err := ioutil.ReadAll(randomSizeReader)
```


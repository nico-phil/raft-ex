package main

func main() {
	server := NewServer()
	err := server.Start()
	if err != nil {
		panic(err)
	}
}

package main

import "os"

func main() {
	test2()
	os.Exit(1) // want "os.Exit in main func"
}

func test2() {
	os.Exit(2)
}

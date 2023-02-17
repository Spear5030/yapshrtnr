package pkg2

import "os"

func main() {
	os.Exit(3) // want "os.Exit in main func"
}

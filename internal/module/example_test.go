package module

import "fmt"

func ExampleShortingURL() {
	short, err := ShortingURL("https://asdawasda.ee")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(short)
	}

	short, err = ShortingURL("ht://asdawasda.ee")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(short)
	}

	short, err = ShortingURL("14432")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(short)
	}
}

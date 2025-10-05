package main

import (
	"fmt"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	client := NewClient()

	body, err := client.Get("https://funpay.com/lots/3740/")
	if err != nil {
		panic(err)
	}

	fmt.Println("Ответ: ", string(body))

}

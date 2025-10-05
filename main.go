package main

import (
	"FunPay-Core/pkg"
)

func main() {
	client := pkg.NewClient()
	client.GetLots("https://funpay.com/lots/221/")
	//client.GetLots("https://funpay.com/lots/222/")
	//client.GetLots("https://funpay.com/lots/223/")
	//client.GetLots("https://funpay.com/lots/224/")
	//client.GetLots("https://funpay.com/lots/225/")

}

package main

import (
"fmt"
        "os"
"gopkg.in/gomail.v2"
)

func main() {
	m := gomail.NewMessage()
	m.SetHeader("From", "gosmartwithgo@gmail.com")
	m.SetHeader("To", "saurabhxdev@gmail.com")
	m.SetHeader("Subject", "Test via manual")
	m.SetBody("text/plain", "Hello")

    pass := os.Getenv("SMTP_PASSWORD")
	d := gomail.NewDialer("smtp.gmail.com", 587, "gosmartwithgo@gmail.com", pass)

	if err := d.DialAndSend(m); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Success!")
	}
}

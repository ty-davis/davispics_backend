package main

import (
	"fmt"
)

type Booking struct {
	Name	 string `json:"name"`
	Email	string `json:"email"`
	Phone	string `json:"phone"`
	FirstDateTime string `json:"first_datetime"`
	SecondDateTime string `json:"second_datetime"`
	ThirdDateTime string `json:"third_datetime"`
	Type	 string `json:"type"`
	Comments string `json:"comments"`
	Captcha  string `json:"captcha"`
}


func sendBookingEmail(booking Booking) []byte {
	msg := []byte(fmt.Sprintf(
		"Subject: New Booking Request\n\n"+
			"Name: %s\n"+
			"Email: %s\n"+
			"Phone: %s\n"+
			"First Date/Time: %s\n"+
			"Second Date/Time: %s\n"+
			"Third Date/Time: %s\n"+
			"Type: %s\n\n"+
			"Comments:\n%s\n",

		booking.Name, booking.Email, booking.Phone, booking.FirstDateTime, booking.SecondDateTime, booking.ThirdDateTime, booking.Type, booking.Comments))

	return msg
}

func (b Booking) ToEmailBytes() []byte {
	return sendBookingEmail(b)
}

func (b Booking) GetCaptcha() string {
	return b.Captcha
}

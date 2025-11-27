package main

import (
	"fmt"
)

type Question struct {
	Name	 string `json:"name"`
	Email	string `json:"email"`
	Phone	string `json:"phone"`
	Question string `json:"question"`
	Captcha  string `json:"captcha"`
}


func sendQuestionEmail(question Question) []byte {
	msg := []byte(fmt.Sprintf(
		"Subject: New Question Asked\n\n"+
			"Name: %s\n"+
			"Email: %s\n"+
			"Phone: %s\n"+
			"Question:\n%s\n",

		question.Name, question.Email, question.Phone, question.Question))

	return msg
}


func (q Question) ToEmailBytes() []byte {
	return sendQuestionEmail(q)
}

func (q Question) GetCaptcha() string {
	return q.Captcha
}

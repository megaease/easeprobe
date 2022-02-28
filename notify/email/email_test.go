package email_test

import (
	"testing"

	"github.com/megaease/easeprobe/notify/email"
)

func TestSendMail(t *testing.T) {
	c := email.NotifyConfig{
		Server: "smtp.exmail.qq.com:465",
		User:   "noreply@megaease.com",
		Pass:   "644D4u43n",
		To:     "service@megaease.com;chenhao@megaease.com,haoel@163.com",
	}

	subject := "Test"
	message := "This is the test email sent by EaseProbe"
	if err := c.SendMail(subject, message); err != nil {
		t.Fatalf("error: %v\n", err)
	}
}

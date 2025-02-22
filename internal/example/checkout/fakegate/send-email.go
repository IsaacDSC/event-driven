package fakegate

import (
	"fmt"
	"math/rand"
	"time"
)

func SentEmail(email string) error {
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(3) == 0 {
		return fmt.Errorf("failed to send email to: %s", email)
	}

	fmt.Println("Email sent to: ", email)
	return nil
}

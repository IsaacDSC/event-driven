package fakegate

import (
	"fmt"
	"math/rand"
	"time"
)

func SentPayment(clientName string, total float64) error {
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(3) == 0 {
		return fmt.Errorf("failed to send email to: %s", clientName)
	}

	fmt.Println("Charged to account this client: ", clientName, " with total: ", total)
	return nil
}

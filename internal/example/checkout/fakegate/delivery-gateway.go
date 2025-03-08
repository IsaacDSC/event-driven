package fakegate

import (
	"fmt"
	"github.com/IsaacDSC/event-driven/internal/example/checkout/entities"
	"math/rand"
	"time"
)

func SentToScheduleDelivery(products []entities.Product) error {
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(3) == 0 {
		return fmt.Errorf("failed to schedule delivery for products: %v", products)
	}

	fmt.Println("Delivery scheduled for products: ", products)
	return nil
}

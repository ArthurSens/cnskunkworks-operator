package runtime

import (
	"sync"

	"github.com/cloud-native-skunkworks/cnskunkworks-operator/pkg/subscription"
)

func RunLoop(subscriptions []subscription.ISubscription) error {

	var wg sync.WaitGroup

	for _, subscription := range subscriptions {
		wg.Add(1)
		wiface, err := subscription.Subscribe()
		if err != nil {
			return err
		}

		go func() {

			for {
				select {
				case msg := <-wiface.ResultChan():
					subscription.Reconcile(msg.Object, msg.Type)
				case isComplete := <-subscription.IsComplete():
					if isComplete {
						wg.Done()
					}
					// TODO: Do we want a way of escaping from our go rountines?
				}
			}

		}() // Could signal handler into them??
	}
	wg.Wait()
	return nil
}

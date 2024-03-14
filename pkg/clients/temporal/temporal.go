package temporal

import (
	"go.temporal.io/sdk/client"
	"sync"
)

var once sync.Once
var oneClient client.Client

func NewClient() (client.Client, error) {
	return client.Dial(client.Options{})
}
func MustNewClient() client.Client {
	once.Do(func() {

		var err error

		if oneClient, err = NewClient(); err != nil {
			panic(err)
		}

	})
	return oneClient
}

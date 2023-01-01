package messenger

import (
	"log"
	"time"
)

func (s *Subscriber) ProcessOutgoingMessageWorker(id int) {
	for msg := range s.Outgoing {
		_, exist := s.SubscribesTo[msg.Topic]
		if !exist {
			var err = &ErrTopicNotFound{TopicName: msg.Topic}
			log.Printf("Error: %s\n", err.Error())
		}

		time.Sleep(2 * time.Second)
		log.Printf("received msg for topic %s and msg is %s\n", msg.Topic, string(msg.Data))
	}
}

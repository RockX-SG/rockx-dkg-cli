package messenger

import "log"

var numWorkers = 5

func (m *Messenger) Publish(topicName string, data []byte) error {
	tp, exist := m.Topics[topicName]
	if !exist {
		return &ErrTopicNotFound{TopicName: topicName}
	}

	m.ChMutex.Lock()
	defer m.ChMutex.Unlock()

	m.Incoming <- &Message{Topic: tp.Name, Data: data}
	return nil
}

func (m *Messenger) ProcessIncomingMessageWorker(id int) {
	for msg := range m.Incoming {
		tp, exist := m.Topics[msg.Topic]
		if !exist {
			var err = &ErrTopicNotFound{TopicName: msg.Topic}
			log.Printf("Error: %s\n", err.Error())
		}

		for _, subscriber := range tp.Subscribers {
			subscriber.ChMutex.Lock()
			subscriber.Outgoing <- msg
			subscriber.ChMutex.Unlock()
		}
	}
}

package messenger

import "fmt"

type ErrTopicNotFound struct {
	TopicName string
}

func (err *ErrTopicNotFound) Error() string {
	return fmt.Sprintf("topic with name %s not found\n", err.TopicName)
}

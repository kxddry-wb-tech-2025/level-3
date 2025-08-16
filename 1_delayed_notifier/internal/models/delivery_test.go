package models

import "testing"

type _testDeliveryImpl struct{}

func (_testDeliveryImpl) Body() []byte    { return nil }
func (_testDeliveryImpl) Ack() error      { return nil }
func (_testDeliveryImpl) Nack(bool) error { return nil }

func TestDeliveryInterfaceSatisfaction(t *testing.T) {
	var _ Delivery = _testDeliveryImpl{}
}

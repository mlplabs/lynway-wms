package wms

import "testing"

func TestService_EncodeQr(t *testing.T) {
	s := Service{}
	str := s.EncodeQr(2, 1234567)
	if str != "iwaTi" {
		t.Errorf("TestService_EncodeQr: expect `iwaTi` but got %s", str)
	}
}

func TestService_DecodeQr(t *testing.T) {
	s := Service{}
	str := "iwaTi"
	eTypeId, eId := s.DecodeQr(str)
	if eTypeId != 2 || eId != 1234567 {
		t.Errorf("TestService_DecodeQr: expect eTypeId `2` but got `%d` and expect eId `1234567` but got `%d` ", eTypeId, eId)
	}
}

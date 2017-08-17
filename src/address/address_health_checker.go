package address

type AddressHealthCheck struct {
}

func (n AddressHealthCheck) GetName() string {
	return "address"
}

func (n AddressHealthCheck) GetHealth() map[string]interface{} {
	return map[string]interface{}{
		"status": "success",
	}
}

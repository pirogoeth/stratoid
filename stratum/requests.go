package stratum

func CreateLoginRequest(username, password string, params Params) *Request {
	req := &Request{
		Method: "login",
	}

	return req.WithFields(params).WithFields(Params{
		"login": username,
		"pass":  password,
	})
}

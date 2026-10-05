package api

type DeviceCode struct {
	DeviceCode      string `json:"deviceCode"`
	UserCode        string `json:"userCode"`
	VerificationURL string `json:"verificationUrl"`
	ExpiresAt       string `json:"expiresAt"`
	Interval        int    `json:"interval"`
}

type DevicePoll struct {
	Status       string `json:"status"`
	SessionToken string `json:"sessionToken,omitempty"`
	CookieName   string `json:"cookieName,omitempty"`
}

func (c *Client) StartDeviceAuth() (DeviceCode, error) {
	return getData[DeviceCode](c, "POST", "/api/device-auth/init", nil)
}

func (c *Client) PollDeviceAuth(deviceCode string) (DevicePoll, error) {
	return getData[DevicePoll](c, "POST", "/api/device-auth/poll", map[string]string{"deviceCode": deviceCode})
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (c *Client) CurrentUser() (*User, error) {
	var out *struct {
		User *User `json:"user"`
	}
	if err := c.do("GET", "/api/auth/get-session", nil, &out); err != nil {
		return nil, err
	}
	if out == nil || out.User == nil {
		return nil, ErrUnauthorized
	}
	return out.User, nil
}

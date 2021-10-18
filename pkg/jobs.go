package pkg

type Domain struct {
	Name   string   `json:"name"`
	Target []string `json:"target"`
}
type SES struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Region    string `json:"region"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	HtmlBody  string `json:"htmlBody"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}
type Config struct {
	Name    string   `json:"name"`
	Domains []Domain `json:"domains,omitempty"`
	SES     SES      `json:"ses,omitempty"`
}

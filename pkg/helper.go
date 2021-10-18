package pkg

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	log "github.com/sirupsen/logrus"
)

var logger = log.Logger{
	Out: os.Stdout,
	Formatter: &log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
	},
	Level: log.InfoLevel,
}

func (d *Domain) getCNAMERecords() string {
	cname, err := net.LookupCNAME(d.Name)
	if err != nil {
		logger.Errorf("Error looking up CNAME for %s: %s", d.Name, err)
		return ""
	}
	return cname
}

func (d *Domain) Check() bool {
	for _, t := range d.Target {
		if strings.Contains(d.getCNAMERecords(), t) {
			logger.Println(d.Name, "targets", t)
			return true
		}
	}
	return false
}

func checkDNSTarget(domain Domain, ses SES) {
	logger.Infof("Checking %s", domain.Name)
	if !domain.Check() {
		//@todo send email with aws ses
		logger.Errorf("%s is not pointing to %s", domain.Name, domain.Target)
		sendEmail(domain, ses)
	}
}

const (
	CharSet = "UTF-8"
)

func Subject(d Domain, ses SES) string {
	return fmt.Sprintf(ses.Subject, d.Name)
}
func HtmlBody(d Domain, ses SES) string {
	return fmt.Sprintf(ses.HtmlBody, d.Name)
}
func TextBody(d Domain, ses SES) string {
	return fmt.Sprintf(ses.Body, d.Name)
}

func sendEmail(d Domain, sesc SES) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(sesc.Region)},
	)
	if err != nil {
		logger.Errorf("Error creating session: %s", err)
		return
	}

	svc := ses.New(sess, &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			sesc.AccessKey,
			sesc.SecretKey,
			"",
		),
	})

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(sesc.Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(HtmlBody(d, sesc)),
				},
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(TextBody(d, sesc)),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(Subject(d, sesc)),
			},
		},
		Source: aws.String(sesc.Sender),
	}
	result, err := svc.SendEmail(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				logger.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				logger.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				logger.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				logger.Println(aerr.Error())
			}
		} else {
			logger.Println(err.Error())
		}
		return
	}
	logger.Println("Email Sent to address: " + sesc.Recipient)
	logger.Println(result.MessageId)
}

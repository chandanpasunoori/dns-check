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
	"golang.org/x/exp/slices"
)

var logger = log.Logger{
	Out: os.Stdout,
	Formatter: &log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
	},
	Level: log.InfoLevel,
}

func (d *Domain) getCNAMERecords() []string {
	cname, err := net.LookupCNAME(d.Name)
	if err != nil {
		logger.Errorf("error looking up CNAME for %s: %s", d.Name, err)
		return []string{}
	}
	return []string{cname}
}

func (d *Domain) getARecords() []string {
	ipList, err := net.LookupIP(d.Name)
	if err != nil {
		logger.Errorf("error looking up A for %s: %s", d.Name, err)
		return []string{}
	}

	var ipListString []string

	for _, ip := range ipList {
		ipListString = append(ipListString, ip.String())
	}

	return ipListString
}

func (d *Domain) Check() (bool, []string) {

	var resolvedTarget []string
	resolvedTarget = append(resolvedTarget, d.getARecords()...)
	resolvedTarget = append(resolvedTarget, d.getCNAMERecords()...)

	for _, t := range d.Target {
		if slices.Contains(resolvedTarget, t) {
			logger.Infof("%s pointing to (%s) as expected (%s)", d.Name, strings.Join(resolvedTarget, ","), t)
			return true, resolvedTarget
		}
	}
	return false, resolvedTarget
}

func checkDNSTarget(domain Domain, ses SES) {
	logger.Infof("checking %s", domain.Name)
	if ok, resolvedList := domain.Check(); !ok {
		logger.Errorf("%s (%s) is not pointing to %s", domain.Name, strings.Join(resolvedList, ","), domain.Target)
		sendEmail(Subject(domain, ses), HtmlBody(domain, ses, resolvedList), TextBody(domain, ses, resolvedList), ses)
	}
}

func errorEmail(ses SES, err error) {
	errorMessage := fmt.Sprintf("Error in DNS Check(%s)", err.Error())
	sendEmail(errorMessage, errorMessage, errorMessage, ses)
}

const (
	CharSet = "UTF-8"
)

func Subject(d Domain, ses SES) string {
	return fmt.Sprintf(ses.Subject, d.Name)
}
func HtmlBody(d Domain, ses SES, resolvedList []string) string {
	return fmt.Sprintf(ses.HtmlBody, d.Name, strings.Join(resolvedList, ","))
}
func TextBody(d Domain, ses SES, resolvedList []string) string {
	return fmt.Sprintf(ses.Body, d.Name, strings.Join(resolvedList, ","))
}

func sendEmail(subject, htmlBody, body string, s SES) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.Region)},
	)
	if err != nil {
		logger.Errorf("error creating session: %s", err)
		errorEmail(s, err)
		return
	}

	svc := ses.New(sess, &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			s.AccessKey,
			s.SecretKey,
			"",
		),
	})

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(s.Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(htmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(body),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(s.Sender),
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
		errorEmail(s, err)
		return
	}
	logger.Println("email sent to address: " + s.Recipient)
	logger.Println(result.MessageId)
}

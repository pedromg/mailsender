package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pedromg/goEncoderBase64"
)

// emailSender is a type that includes the configs and the send function.
// This type is used both in the main code for sending with the smtp.SendMail function
// and a fakeSend mail in tests.
type emailSender struct {
	c    *Configs
	send func(string, smtp.Auth, string, []string, []byte) error
}

var (
	ErrFileNotFound           = errors.New("File %s not Found")
	ErrFileTooLarge           = errors.New("File %s is too large")
	ErrFileIsDir              = errors.New("%s is a directory")
	ErrFileParse              = errors.New("Error marshaling JSON file: %s")
	ErrSubjectTextTooLarge    = errors.New("Subject message is too large")
	ErrBodyTextTooLarge       = errors.New("Body is too large")
	ErrMessageTextTooLarge    = errors.New("Message is too large")
	ErrServerNameTextTooLarge = errors.New("Server name is too large")
	ErrAppNameTextTooLarge    = errors.New("App name is too large")
	ErrSMTPEmailAddressError  = errors.New("Invalid SMTP Email address")
	ErrEmailAddressError      = errors.New("Invalid Email address")
	ErrSend                   = errors.New("Mail send error: %s")
)

// Configs contains all the required config information
// Logger is a custom log file using the internal log package
type Configs struct {
	Configs      string `json:"configs"`
	ServerName   string `json:"server_name"`
	AppName      string `json:"app_name"`
	Secure       bool   `json:"secure"`
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPEmail    string `json:"smtp_email"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	EmailAddress string `json:"email_address"`
	Subject      string `json:"subject"`
	Body         string `json:"body"`
	Message      string `json:"message"`
	Log          bool   `json:"log"`
	LogFile      string `json:"log_file"`

	Logger *log.Logger
}

// sendIt sends the email. It can be tested using a fake function call set
// in e.send field.
func (e *emailSender) sendIt(msg string) error {
	f := mail.Address{e.c.SMTPEmail, e.c.SMTPEmail}
	t := mail.Address{e.c.EmailAddress, e.c.EmailAddress}

	auth := smtp.PlainAuth("", e.c.SMTPUsername, e.c.SMTPPassword, e.c.SMTPHost)
	return e.send(e.c.SMTPHost+":"+strconv.Itoa(e.c.SMTPPort), auth, f.Address, []string{t.Address}, []byte(msg))
}

// prepare prepares the email parts to be sent, headers, id, content-types, etc.
func (c *Configs) prepare() string {
	startTime := time.Now()
	header := make(map[string]string)
	header["From"] = c.SMTPEmail
	header["To"] = c.EmailAddress
	theMesgID := "<" + strconv.Itoa(rand.Intn(999999999)) + "__" +
		startTime.Format("2006-01-02T15:04:05.999999999Z07:00") +
		"==@" + c.SMTPHost + ">"
	header["Message-id"] = theMesgID
	header["Date"] = startTime.Format("Mon, 02 Jan 2006 15:04:05 +0000")
	header["Subject"] = c.Subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"
	msg := ""
	for k, v := range header {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n"
	msg += goEncoderBase64.Base64MimeEncoder(c.Body)

	return msg
}

// logLine builds a detailed log line to be appended to log file
func (c *Configs) logLine(s ...string) string {
	t := time.Now().Format(time.RFC3339)
	return t + " " + c.ServerName + " - " + c.AppName + " - " + strings.Join(s, " ") + "\n"
}

// logInit initializes a logger.
// func New(out io.Writer, prefix string, flag int) *Logger
func (c *Configs) logInit() error {
	fd, err := os.OpenFile(c.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	// defer fd.Close()
	c.Logger = log.New(fd, "MailSender: ", log.Ldate|log.Ltime)

	return nil
}

// jsonToConfig populates the Config struct with data from the JSON file
// The populated c has priority since it was via params
// Body and Subject will have a generic form if both (params and JSON) are
// empty
func (c *Configs) fromJSON(data []byte) error {
	tmp := Configs{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return fmt.Errorf(ErrFileParse.Error(), err)
	}

	if c.ServerName == "" {
		c.ServerName = tmp.ServerName
	}
	if c.AppName == "" {
		c.AppName = tmp.AppName
	}
	if c.SMTPHost == "" {
		c.SMTPHost = tmp.SMTPHost
	}
	if c.SMTPPort == 0 {
		c.SMTPPort = tmp.SMTPPort
	}
	if c.SMTPEmail == "" {
		c.SMTPEmail = tmp.SMTPEmail
	}
	if c.SMTPUsername == "" {
		c.SMTPUsername = tmp.SMTPUsername
	}
	if c.SMTPPassword == "" {
		c.SMTPPassword = tmp.SMTPPassword
	}
	if c.EmailAddress == "" {
		c.EmailAddress = tmp.EmailAddress
	}
	if c.Message == "" {
		c.Message = tmp.Message
	}
	if c.Subject == "" {
		c.Subject = tmp.Subject
	}
	if c.Subject == "" {
		c.Subject = fmt.Sprintf("Notification from %s - %s", c.ServerName, c.AppName)
	}
	if c.Body == "" {
		c.Body = tmp.Body
	}
	switch c.Body == "" {
	case true:
		switch c.Message == "" {
		case true:
			c.Body = fmt.Sprintf("Server: %s\nApplication: %s", c.ServerName, c.AppName)
		case false:
			c.Body = fmt.Sprintf("Server: %s\nApplication: %s\n\nMessage: %s", c.ServerName, c.AppName, c.Message)
		}
	case false:
		if c.Message != "" {
			// in this case the user may have added a %s to the body template for the message
			c.Body = fmt.Sprintf(c.Body, c.Message)
		}
	}
	if c.Body == "" {
	}
	if c.LogFile == "" {
		c.LogFile = tmp.LogFile
	}
	if c.LogFile == "" {
		c.LogFile = "./mailsender.log"
	}

	return nil
}

// load the configs from JSON file
func (c *Configs) load(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err == nil {
		err = c.fromJSON(data)
	}
	return err
}

// validate the configs fields.
// The JSON file must exist, and be smaller then 2000 bytes.
func (c *Configs) validate() error {
	// configs file
	f, err := os.Stat(c.Configs)
	if os.IsNotExist(err) {
		return ErrFileNotFound
	}
	if f.Size() > 2000 {
		return ErrFileTooLarge
	}
	if f.IsDir() {
		return fmt.Errorf(ErrFileIsDir.Error(), f.Name())
	}

	// emails
	_, err = mail.ParseAddress(c.SMTPEmail)
	if err != nil {
		return ErrSMTPEmailAddressError
	}
	_, err = mail.ParseAddress(c.EmailAddress)
	if err != nil {
		return ErrEmailAddressError
	}

	// text lengths
	if len(c.ServerName) > 254 {
		return ErrServerNameTextTooLarge
	}
	if len(c.AppName) > 254 {
		return ErrAppNameTextTooLarge
	}
	if len(c.Subject) > 254 {
		return ErrSubjectTextTooLarge
	}
	if len(c.Body) > 499 {
		return ErrBodyTextTooLarge
	}
	if len(c.Message) > 499 {
		return ErrMessageTextTooLarge
	}

	return nil
}

func main() {

	// load flags. These will have prioriry over JSON configs.
	var flagHelp bool
	var flagConfigs string
	var flagServerName string
	var flagAppName string
	var flagSecure bool
	var flagSMTPHost string
	var flagSMTPPort int
	var flagSMTPEmail string
	var flagSMTPUsername string
	var flagSMTPPassword string
	var flagEmailAddress string
	var flagSubject string
	var flagBody string
	var flagMessage string
	var flagLog bool
	var flagLogFile string

	flag.BoolVar(&flagHelp, "help", false, "help")
	flag.BoolVar(&flagHelp, "h", false, "help")
	flag.StringVar(&flagConfigs, "configs", "./mailsender.json", "configs JSON file, defaults to ./mailsender.json")
	flag.StringVar(&flagServerName, "server", "", "information about server name")
	flag.StringVar(&flagAppName, "app", "", "information about app name")
	flag.BoolVar(&flagSecure, "secure", true, "secure https")
	flag.StringVar(&flagSMTPHost, "host", "", "host")
	flag.IntVar(&flagSMTPPort, "port", 0, "port")
	flag.StringVar(&flagSMTPEmail, "from", "", "sender email address")
	flag.StringVar(&flagSMTPUsername, "user", "", "SMTP username")
	flag.StringVar(&flagSMTPPassword, "pass", "", "SMTP password")
	flag.StringVar(&flagEmailAddress, "to", "", "destination email address")
	flag.StringVar(&flagSubject, "subject", "", "email subject")
	flag.StringVar(&flagBody, "body", "", "email body template, accepts %s")
	flag.StringVar(&flagMessage, "message", "", "message passed for the body template")
	flag.BoolVar(&flagLog, "log", true, "log to file")
	flag.StringVar(&flagLogFile, "logfile", "", "/path/to/log/filename default location is ./mailsender.log")

	flag.Parse()

	if flagHelp {
		flag.Usage() //flag.PrintDefaults()
		os.Exit(0)
	}

	configs := &Configs{
		Configs:      flagConfigs,
		ServerName:   flagServerName,
		AppName:      flagAppName,
		Secure:       flagSecure,
		SMTPHost:     flagSMTPHost,
		SMTPPort:     flagSMTPPort,
		SMTPEmail:    flagSMTPEmail,
		SMTPUsername: flagSMTPUsername,
		SMTPPassword: flagSMTPPassword,
		EmailAddress: flagEmailAddress,
		Subject:      flagSubject,
		Body:         flagBody,
		Message:      flagMessage,
		Log:          flagLog,
		LogFile:      flagLogFile,
	}

	// JSON file
	err := configs.load(configs.Configs)
	if err != nil {
		flag.PrintDefaults()
		log.Fatalf("Please check mailsender -help, there was an error loading JSON data: %s", err)
	}

	if err := configs.validate(); err != nil {
		flag.PrintDefaults()
		log.Fatalf("Please check mailsender -help, there was an error: %s", err)
	}

	// logger
	if configs.Log {
		err = configs.logInit()
		if err != nil {
			flag.PrintDefaults()
			log.Fatalf("Please check mailsender -help, there was an error creating log file: %s", err)
		}
		configs.Logger.Print(configs.logLine("started"))
	}

	// send email using the custom function to the smtp.SendMail
	e := &emailSender{
		c:    configs,
		send: smtp.SendMail,
	}
	err = e.sendIt(configs.prepare())
	if err != nil {
		if configs.Log {
			configs.Logger.Print(configs.logLine("Error sending email ", err.Error()))
		}
		log.Fatalf(ErrSend.Error(), err)
	}
	// log
	if configs.Log {
		configs.Logger.Print(configs.logLine("Send OK"))
	}

}

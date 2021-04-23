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
	"time"

	"github.com/pedromg/goEncoderBase64"
)

var (
	FileNotFound           = errors.New("File %s not Found")
	FileTooLarge           = errors.New("File %s is too large")
	FileIsDir              = errors.New("%s is a directory")
	SubjectTextTooLarge    = errors.New("Subject message is too large")
	BodyTextTooLarge       = errors.New("Body is too large")
	MessageTextTooLarge    = errors.New("Message is too large")
	ServerNameTextTooLarge = errors.New("Server name is too large")
	AppNameTextTooLarge    = errors.New("App name is too large")
	SMTPEmailAddressError  = errors.New("Invalid SMTP Email address")
	EmailAddressError      = errors.New("Invalid Email address")
)

// Configs contains all the required config information
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

func (c *Configs) senderEmail(msg string) error {
	f := mail.Address{c.SMTPEmail, c.SMTPEmail}
	t := mail.Address{c.EmailAddress, c.EmailAddress}

	auth := smtp.PlainAuth("", c.SMTPUsername, c.SMTPPassword, c.SMTPHost)
	return smtp.SendMail(c.SMTPHost+":"+strconv.Itoa(c.SMTPPort), auth, f.Address, []string{t.Address}, []byte(msg))
}

func (c *Configs) send() error {
	startTime := time.Now()
	header := make(map[string]string)
	header["From"] = c.SMTPEmail
	header["To"] = c.EmailAddress
	theMesgID := "<" + strconv.Itoa(rand.Intn(999999999)) + "__" +
		startTime.Format("2006-01-02T15:04:05.999999999Z07:00") +
		"==@" + c.SMTPHost + ">"
	header["Message-id"] = theMesgID
	header["Date"] = startTime.Format("Mon, 02 Jan 2006 15:04:05 +0000")
	header["Subject"] = "mailsender Alert for " + c.Subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"
	body := "++ mailsender ALERT ++ \n\n "
	body += fmt.Sprintf(c.Body, c.Message)
	msg := ""
	for k, v := range header {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n"
	msg += goEncoderBase64.Base64MimeEncoder(body)

	return c.senderEmail(msg)
}

// logLine builds a detailed log line to be appended to log file
func (c *Configs) logLine(s string) string {
	return time.Now().String() + " " + c.ServerName + " - " + c.AppName + " - " + s + "\n"
}

// log to the configs filename
func (c *Configs) log(s string) error {
	fd, err := os.OpenFile(c.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err == nil {
		defer fd.Close()
		_, err = fd.WriteString(c.logLine(s))
	}
	return err
}

// jsonToConfig populates the Config struct with data from the JSON file
// The populated c has priority since it was via params
// Body and Subject will have a generic form if both (params and JSON) are
// empty
func (c *Configs) fromJSON(data []byte) error {
	tmp := Configs{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
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
	if c.Subject == "" {
		c.Subject = tmp.Subject
	}
	if c.Subject == "" {
		c.Subject = fmt.Sprintf("Notification from %s - %s", c.ServerName, c.AppName)
	}
	if c.Body == "" {
		c.Body = tmp.Body
	}
	if c.Body == "" {
		c.Body = fmt.Sprintf("%s\n%s\n\n%s", c.ServerName, c.AppName, c.Message)
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
		return FileNotFound
	}
	if f.Size() > 2000 {
		return FileTooLarge
	}
	if f.IsDir() {
		return fmt.Errorf(FileIsDir.Error(), f.Name())
	}

	// emails
	_, err = mail.ParseAddress(c.SMTPEmail)
	if err != nil {
		return SMTPEmailAddressError
	}
	_, err = mail.ParseAddress(c.EmailAddress)
	if err != nil {
		return EmailAddressError
	}

	// text lengths
	if len(c.ServerName) > 254 {
		return ServerNameTextTooLarge
	}
	if len(c.AppName) > 254 {
		return AppNameTextTooLarge
	}
	if len(c.Subject) > 254 {
		return SubjectTextTooLarge
	}
	if len(c.Body) > 499 {
		return BodyTextTooLarge
	}
	if len(c.Message) > 499 {
		return MessageTextTooLarge
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
	flag.StringVar(&flagBody, "", "", "email body template, accepts %s")
	flag.StringVar(&flagMessage, "", "", "message passed for the body template")
	flag.BoolVar(&flagLog, "log", true, "log to file")
	flag.StringVar(&flagLogFile, "logfile", "./mailsender.log", "/path/to/log/filename")

	flag.Parse()

	if flagHelp {
		flag.PrintDefaults()
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
		log.Fatalf("\n\nPlease check mailsender -help, there was an error loading JSON data: %s", err)
	}

	if err := configs.validate(); err != nil {
		flag.PrintDefaults()
		log.Fatalf("Please check mailsender -help, there was an error: %s", err)
	}

	// logger
	if configs.Log {
		fd, err := os.OpenFile(configs.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			flag.PrintDefaults()
			log.Fatalf("\n\nPlease check mailsender -help, there was an error creating the log file: %s", err)
		}
		defer fd.Close()
		configs.Logger = log.New(fd, "MailSender: ", log.Lshortfile)
	}

	// send email

	// log

}

package main

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	jsonFile = "./test/mailsender.json"
)

var (
	l100 = strings.Repeat(".", 100)
	l500 = strings.Repeat(l100, 5)
)

func TestConfigValidate(t *testing.T) {
	type configsCase struct {
		C    *Configs
		Pass bool
		Errs []string
	}

	cases := []configsCase{
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
				ServerName:   "LinuxBox",
				AppName:      "App",
				Subject:      "subject line",
				Body:         "body text",
			},
			true,
			nil,
		},

		configsCase{
			&Configs{
				Configs:      "./test/no_file",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender_too_big.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "bad email",
				EmailAddress: "you@example.com",
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "bad email",
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
				ServerName:   l500,
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
				AppName:      l500,
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
				Subject:      l500,
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
				Body:         l500,
			},
			false,
			nil,
		},
		configsCase{
			&Configs{
				Configs:      "./test/mailsender.json",
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
				Message:      l500,
			},
			false,
			nil,
		},
	}

	for nr, c := range cases {
		if err := c.C.validate(); (err != nil) == c.Pass {
			t.Errorf("Case %d Pass should be %t", nr, c.Pass)
		}
	}
}

func TestFromJSON(t *testing.T) {
	type Case struct {
		Data string
		Pass bool
	}

	cases := []Case{
		{
			"{ }",
			true,
		},
		{
			"{ \"\"}",
			false,
		},
		{
			"{ \"neglected_well_formed_garbage\": \"\"}",
			true,
		},
		{
			"{ \"server_name\": \"server\", \"app_name\": \"app\", \"secure\": true }",
			true,
		},
	}

	for nr, c := range cases {
		configs := &Configs{}
		if err := configs.fromJSON([]byte(c.Data)); (err != nil) == c.Pass {
			t.Errorf("Case %d Pass should be %t", nr, c.Pass)
		}

		// in case of the JSON parse passes, check for some other conditions
		if c.Pass {
			// t.Logf(">>> Subject: %s", configs.Subject)
			if configs.Subject == "" {
				t.Errorf("Case %d error: Subject should not be empty: %s", nr, configs.Subject)
			}
			// t.Logf(">>> Body: %s", configs.Body)
			if configs.Body == "" {
				t.Errorf("Case %d error: Body should not be empty: %s", nr, configs.Body)
			}
		}

	}
}

func TestLoadConfigs(t *testing.T) {
	type Case struct {
		Filename string
		Pass     bool
	}

	cases := []Case{
		{
			"",
			false,
		},
		{
			"./test",
			false,
		},
		{
			"./test/mailsender_too_big.json",
			false,
		},
		{
			"./test/mailsender.json",
			true,
		},
	}

	for nr, c := range cases {
		configs := &Configs{}
		if err := configs.load(c.Filename); (err != nil) == c.Pass {
			t.Errorf("Case %d Pass should be %t", nr, c.Pass)
		}

		// in case of the JSON parse passes, check for some other conditions
		if c.Pass {
			// t.Logf(">>> Subject: %s", configs.Subject)
			if configs.Subject == "" {
				t.Errorf("Case %d error: Subject should not be empty: %s", nr, configs.Subject)
			}
			// t.Logf(">>> Body: %s", configs.Body)
			if configs.Body == "" {
				t.Errorf("Case %d error: Body should not be empty: %s", nr, configs.Body)
			}
		}
	}
}

func TestLogLine(t *testing.T) {
	type Case struct {
		s   string
		c   *Configs
		res string
	}

	cases := []Case{
		{
			"",
			&Configs{
				ServerName: "",
				AppName:    "",
			},
			"- " + "-",
		},
		{
			"",
			&Configs{
				ServerName: "S",
				AppName:    "A",
			},
			"S" + " - " + "A" + " -",
		},
		{
			"Alpha Centauri",
			&Configs{
				ServerName: "S 1",
				AppName:    "A 1",
			},
			"S 1" + " - " + "A 1" + " - " + "Alpha Centauri",
		},
	}

	for n, c := range cases {
		// first section is a time
		strToParse := strings.Fields(c.c.logLine(c.s))[0]
		_, err := time.Parse(time.RFC3339, strToParse)
		if err != nil {
			t.Errorf("Logline time parse error: %s > %s", strToParse, err)
		}

		// last section is the buils string
		strRes := strings.Join(strings.Fields(c.res)[0:], " ")
		if strRes != c.res {
			t.Errorf("Case %d, logLine string error, should be [%s] was [%s]", n, c.res, strRes)
		}

	}

}

// TestLog logs files and evaluate size.
// It uses the system TempDir.
// For testing purposes the TempDir can be redefined.
func TestLog(t *testing.T) {

	type Case struct {
		c    *Configs
		size int64
	}

	cases := []Case{
		{
			&Configs{
				LogFile: "test.log",
			},
			48,
		},
	}

	for n, c := range cases {

		filename := fmt.Sprintf("%s%s", os.TempDir(), c.c.LogFile)
		c.c.LogFile = filename

		err := c.c.logInit()
		if err != nil {
			t.Errorf("Case %d, initializing log error %s", n, err)
		}

		f, err := os.OpenFile(filename, os.O_RDONLY, 0444) // read, read permission
		if err != nil {
			t.Errorf("Case %d, File RFONLY open error: %s", n, err)
		}

		stat, err := f.Stat()
		if err != nil {
			t.Errorf("Case %d, Stat 0 file error: %s", n, err)
		}

		// log to the file
		c.c.Logger.Printf("test case %d (b)", n)

		stat, err = f.Stat()
		if err != nil {
			t.Errorf("Case %d, Stat 1 file error: %s", n, err)
		}

		if stat.Size() != c.size {
			t.Errorf("Case %d size diff, [expected: %d] ::: [read: %d]", n, c.size, stat.Size())
		}

		// remove log for each iteration
		os.Remove(f.Name())

	}
}

func TestPrepare(t *testing.T) {

	type Case struct {
		n    int
		c    *Configs
		pass bool
	}

	cases := []Case{
		{
			1,
			&Configs{
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
				ServerName:   "LinuxBox",
				AppName:      "App",
				Subject:      "subject line",
				Body:         "body text",
			},
			true,
		},
		{
			2,
			&Configs{
				SMTPEmail:    "me@example.com",
				EmailAddress: "you@example.com",
			},
			false, // lacks subject
		},
		{
			2,
			&Configs{
				EmailAddress: "you@example.com",
				Subject:      "subject line",
			},
			false, // lacks SMTPEmail
		},
	}

	for n, c := range cases {
		m := c.c.prepare()
		// t.Log(m)
		if m == "" {
			t.Errorf("Case %d message diff", n)
		}
		switch {
		case strings.Count(m, "Message-id") != 1:
			t.Errorf("Case %d message lacks Message-id", n)
		case strings.Count(m, "Date") != 1:
			t.Errorf("Case %d message lacks Date", n)
		case strings.Count(m, "Subject") != 1:
			t.Errorf("Case %d message lacks Subject", n)
		case strings.Count(m, "Content-Type") != 1:
			t.Errorf("Case %d message lacks Content-Type", n)
		case strings.Count(m, "Content-Transfer-Encoding") != 1:
			t.Errorf("Case %d message lacks Content-Transfer-Encoding", n)
		case strings.Count(m, "From") != 1:
			t.Errorf("Case %d message lacks From", n)
		case strings.Count(m, c.c.SMTPEmail) != 1:
			if c.pass {
				t.Errorf("Case %d message lacks SMTPEmail (%s)", n, c.c.SMTPEmail)
			}
		case strings.Count(m, c.c.Subject) != 1:
			if c.pass {
				t.Errorf("Case %d message lacks Subject (%s) ", n, c.c.Subject)
			}
		}
	}
}

// fakeSend has the same signature of the smtp.SendMail()
func fakeSend(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	return nil
}

func TestSend(t *testing.T) {

	type Case struct {
		n    int
		s    *emailSender
		pass bool
	}

	cases := []Case{
		{
			1,
			&emailSender{
				c: &Configs{
					SMTPHost:     "localhost",
					SMTPPort:     25,
					SMTPEmail:    "me@example.com",
					SMTPUsername: "me",
					SMTPPassword: "pass",
					EmailAddress: "you@example.com",
					ServerName:   "LinuxBox",
					AppName:      "App",
					Subject:      "subject line",
					Body:         "body text",
				},
				send: fakeSend,
			},
			true,
		},
	}

	for n, c := range cases {
		if err := c.s.sendIt(c.s.c.prepare()); err != nil {
			t.Errorf("Case %d send error %s ", n, err)
		}
	}

}
